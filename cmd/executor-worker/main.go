package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aescanero/dago-adapters/pkg/llm"
	"github.com/aescanero/dago-node-executor/internal/config"
	"github.com/aescanero/dago-node-executor/internal/executor"
	"github.com/aescanero/dago-node-executor/internal/worker"
	"github.com/aescanero/dago-node-executor/pkg/tools/function"
	"github.com/aescanero/dago-node-executor/pkg/tools/mcp"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Version is set by build flags
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := initLogger(cfg.LogLevel)
	defer func() { _ = logger.Sync() }()

	logger.Info("starting executor worker",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
		zap.String("worker_id", cfg.WorkerID))

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("failed to connect to Redis", zap.Error(err))
	}
	logger.Info("connected to Redis", zap.String("addr", cfg.RedisAddr))

	// Initialize LLM client using dago-adapters
	llmClient, err := llm.NewClient(&llm.Config{
		Provider: cfg.LLMProvider,
		APIKey:   cfg.LLMAPIKey,
		BaseURL:  cfg.LLMBaseURL,
		Logger:   logger,
	})
	if err != nil {
		logger.Fatal("failed to create LLM client", zap.Error(err))
	}

	// Initialize tool clients
	mcpClient := mcp.NewClient(cfg.GetMCPServers(), logger)
	functionRegistry := function.NewRegistry(logger)

	// Register built-in tools
	registerBuiltInTools(functionRegistry)

	// Create composite tool client (try MCP first, then function registry)
	toolClient := newCompositeToolClient(mcpClient, functionRegistry, logger)

	// Initialize executor
	exec := executor.NewExecutor(llmClient, toolClient, logger, cfg.MaxIterations)

	// Create worker
	w := worker.NewWorker(&worker.Config{
		ID:          cfg.WorkerID,
		RedisClient: redisClient,
		Executor:    exec,
		Logger:      logger,
	})

	// Start health server
	healthServer := worker.NewHealthServer(w, cfg.HealthPort, logger)
	if err := healthServer.Start(); err != nil {
		logger.Fatal("failed to start health server", zap.Error(err))
	}

	// Start worker
	if err := w.Start(); err != nil {
		logger.Fatal("failed to start worker", zap.Error(err))
	}

	logger.Info("executor worker started",
		zap.String("worker_id", cfg.WorkerID),
		zap.Int("health_port", cfg.HealthPort))

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("received shutdown signal")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop components
	if err := healthServer.Stop(shutdownCtx); err != nil {
		logger.Error("health server shutdown error", zap.Error(err))
	}

	if err := w.Stop(shutdownCtx); err != nil {
		logger.Error("worker shutdown error", zap.Error(err))
	}

	if err := redisClient.Close(); err != nil {
		logger.Error("Redis close error", zap.Error(err))
	}

	logger.Info("executor worker shut down complete")
}

// initLogger initializes the logger
func initLogger(level string) *zap.Logger {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	return logger
}

// compositeToolClient tries MCP first, then falls back to function registry
type compositeToolClient struct {
	mcp      *mcp.Client
	registry *function.Registry
	logger   *zap.Logger
}

func newCompositeToolClient(mcpClient *mcp.Client, registry *function.Registry, logger *zap.Logger) *compositeToolClient {
	return &compositeToolClient{
		mcp:      mcpClient,
		registry: registry,
		logger:   logger,
	}
}

func (c *compositeToolClient) Execute(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	// Try MCP first
	result, err := c.mcp.Execute(ctx, toolName, params)
	if err == nil {
		return result, nil
	}

	c.logger.Debug("MCP execution failed, trying function registry",
		zap.String("tool", toolName),
		zap.Error(err))

	// Fallback to function registry
	return c.registry.Execute(ctx, toolName, params)
}

func (c *compositeToolClient) ListTools(ctx context.Context) ([]string, error) {
	// Get tools from both sources
	mcpTools, _ := c.mcp.ListTools(ctx)
	registryTools, _ := c.registry.ListTools(ctx)

	// Combine and deduplicate
	toolSet := make(map[string]bool)
	for _, tool := range mcpTools {
		toolSet[tool] = true
	}
	for _, tool := range registryTools {
		toolSet[tool] = true
	}

	tools := make([]string, 0, len(toolSet))
	for tool := range toolSet {
		tools = append(tools, tool)
	}

	return tools, nil
}

// registerBuiltInTools registers built-in tools
func registerBuiltInTools(registry *function.Registry) {
	// MVP: No built-in tools yet
	// TODO: Add common built-in tools (e.g., string manipulation, math)
}
