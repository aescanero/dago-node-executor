# DA Node Executor - Complete Project Structure

## Repository Statistics

- **Total Files Created**: 28
- **Go Source Files**: 17
- **YAML Configuration Files**: 3
- **Documentation Files**: 6
- **Shell Scripts**: 2

## Complete Directory Structure

```
dago-node-executor/
├── .github/
│   └── workflows/              # GitHub Actions CI/CD
│       ├── ci.yml             # Tests + lint
│       ├── release.yml        # Build binary + create release
│       └── docker.yml         # Build Docker image, push with DOCKER_TOKEN
│
├── docs/                      # Documentation
│   ├── README.md             # Worker architecture and internals
│   ├── MODES.md              # Execution modes documentation
│   └── CHANGELOG.md          # Version history
│
├── cmd/
│   └── executor-worker/
│       └── main.go            # Main entry point (317 lines)
│
├── internal/                  # Private code
│   ├── config/
│   │   ├── config.go         # Configuration from env (115 lines)
│   │   └── doc.go
│   │
│   ├── executor/             # Execution logic
│   │   ├── executor.go       # Main executor (168 lines)
│   │   ├── mode.go           # Mode detection (38 lines)
│   │   ├── agent.go          # Agent mode (113 lines)
│   │   ├── llm.go            # LLM mode (81 lines)
│   │   ├── tool.go           # Tool mode (74 lines)
│   │   └── doc.go
│   │
│   └── worker/               # Worker implementation
│       ├── worker.go         # Worker lifecycle (283 lines)
│       ├── health.go         # Health checks (89 lines)
│       └── doc.go
│
├── pkg/                       # Public (importable)
│   └── tools/                 # Tool adapters
│       ├── mcp/
│       │   ├── client.go     # MCP client (primary)
│       │   └── doc.go
│       ├── function/
│       │   ├── registry.go   # Function registry (fallback)
│       │   └── doc.go
│       └── api/
│           ├── client.go     # API client (deprecated)
│           └── doc.go
│
├── deployments/
│   └── docker/
│       ├── Dockerfile         # Multi-stage build
│       └── .dockerignore
│
├── scripts/
│   ├── build.sh               # Build script
│   └── run-local.sh           # Local development script
│
├── tests/
│   ├── integration/
│   │   └── README.md
│   └── e2e/
│       └── README.md
│
├── .gitignore
├── .dockerignore
├── go.mod                     # Depends on dago-libs v1.0.0
├── go.sum
├── Makefile
├── README.md
├── LICENSE
└── PROJECT_STRUCTURE.md       # This file
```

## Key Components

### 1. Executor Logic (`internal/executor/`)

#### Mode Detection
- Automatically detects execution mode from config
- Agent: Has llm_config + tools
- LLM: Has llm_config only
- Tool: Has tool_name only

#### Agent Mode
- Reasoning-action loop
- Tool selection and execution
- Max iterations limit (default: 10)
- Conversation history tracking

#### LLM Mode
- Single LLM completion
- Prompt templating with {{variables}}
- Structured output parsing

#### Tool Mode
- Direct tool execution
- Parameter template resolution
- Fast execution path

### 2. Worker Implementation (`internal/worker/`)

#### Worker Lifecycle
- Connects to Redis Streams
- Subscribes to "executor.work" stream
- Processes work items
- Publishes results to "node.completed/failed"
- Graceful shutdown

#### Health Checks
- HTTP health endpoint on port 8081
- Redis connection check
- Readiness probe

### 3. Tool Integration (`pkg/tools/`)

#### MCP Client (Primary)
- Model Context Protocol support
- Tool discovery
- Session management
- MVP: Placeholder pending SDK

#### Function Registry (Fallback)
- Simple function registration
- Built-in tools
- Thread-safe execution

#### API Client (Deprecated)
- Legacy REST API support
- Marked for removal

### 4. Main Entry Point (`cmd/executor-worker/main.go`)

Complete worker initialization:
- Configuration loading
- Redis connection
- LLM client setup (Anthropic)
- Tool client initialization
- Composite tool client (MCP + function registry)
- Worker creation and startup
- Health server
- Graceful shutdown

### 5. Configuration (`internal/config/`)

Environment variables:
- `WORKER_ID`: Unique worker identifier
- `REDIS_ADDR`, `REDIS_PASS`, `REDIS_DB`: Redis connection
- `LLM_PROVIDER`, `LLM_API_KEY`, `LLM_MODEL`: LLM configuration
- `MCP_SERVERS`: Comma-separated MCP servers
- `MAX_ITERATIONS`: Agent loop limit
- `LOG_LEVEL`: Logging level
- `HEALTH_PORT`: Health check port

### 6. Deployment

#### Dockerfile
- Multi-stage build (Go builder + Alpine runtime)
- Non-root user
- Health checks
- Minimal image size

#### GitHub Actions
- **CI**: Tests with Redis, linting, build verification
- **Release**: Multi-platform binaries (Linux, macOS, Windows - amd64/arm64)
- **Docker**: Multi-platform images, push to aescanero/dago-node-executor

### 7. Scripts

#### build.sh
- Single platform build
- Multi-platform build with `./build.sh all`
- Version and build time injection

#### run-local.sh
- Local development runner
- Automatic Redis startup if needed
- Environment validation

## Data Flow

### Work Processing Flow

```
1. Redis Streams (executor.work)
   ↓
2. Worker reads message
   ↓
3. Load graph state from Redis
   ↓
4. Detect execution mode
   ↓
5. Execute node:
   - Agent: Reasoning-action loop
   - LLM: Single completion
   - Tool: Direct execution
   ↓
6. Update graph state
   ↓
7. Save state to Redis
   ↓
8. Publish result (node.completed/failed)
   ↓
9. Acknowledge message
```

### Agent Mode Flow

```
1. Construct prompt with tools
   ↓
2. Call LLM
   ↓
3. Parse response for tool calls
   ↓
4. If tool calls:
   - Execute via MCP
   - Add results to conversation
   - Repeat (max iterations)
   ↓
5. If no tool calls:
   - Return final answer
```

## Scaling

### Horizontal Scaling

Run multiple workers:
```bash
# Worker 1
WORKER_ID=executor-1 docker run aescanero/dago-node-executor

# Worker 2
WORKER_ID=executor-2 docker run aescanero/dago-node-executor

# Worker 3
WORKER_ID=executor-3 docker run aescanero/dago-node-executor
```

Redis Streams automatically distributes work across workers in consumer group.

### Load Distribution

- Round-robin work distribution
- Pending message tracking
- Automatic redelivery on failure
- Each message processed by exactly one worker

## Dependencies

### External Dependencies
- **dago-libs v1.0.0**: Domain models and port interfaces
- **Redis Go Client v9**: Redis Streams integration
- **Anthropic SDK**: LLM client
- **Zap**: Structured logging

### Infrastructure Requirements
- **Redis 7.0+**: Work distribution and state storage
- **Anthropic API**: LLM provider

## MVP Focus

### Implemented
✅ Three execution modes (agent, llm, tool)
✅ Redis Streams work subscription
✅ Worker lifecycle management
✅ Health checks
✅ Docker deployment
✅ GitHub Actions CI/CD
✅ Horizontal scaling support
✅ Comprehensive documentation

### Simplified for MVP
- MCP client (placeholder pending SDK)
- Tool call parsing (basic implementation)
- Error recovery (simple retry logic)
- Built-in tools (none yet)

### Future Enhancements
- Full MCP protocol implementation
- Advanced agent strategies
- Custom tool integrations
- Performance optimizations
- Enhanced error recovery
- Metrics and monitoring
- Tool call streaming

## Development

### Running Locally

```bash
# Start Redis
docker run -d -p 6379:6379 redis:7-alpine

# Set environment
export REDIS_ADDR=localhost:6379
export LLM_API_KEY=your-key

# Build and run
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

### Building

```bash
# Single platform
make build

# All platforms
./scripts/build.sh all

# Docker image
make docker-build
```

## Architecture Principles

1. **Clean Architecture**: Clear separation of concerns
2. **Mode-based Execution**: Automatic mode detection
3. **Horizontal Scalability**: Stateless workers
4. **Simple MVP**: Focus on core functionality
5. **MCP First**: Primary tool integration method
6. **Dependency Injection**: Testable components
7. **Graceful Shutdown**: No work loss
8. **Health Monitoring**: Observable workers

## Documentation

- **README.md**: Quick start and overview
- **docs/README.md**: Architecture and internals
- **docs/MODES.md**: Execution modes explained
- **docs/CHANGELOG.md**: Version history
- **PROJECT_STRUCTURE.md**: This file

## License

MIT License - see LICENSE file for details.

## Links

- **Domain**: https://disasterproject.com
- **GitHub**: https://github.com/aescanero/dago-node-executor
- **Docker Hub**: https://hub.docker.com/r/aescanero/dago-node-executor
- **Dependencies**: https://github.com/aescanero/dago-libs
