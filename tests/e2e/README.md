# End-to-End Tests

End-to-end tests for DA Node Executor.

## Prerequisites

- Docker & Docker Compose
- Go 1.21+
- LLM API key

## Running Tests

```bash
# Start Redis
docker run -d --name redis-e2e -p 6379:6379 redis:7-alpine

# Run E2E tests
export LLM_API_KEY=your-key
go test -v ./tests/e2e/...

# Stop Redis
docker stop redis-e2e && docker rm redis-e2e
```

## Test Structure

- `workflow_test.go`: Complete workflow tests
- `modes_test.go`: Test all execution modes
- `scaling_test.go`: Multi-worker scaling tests

## Writing Tests

E2E tests should:
- Test complete workflows
- Use real worker instances
- Validate full execution flow
- Test scaling behavior
