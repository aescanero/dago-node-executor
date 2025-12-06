package executor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aescanero/dago-libs/pkg/domain"
	"go.uber.org/zap"
)

// executeAgent executes a node in agent mode (reasoning-action loop)
func (e *Executor) executeAgent(ctx context.Context, state *domain.GraphState, config *NodeConfig) (interface{}, error) {
	llmConfig := getMapConfig(config.Config, "llm_config")
	if llmConfig == nil {
		return nil, fmt.Errorf("llm_config required for agent mode")
	}

	tools := getSliceConfig(config.Config, "tools")
	if len(tools) == 0 {
		return nil, fmt.Errorf("tools required for agent mode")
	}

	maxIterations := getIntConfig(config.Config, "max_iterations", e.maxIterations)
	system := getStringConfig(llmConfig, "system", "You are a helpful AI assistant with access to tools.")

	// Build conversation history
	messages := []domain.Message{}

	// Add initial user message with task
	task := getStringConfig(config.Config, "task", "Complete the assigned task.")
	messages = append(messages, domain.Message{
		Role:    "user",
		Content: task,
	})

	// Agent loop
	for iteration := 0; iteration < maxIterations; iteration++ {
		e.logger.Debug("agent iteration",
			zap.String("node_id", config.NodeID),
			zap.Int("iteration", iteration))

		// Construct LLM request with tools
		req := &domain.LLMRequest{
			Model:       getStringConfig(llmConfig, "model", "claude-sonnet-4-20250514"),
			System:      system,
			Messages:    messages,
			Temperature: getFloatConfig(llmConfig, "temperature", 0.7),
			MaxTokens:   getIntConfig(llmConfig, "max_tokens", 4096),
		}

		// Call LLM
		respInterface, err := e.llmClient.GenerateCompletion(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("LLM call failed at iteration %d: %w", iteration, err)
		}

		// Type assert response
		resp, ok := respInterface.(*domain.LLMResponse)
		if !ok {
			return nil, fmt.Errorf("unexpected response type from LLM at iteration %d", iteration)
		}

		// Add assistant response to conversation
		messages = append(messages, domain.Message{
			Role:    "assistant",
			Content: resp.Content,
		})

		// Check if agent is done (no tool calls)
		toolCalls := resp.ToolCalls
		if len(toolCalls) == 0 {
			e.logger.Info("agent completed",
				zap.String("node_id", config.NodeID),
				zap.Int("iterations", iteration+1))
			return map[string]interface{}{
				"result":     resp.Content,
				"iterations": iteration + 1,
			}, nil
		}

		// Execute tools
		e.logger.Debug("executing tools",
			zap.String("node_id", config.NodeID),
			zap.Int("tool_count", len(toolCalls)))

		for _, toolCall := range toolCalls {
			result, err := e.toolClient.Execute(ctx, toolCall.Name, toolCall.Input)
			if err != nil {
				e.logger.Error("tool execution failed",
					zap.String("tool", toolCall.Name),
					zap.Error(err))
				result = map[string]interface{}{
					"error": err.Error(),
				}
			}

			// Add tool result to conversation
			resultJSON, _ := json.Marshal(result)
			messages = append(messages, domain.Message{
				Role:    "user",
				Content: fmt.Sprintf("Tool result for %s: %s", toolCall.Name, string(resultJSON)),
			})
		}
	}

	return nil, fmt.Errorf("max iterations (%d) reached without completion", maxIterations)
}
