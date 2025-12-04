package function

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// ToolFunc represents a tool function
type ToolFunc func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// Registry is a registry for built-in tool functions
type Registry struct {
	tools  map[string]ToolFunc
	mu     sync.RWMutex
	logger *zap.Logger
}

// NewRegistry creates a new tool registry
func NewRegistry(logger *zap.Logger) *Registry {
	return &Registry{
		tools:  make(map[string]ToolFunc),
		logger: logger,
	}
}

// Register registers a tool function
func (r *Registry) Register(name string, fn ToolFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools[name] = fn
	r.logger.Info("tool registered", zap.String("tool", name))
}

// Execute executes a registered tool
func (r *Registry) Execute(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	r.mu.RLock()
	fn, ok := r.tools[toolName]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	r.logger.Debug("executing registered tool",
		zap.String("tool", toolName),
		zap.Any("params", params))

	return fn(ctx, params)
}

// ListTools lists all registered tools
func (r *Registry) ListTools(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]string, 0, len(r.tools))
	for name := range r.tools {
		tools = append(tools, name)
	}

	return tools, nil
}

// Has checks if a tool is registered
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.tools[name]
	return ok
}
