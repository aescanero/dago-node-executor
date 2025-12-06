package executor

// ExecutionMode represents the execution mode
type ExecutionMode string

const (
	// ModeAgent: Reasoning-action loop with tools
	ModeAgent ExecutionMode = "agent"

	// ModeLLM: Single LLM completion
	ModeLLM ExecutionMode = "llm"

	// ModeTool: Direct tool execution
	ModeTool ExecutionMode = "tool"
)

// DetectMode detects the execution mode based on configuration
func DetectMode(config *NodeConfig) ExecutionMode {
	// Check for tool mode (has tool_name)
	if toolName := getStringConfig(config.Config, "tool_name", ""); toolName != "" {
		return ModeTool
	}

	// Check for llm_config
	llmConfig := getMapConfig(config.Config, "llm_config")
	if llmConfig == nil {
		// No LLM config, default to tool mode if no explicit mode
		return ModeTool
	}

	// Check for tools (indicates agent mode)
	tools := getSliceConfig(config.Config, "tools")
	if len(tools) > 0 {
		return ModeAgent
	}

	// Has LLM config but no tools - LLM mode
	return ModeLLM
}
