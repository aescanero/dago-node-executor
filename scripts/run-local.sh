#!/bin/bash

# Run executor worker locally for development

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting DA Node Executor (Local Development)${NC}"
echo ""

# Check if Redis is running
if ! redis-cli ping > /dev/null 2>&1; then
    echo -e "${YELLOW}Redis is not running. Starting Redis in Docker...${NC}"
    docker run -d --name redis-executor -p 6379:6379 redis:7-alpine
    sleep 2
fi

# Check environment variables
if [ -z "${LLM_API_KEY}" ]; then
    echo -e "${RED}Error: LLM_API_KEY environment variable is required${NC}"
    exit 1
fi

# Set defaults
export WORKER_ID=${WORKER_ID:-"executor-local"}
export REDIS_ADDR=${REDIS_ADDR:-"localhost:6379"}
export LOG_LEVEL=${LOG_LEVEL:-"debug"}

echo "Configuration:"
echo "  Worker ID: ${WORKER_ID}"
echo "  Redis: ${REDIS_ADDR}"
echo "  Log Level: ${LOG_LEVEL}"
echo ""

# Build if needed
if [ ! -f "./executor-worker" ]; then
    echo -e "${YELLOW}Building executor worker...${NC}"
    make build
fi

# Run
echo -e "${GREEN}Starting worker...${NC}"
./executor-worker
