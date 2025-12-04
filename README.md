# DA Node Executor

Executor worker for the DA Orchestrator - executes graph nodes in agent, LLM, or tool modes.

## Overview

The DA Node Executor is a horizontally scalable worker that:

- **Subscribes to Redis Streams** for work distribution
- **Executes nodes** in three modes: agent, llm, tool
- **Integrates with MCP** for tool execution (primary)
- **Scales horizontally** - run multiple instances for high throughput

## Architecture

```
┌─────────────────────────────────────────────────┐
│           Redis Streams (executor.work)         │
└─────────────────┬───────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────┐
│              Executor Worker                     │
│  ┌──────────────────────────────────────────┐  │
│  │  Mode Detection & Execution              │  │
│  │  - Agent Mode (reasoning-action loop)    │  │
│  │  - LLM Mode (single completion)          │  │
│  │  - Tool Mode (direct execution)          │  │
│  └──────────────────────────────────────────┘  │
│                                                  │
│  ┌──────────┐  ┌─────────┐  ┌──────────────┐  │
│  │   LLM    │  │   MCP   │  │   Storage    │  │
│  │  Client  │  │  Client │  │   (Redis)    │  │
│  └──────────┘  └─────────┘  └──────────────┘  │
└─────────────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────┐
│       Redis Streams (node.completed/failed)     │
└─────────────────────────────────────────────────┘
```

## Quick Start

### Using Docker

```bash
docker run -d \
  --name executor-worker \
  -e REDIS_ADDR=redis:6379 \
  -e LLM_API_KEY=your-api-key \
  aescanero/dago-node-executor:latest
```

### Using Docker Compose

```bash
# In main dago repository
docker-compose up -d executor-worker
```

### Building from Source

```bash
# Install dependencies
make deps

# Build binary
make build

# Run locally
export REDIS_ADDR=localhost:6379
export LLM_API_KEY=your-api-key
./executor-worker
```

## Configuration

Configuration via environment variables:

| Variable          | Default            | Description                    |
|-------------------|--------------------|--------------------------------|
| `WORKER_ID`       | `executor-1`       | Worker identifier              |
| `REDIS_ADDR`      | `localhost:6379`   | Redis server address           |
| `REDIS_PASS`      | (empty)            | Redis password                 |
| `LLM_PROVIDER`    | `anthropic`        | LLM provider                   |
| `LLM_API_KEY`     | (required)         | LLM API key                    |
| `LLM_MODEL`       | `claude-sonnet-4-20250514` | Default LLM model    |
| `MCP_SERVERS`     | (empty)            | Comma-separated MCP servers    |
| `MAX_ITERATIONS`  | `10`               | Max agent loop iterations      |
| `LOG_LEVEL`       | `info`             | Log level (debug,info,warn,error)|

## Execution Modes

### Agent Mode

Reasoning-action loop with tool execution:

```json
{
  "type": "agent",
  "llm_config": {
    "model": "claude-sonnet-4-20250514",
    "temperature": 0.7
  },
  "tools": ["search", "calculate"],
  "max_iterations": 10
}
```

### LLM Mode

Single LLM completion:

```json
{
  "type": "llm",
  "llm_config": {
    "model": "claude-sonnet-4-20250514",
    "prompt": "Analyze the following data..."
  }
}
```

### Tool Mode

Direct tool execution:

```json
{
  "type": "tool",
  "tool_name": "search",
  "tool_params": {
    "query": "latest AI research"
  }
}
```

See [docs/MODES.md](docs/MODES.md) for detailed mode documentation.

## Scaling

Run multiple executor workers for horizontal scaling:

```bash
# Run 3 workers
docker-compose up -d --scale executor-worker=3
```

Each worker:
- Has a unique ID
- Subscribes to the same Redis Stream
- Processes work independently
- Can be stopped/started without affecting others

## Development

### Prerequisites

- Go 1.25.5+
- Redis 7.0+
- Docker (optional)

### Running Tests

```bash
# Unit tests
make test

# Integration tests (requires Redis)
go test ./tests/integration/...
```

### Project Structure

```
dago-node-executor/
├── cmd/executor-worker/    # Main entry point
├── internal/
│   ├── executor/           # Execution logic (agent, llm, tool)
│   ├── worker/             # Worker lifecycle
│   └── config/             # Configuration
├── pkg/tools/              # Tool adapters (MCP, function, API)
├── deployments/docker/     # Docker files
└── docs/                   # Documentation
```

## Documentation

- [Execution Modes](docs/MODES.md) - Detailed mode documentation
- [Worker Documentation](docs/README.md) - Architecture and internals
- [Changelog](docs/CHANGELOG.md) - Version history

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Links

- **Domain**: [disasterproject.com](https://disasterproject.com)
- **GitHub**: [github.com/aescanero/dago-node-executor](https://github.com/aescanero/dago-node-executor)
- **Docker Hub**: [aescanero/dago-node-executor](https://hub.docker.com/r/aescanero/dago-node-executor)
- **Dependencies**: [dago-libs](https://github.com/aescanero/dago-libs)
