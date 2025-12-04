// Package worker implements the worker lifecycle and Redis Stream integration.
//
// The worker:
//   - Subscribes to Redis Streams for work distribution
//   - Processes work items using the executor
//   - Publishes results back to Redis Streams
//   - Provides health check endpoints
package worker
