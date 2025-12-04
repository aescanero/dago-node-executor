# Integration Tests

Integration tests for DA Node Executor.

## Prerequisites

- Go 1.21+
- Redis running on localhost:6379
- LLM API key (for full tests)

## Running Tests

```bash
# Run all integration tests
go test -v ./tests/integration/...

# Run with Redis
docker run -d --name redis-test -p 6379:6379 redis:7-alpine
go test -v ./tests/integration/...
docker stop redis-test && docker rm redis-test
```

## Test Structure

- `executor_test.go`: Executor integration tests
- `worker_test.go`: Worker integration tests
- `tools_test.go`: Tool client integration tests

## Writing Tests

Integration tests should:
- Test real components (Redis, etc.)
- Clean up resources after tests
- Use test-specific prefixes
- Be idempotent
