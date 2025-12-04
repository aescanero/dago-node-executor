package api

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Client implements API-based tool execution (deprecated in favor of MCP)
type Client struct {
	baseURL string
	logger  *zap.Logger
}

// NewClient creates a new API client
func NewClient(baseURL string, logger *zap.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		logger:  logger,
	}
}

// Execute executes a tool via REST API
func (c *Client) Execute(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	c.logger.Warn("API client is deprecated, use MCP instead",
		zap.String("tool", toolName))

	return nil, fmt.Errorf("API client deprecated, use MCP")
}

// ListTools lists available tools
func (c *Client) ListTools(ctx context.Context) ([]string, error) {
	return []string{}, fmt.Errorf("API client deprecated, use MCP")
}
