// Package executor provides node execution logic for different modes (agent, llm, tool).
package executor

import (
	"context"
	"fmt"

	"github.com/aescanero/dago-libs/pkg/domain"
	"github.com/aescanero/dago-libs/pkg/ports"
	"go.uber.org/zap"
)

// Executor executes nodes in different modes
type Executor struct {
	llmClient     ports.LLMClient
	toolClient    ToolClient
	logger        *zap.Logger
	maxIterations int
}

// NodeConfig represents the configuration for a node execution
type NodeConfig struct {
	NodeID string
	Config map[string]interface{}
}

// ToolClient defines the interface for tool execution
type ToolClient interface {
	Execute(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error)
	ListTools(ctx context.Context) ([]string, error)
}

// NewExecutor creates a new executor
func NewExecutor(llmClient ports.LLMClient, toolClient ToolClient, logger *zap.Logger, maxIterations int) *Executor {
	if maxIterations == 0 {
		maxIterations = 10 // default
	}

	return &Executor{
		llmClient:     llmClient,
		toolClient:    toolClient,
		logger:        logger,
		maxIterations: maxIterations,
	}
}

// Execute executes a node based on its configuration
func (e *Executor) Execute(ctx context.Context, state *domain.GraphState, config *NodeConfig) (interface{}, error) {
	mode := DetectMode(config)

	e.logger.Info("executing node",
		zap.String("node_id", config.NodeID),
		zap.String("mode", string(mode)))

	switch mode {
	case ModeAgent:
		return e.executeAgent(ctx, state, config)
	case ModeLLM:
		return e.executeLLM(ctx, state, config)
	case ModeTool:
		return e.executeTool(ctx, state, config)
	default:
		return nil, fmt.Errorf("unknown execution mode: %s", mode)
	}
}

// Helper functions for config extraction

func getStringConfig(config map[string]interface{}, key string, defaultValue string) string {
	if v, ok := config[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultValue
}

func getIntConfig(config map[string]interface{}, key string, defaultValue int) int {
	if v, ok := config[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		}
	}
	return defaultValue
}

func getFloat64Config(config map[string]interface{}, key string, defaultValue float64) float64 {
	if v, ok := config[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return defaultValue
}

// Alias for backwards compatibility
func getFloatConfig(config map[string]interface{}, key string, defaultValue float64) float64 {
	return getFloat64Config(config, key, defaultValue)
}

func getMapConfig(config map[string]interface{}, key string) map[string]interface{} {
	if v, ok := config[key]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

func getSliceConfig(config map[string]interface{}, key string) []interface{} {
	if v, ok := config[key]; ok {
		if s, ok := v.([]interface{}); ok {
			return s
		}
	}
	return nil
}
