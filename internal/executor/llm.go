package executor

import (
	"context"
	"fmt"
	"strings"

	"github.com/aescanero/dago-libs/pkg/domain"
	"go.uber.org/zap"
)

// executeLLM executes a node in LLM mode (single completion)
func (e *Executor) executeLLM(ctx context.Context, state *domain.GraphState, config *NodeConfig) (interface{}, error) {
	llmConfig := getMapConfig(config.Config, "llm_config")
	if llmConfig == nil {
		return nil, fmt.Errorf("llm_config required for LLM mode")
	}

	// Get prompt template
	promptTemplate := getStringConfig(llmConfig, "prompt", "")
	if promptTemplate == "" {
		return nil, fmt.Errorf("prompt required in llm_config")
	}

	// Render prompt with state variables
	prompt := e.renderPrompt(promptTemplate, state, config)

	e.logger.Debug("rendered prompt",
		zap.String("node_id", config.NodeID),
		zap.Int("prompt_length", len(prompt)))

	// Build LLM request
	req := &domain.LLMRequest{
		Model:       getStringConfig(llmConfig, "model", "claude-sonnet-4-20250514"),
		System:      getStringConfig(llmConfig, "system", ""),
		Messages: []domain.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: getFloatConfig(llmConfig, "temperature", 0.7),
		MaxTokens:   getIntConfig(llmConfig, "max_tokens", 4096),
	}

	// Call LLM
	respInterface, err := e.llmClient.GenerateCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Type assert response
	resp, ok := respInterface.(*domain.LLMResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type from LLM")
	}

	e.logger.Debug("LLM response received",
		zap.String("node_id", config.NodeID),
		zap.Int("input_tokens", resp.Usage.InputTokens),
		zap.Int("output_tokens", resp.Usage.OutputTokens))

	return map[string]interface{}{
		"content": resp.Content,
		"usage":   resp.Usage,
	}, nil
}

// renderPrompt renders a prompt template with state variables
func (e *Executor) renderPrompt(template string, state *domain.GraphState, config *NodeConfig) string {
	// Simple template variable replacement
	// Format: {{variable_name}}

	prompt := template

	// Replace with state inputs
	for key, value := range state.Inputs {
		placeholder := fmt.Sprintf("{{%s}}", key)
		if str, ok := value.(string); ok {
			prompt = strings.ReplaceAll(prompt, placeholder, str)
		} else {
			// Convert non-string values to string
			prompt = strings.ReplaceAll(prompt, placeholder, fmt.Sprintf("%v", value))
		}
	}

	// Replace with node outputs from previous nodes
	for nodeID, nodeState := range state.NodeStates {
		if nodeState.Output != nil {
			placeholder := fmt.Sprintf("{{%s}}", nodeID)
			if str, ok := nodeState.Output.(string); ok {
				prompt = strings.ReplaceAll(prompt, placeholder, str)
			} else {
				prompt = strings.ReplaceAll(prompt, placeholder, fmt.Sprintf("%v", nodeState.Output))
			}
		}
	}

	return prompt
}
