package executor

import (
	"context"
	"fmt"

	"github.com/aescanero/dago-libs/pkg/domain"
	"go.uber.org/zap"
)

// executeTool executes a node in tool mode (direct tool execution)
func (e *Executor) executeTool(ctx context.Context, state *domain.GraphState, config *NodeConfig) (interface{}, error) {
	toolName := getStringConfig(config.Config, "tool_name", "")
	if toolName == "" {
		return nil, fmt.Errorf("tool_name required for tool mode")
	}

	// Get tool parameters
	toolParams := getMapConfig(config.Config, "tool_params")
	if toolParams == nil {
		toolParams = make(map[string]interface{})
	}

	// Resolve parameter templates
	resolvedParams := e.resolveParams(toolParams, state)

	e.logger.Debug("executing tool",
		zap.String("node_id", config.NodeID),
		zap.String("tool", toolName),
		zap.Any("params", resolvedParams))

	// Execute tool
	result, err := e.toolClient.Execute(ctx, toolName, resolvedParams)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	e.logger.Debug("tool execution completed",
		zap.String("node_id", config.NodeID),
		zap.String("tool", toolName))

	return result, nil
}

// resolveParams resolves parameter templates with state values
func (e *Executor) resolveParams(params map[string]interface{}, state *domain.GraphState) map[string]interface{} {
	resolved := make(map[string]interface{})

	for key, value := range params {
		switch v := value.(type) {
		case string:
			// Check if it's a template {{variable}}
			if len(v) > 4 && v[:2] == "{{" && v[len(v)-2:] == "}}" {
				varName := v[2 : len(v)-2]

				// Try to get from inputs
				if inputVal, ok := state.Inputs[varName]; ok {
					resolved[key] = inputVal
					continue
				}

				// Try to get from node outputs
				if nodeState, ok := state.NodeStates[varName]; ok && nodeState.Output != nil {
					resolved[key] = nodeState.Output
					continue
				}

				// Keep original if not found
				resolved[key] = v
			} else {
				resolved[key] = v
			}
		case map[string]interface{}:
			// Recursively resolve nested maps
			resolved[key] = e.resolveParams(v, state)
		default:
			resolved[key] = v
		}
	}

	return resolved
}
