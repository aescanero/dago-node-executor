# Execution Modes

The DA Node Executor supports three execution modes, automatically detected based on node configuration.

## Mode Detection

Mode is determined by the presence of specific configuration fields:

| Mode  | llm_config | tools | tool_name |
|-------|-----------|-------|-----------|
| Agent | ✓         | ✓     | -         |
| LLM   | ✓         | -     | -         |
| Tool  | -         | -     | ✓         |

## Agent Mode

### Overview

Agent mode implements a reasoning-action loop where the LLM can:
1. Reason about the task
2. Decide which tools to use
3. Execute tools
4. Observe results
5. Repeat until done

### Configuration

```json
{
  "type": "agent",
  "llm_config": {
    "model": "claude-sonnet-4-20250514",
    "temperature": 0.7,
    "max_tokens": 4096
  },
  "tools": ["search", "calculate", "read_file"],
  "max_iterations": 10,
  "system": "You are a helpful research assistant."
}
```

### Execution Flow

```
1. Construct prompt with:
   - System message
   - Available tools (from MCP discovery)
   - Task description
   - Previous conversation history

2. Call LLM with tools enabled
   ↓
3. Parse response:
   - If tool calls → Execute via MCP
   - If no tool calls → Done
   ↓
4. Add tool results to conversation
   ↓
5. Repeat from step 2 (up to max_iterations)
```

### Example

**Input:**
```json
{
  "task": "Find the latest AI research on reasoning",
  "tools": ["web_search", "summarize"]
}
```

**Agent Loop:**

*Iteration 1:*
```
Thought: I need to search for AI research on reasoning
Action: web_search("latest AI research on reasoning 2025")
```

*Iteration 2:*
```
Observation: Found 5 research papers
Thought: I should summarize the key findings
Action: summarize(papers)
```

*Iteration 3:*
```
Observation: Summary completed
Thought: I have the information needed
Final Answer: [Summary of latest AI reasoning research]
```

### Best Practices

- Set reasonable max_iterations (default: 10)
- Provide clear system instructions
- Define specific tools needed
- Handle tool errors gracefully

### When to Use

- Complex tasks requiring multiple steps
- Tasks needing external information
- Tasks requiring tool orchestration
- Open-ended problem solving

## LLM Mode

### Overview

LLM mode performs a single completion without tool execution. Useful for:
- Text generation
- Analysis
- Classification
- Transformation

### Configuration

```json
{
  "type": "llm",
  "llm_config": {
    "model": "claude-sonnet-4-20250514",
    "temperature": 0.3,
    "system": "You are a data analyst.",
    "prompt": "Analyze this data: {{input}}"
  }
}
```

### Execution Flow

```
1. Render prompt with state variables
   ↓
2. Call LLM (single request)
   ↓
3. Parse response
   ↓
4. Return result
```

### Prompt Templating

Use `{{variable}}` syntax for state variables:

```
Analyze the following data:
{{data}}

Previous analysis:
{{previous_results}}

Provide insights in JSON format.
```

### Structured Output

Request structured output in the prompt:

```json
{
  "llm_config": {
    "prompt": "Analyze sentiment. Return JSON: {\"sentiment\": \"positive|negative|neutral\", \"score\": 0-1}"
  }
}
```

### Best Practices

- Use lower temperature for deterministic tasks
- Specify output format clearly
- Keep prompts focused
- Use system message for role definition

### When to Use

- Simple text tasks
- No tool execution needed
- Deterministic output required
- Single-step processing

## Tool Mode

### Overview

Tool mode directly executes a specified tool without LLM involvement. Useful for:
- Data retrieval
- API calls
- File operations
- Calculations

### Configuration

```json
{
  "type": "tool",
  "tool_name": "web_search",
  "tool_params": {
    "query": "{{search_query}}",
    "max_results": 10
  }
}
```

### Execution Flow

```
1. Extract tool name and parameters
   ↓
2. Resolve parameter templates
   ↓
3. Execute tool via MCP
   ↓
4. Return result
```

### Optional LLM Parameter Extraction

Use LLM to extract parameters from natural language:

```json
{
  "type": "tool",
  "tool_name": "search",
  "extract_params": {
    "from": "{{user_query}}",
    "llm_config": {
      "prompt": "Extract search parameters from: {{user_query}}"
    }
  }
}
```

### Best Practices

- Validate parameters
- Handle tool errors
- Use for deterministic operations
- Avoid when reasoning needed

### When to Use

- Direct API/tool call needed
- Parameters known upfront
- No reasoning required
- Fast execution needed

## Mode Comparison

| Feature           | Agent | LLM | Tool |
|-------------------|-------|-----|------|
| LLM Calls         | Many  | One | Optional |
| Tool Execution    | Yes   | No  | Yes  |
| Reasoning Loop    | Yes   | No  | No   |
| Complexity        | High  | Low | Low  |
| Execution Time    | Slow  | Fast| Fast |
| Token Usage       | High  | Low | Low  |
| Flexibility       | High  | Med | Low  |

## Choosing the Right Mode

### Use Agent Mode When:
- Task requires multiple steps
- Need to use multiple tools
- Task is open-ended
- Reasoning is required

### Use LLM Mode When:
- Single text operation
- Analysis or generation
- No external tools needed
- Fast execution preferred

### Use Tool Mode When:
- Direct tool call
- Parameters known
- No reasoning needed
- Fastest execution required

## Error Handling

### All Modes

- Validate configuration
- Retry on transient errors
- Log execution details
- Return structured errors

### Agent Mode Specific

- Limit iterations to prevent loops
- Handle tool execution failures
- Track conversation history
- Detect stuck loops

### LLM Mode Specific

- Handle rate limits
- Validate response format
- Parse structured output
- Handle empty responses

### Tool Mode Specific

- Validate tool availability
- Handle parameter errors
- Check tool result format
- Handle tool timeouts

## Examples

See `examples/` directory for:
- `agent_example.json` - Agent mode configuration
- `llm_example.json` - LLM mode configuration
- `tool_example.json` - Tool mode configuration
