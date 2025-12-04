# DA Node Executor - Worker Documentation

## Architecture

The DA Node Executor is a horizontally scalable worker that executes graph nodes based on their configuration.

### Component Overview

```
┌────────────────────────────────────────────────┐
│            Worker Lifecycle                     │
│  ┌──────────────────────────────────────────┐ │
│  │  1. Connect to Redis Streams             │ │
│  │  2. Subscribe to "executor.work"         │ │
│  │  3. Wait for work events                 │ │
│  │  4. Process node execution               │ │
│  │  5. Publish results                      │ │
│  └──────────────────────────────────────────┘ │
└────────────────────────────────────────────────┘

┌────────────────────────────────────────────────┐
│           Executor Logic                        │
│  ┌──────────────────────────────────────────┐ │
│  │  Mode Detection                          │ │
│  │    ├─ Agent: llm_config + tools         │ │
│  │    ├─ LLM: llm_config only              │ │
│  │    └─ Tool: tool_name only              │ │
│  └──────────────────────────────────────────┘ │
│  ┌──────────────────────────────────────────┐ │
│  │  Execution                               │ │
│  │    ├─ Agent Mode: Reasoning loop        │ │
│  │    ├─ LLM Mode: Single completion       │ │
│  │    └─ Tool Mode: Direct execution       │ │
│  └──────────────────────────────────────────┘ │
└────────────────────────────────────────────────┘

┌────────────────────────────────────────────────┐
│            Tool Integration                     │
│  ┌──────────────────────────────────────────┐ │
│  │  MCP Client (Primary)                    │ │
│  │    - Tool discovery                      │ │
│  │    - Session management                  │ │
│  │    - Execution                           │ │
│  └──────────────────────────────────────────┘ │
│  ┌──────────────────────────────────────────┐ │
│  │  Function Registry (Fallback)            │ │
│  │    - Built-in tools                      │ │
│  │    - Simple registration                 │ │
│  └──────────────────────────────────────────┘ │
└────────────────────────────────────────────────┘
```

## Worker Flow

### 1. Initialization

```go
// Load configuration
cfg := config.Load()

// Connect to Redis
redisClient := redis.NewClient(&redis.Options{
    Addr: cfg.RedisAddr,
})

// Initialize LLM client
llmClient := llm.NewClient(cfg.LLMProvider, cfg.LLMAPIKey)

// Initialize tool clients
mcpClient := mcp.NewClient(cfg.MCPServers)
toolRegistry := function.NewRegistry()

// Create executor
executor := executor.NewExecutor(llmClient, mcpClient, toolRegistry)

// Create and start worker
worker := worker.NewWorker(cfg.WorkerID, redisClient, executor)
worker.Start()
```

### 2. Work Processing

```go
// Worker loop
for {
    // Read from Redis Stream
    event := redisClient.XReadGroup("executor.work")

    // Parse work configuration
    nodeConfig := parseNodeConfig(event.Data)

    // Load state from Redis
    state := loadState(nodeConfig.GraphID)

    // Execute node
    result, err := executor.Execute(ctx, state, nodeConfig)

    // Save updated state
    saveState(state)

    // Publish result
    if err != nil {
        publishEvent("node.failed", result)
    } else {
        publishEvent("node.completed", result)
    }
}
```

### 3. Execution Modes

#### Agent Mode

Reasoning-action loop with tool execution:

1. Construct prompt with available tools
2. Call LLM for reasoning
3. Parse tool calls from response
4. Execute tools via MCP
5. Add tool results to conversation
6. Repeat until done or max iterations

#### LLM Mode

Single LLM completion:

1. Render prompt with state
2. Call LLM
3. Parse structured output
4. Return result

#### Tool Mode

Direct tool execution:

1. Extract tool name and parameters
2. Execute via MCP or function registry
3. Return result

## Error Handling

### Retry Logic

```go
maxRetries := 3
for attempt := 0; attempt < maxRetries; attempt++ {
    result, err := executor.Execute(ctx, state, nodeConfig)
    if err == nil {
        return result, nil
    }

    if !isRetriable(err) {
        return nil, err
    }

    time.Sleep(backoff(attempt))
}
```

### Error Categories

- **Retriable**: Network errors, rate limits, temporary LLM errors
- **Non-retriable**: Configuration errors, validation errors, permanent failures

## Scaling

### Horizontal Scaling

Run multiple workers:

```bash
# Worker 1
WORKER_ID=executor-1 ./executor-worker

# Worker 2
WORKER_ID=executor-2 ./executor-worker

# Worker 3
WORKER_ID=executor-3 ./executor-worker
```

Each worker:
- Has a unique ID
- Subscribes to the same Redis Stream consumer group
- Processes work independently
- Redis ensures each work item is processed by only one worker

### Load Distribution

Redis Streams automatically distributes work across workers in the consumer group:
- Round-robin distribution
- Pending message tracking
- Automatic redelivery on failure

## Configuration

### Environment Variables

- `WORKER_ID`: Unique worker identifier
- `REDIS_ADDR`: Redis server address
- `REDIS_PASS`: Redis password
- `LLM_PROVIDER`: LLM provider (anthropic)
- `LLM_API_KEY`: LLM API key
- `LLM_MODEL`: Default LLM model
- `MCP_SERVERS`: Comma-separated MCP server URLs
- `MAX_ITERATIONS`: Max agent loop iterations
- `LOG_LEVEL`: Log level

### MCP Configuration

MCP servers can be configured via:

```bash
export MCP_SERVERS="http://tools-server:8080,http://search-server:8081"
```

Each server should support the MCP protocol for tool discovery and execution.

## Monitoring

### Health Checks

The worker exposes a health endpoint:

```bash
curl http://localhost:8081/health
```

Response:
```json
{
  "status": "healthy",
  "worker_id": "executor-1",
  "redis": "connected",
  "last_processed": "2025-12-02T10:30:00Z"
}
```

### Metrics

Key metrics to monitor:
- Work items processed
- Execution time per mode
- Error rate
- Tool execution count
- LLM API latency

## Development

### Running Locally

```bash
# Start Redis
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Set environment
export REDIS_ADDR=localhost:6379
export LLM_API_KEY=your-key

# Run worker
make run-local
```

### Testing

```bash
# Unit tests
make test

# Integration tests (requires Redis)
go test ./tests/integration/...

# E2E tests
go test ./tests/e2e/...
```

## Troubleshooting

### Common Issues

**Worker not processing work**
- Check Redis connection
- Verify consumer group exists
- Check for pending messages

**Tool execution failures**
- Verify MCP server connectivity
- Check tool configuration
- Review tool logs

**LLM API errors**
- Verify API key
- Check rate limits
- Review error messages
