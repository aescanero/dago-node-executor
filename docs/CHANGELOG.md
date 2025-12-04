# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Advanced agent strategies
- Custom tool integrations
- Performance optimizations
- Enhanced error recovery

## [1.0.0] - 2025-12-02

### Added
- Initial MVP release
- Three execution modes: agent, llm, tool
- Redis Streams work subscription
- MCP tool integration (primary)
- Function registry (fallback)
- Anthropic Claude LLM support
- Horizontal scaling support
- Docker container deployment
- GitHub Actions CI/CD
- Comprehensive documentation

### Execution Modes
- **Agent Mode**: Reasoning-action loop with tool execution
- **LLM Mode**: Single LLM completion
- **Tool Mode**: Direct tool execution

### Infrastructure
- Redis Streams for work distribution
- Redis for state storage
- Environment-based configuration
- Health check endpoint
- Structured logging

### Dependencies
- dago-libs v1.0.0
- Redis Go client v9
- Anthropic SDK
- Zap logger

### Notes
- MVP focus on core functionality
- MCP as primary tool integration
- Simple retry logic
- Basic error handling

## [0.1.0] - 2025-11-15

### Added
- Project initialization
- Basic project structure

[Unreleased]: https://github.com/aescanero/dago-node-executor/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/aescanero/dago-node-executor/releases/tag/v1.0.0
[0.1.0]: https://github.com/aescanero/dago-node-executor/releases/tag/v0.1.0
