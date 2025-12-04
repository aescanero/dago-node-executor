module github.com/aescanero/dago-node-executor

go 1.25.5

require (
	github.com/aescanero/dago-libs v0.2.0

	// LLM client
	github.com/anthropics/anthropic-sdk-go v1.17.0

	// Configuration
	github.com/caarlos0/env/v10 v10.0.0
	github.com/google/uuid v1.5.0

	// Redis Streams for events
	github.com/redis/go-redis/v9 v9.3.0

	// Logging
	go.uber.org/zap v1.26.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)
