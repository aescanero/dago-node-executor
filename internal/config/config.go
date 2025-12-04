package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v10"
)

// Config holds all configuration for the executor worker
type Config struct {
	// Worker
	WorkerID string `env:"WORKER_ID" envDefault:"executor-1"`

	// Redis
	RedisAddr string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisPass string `env:"REDIS_PASS"`
	RedisDB   int    `env:"REDIS_DB" envDefault:"0"`

	// LLM
	LLMProvider string `env:"LLM_PROVIDER" envDefault:"anthropic"`
	LLMAPIKey   string `env:"LLM_API_KEY"`
	LLMModel    string `env:"LLM_MODEL" envDefault:"claude-sonnet-4-20250514"`

	// MCP
	MCPServers []string `env:"MCP_SERVERS" envSeparator:","`

	// Agent
	MaxIterations int `env:"MAX_ITERATIONS" envDefault:"10"`

	// Logging
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	// Health
	HealthPort int `env:"HEALTH_PORT" envDefault:"8081"`
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.WorkerID == "" {
		return fmt.Errorf("worker ID is required")
	}

	if c.RedisAddr == "" {
		return fmt.Errorf("Redis address is required")
	}

	if c.LLMAPIKey == "" {
		return fmt.Errorf("LLM API key is required")
	}

	if c.LLMProvider != "anthropic" {
		return fmt.Errorf("unsupported LLM provider: %s (only 'anthropic' supported in MVP)", c.LLMProvider)
	}

	if c.MaxIterations < 1 {
		return fmt.Errorf("max iterations must be at least 1")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}

	return nil
}

// GetMCPServers returns the list of MCP servers
func (c *Config) GetMCPServers() []string {
	var servers []string
	for _, server := range c.MCPServers {
		server = strings.TrimSpace(server)
		if server != "" {
			servers = append(servers, server)
		}
	}
	return servers
}
