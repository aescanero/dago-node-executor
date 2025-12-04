// Package executor implements the core execution logic for nodes.
//
// Supports three execution modes:
//   - Agent: Reasoning-action loop with tool execution
//   - LLM: Single LLM completion
//   - Tool: Direct tool execution
//
// Mode is automatically detected based on node configuration.
package executor
