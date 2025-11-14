#!/bin/bash
#
# Demo script for the will-it-compile TUI
#
# This script helps you test the TUI by:
# 1. Checking if the API server is running
# 2. Starting it if needed
# 3. Launching the TUI

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== will-it-compile TUI Demo ===${NC}\n"

# Check if API server is running
echo -e "${YELLOW}Checking API server...${NC}"
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ API server is running${NC}\n"
else
    echo -e "${YELLOW}⚠ API server not running. Starting it...${NC}"

    # Check if binary exists
    if [ ! -f "./bin/will-it-compile-api" ]; then
        echo -e "${YELLOW}Building API server...${NC}"
        make build-api
    fi

    # Start server in background
    ./bin/will-it-compile-api &
    API_PID=$!

    # Wait for server to be ready
    echo -e "${YELLOW}Waiting for server to start...${NC}"
    for i in {1..10}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo -e "${GREEN}✓ API server started (PID: $API_PID)${NC}\n"
            break
        fi
        sleep 1
    done
fi

# Check if TUI binary exists
if [ ! -f "./bin/will-it-compile-tui" ]; then
    echo -e "${YELLOW}Building TUI...${NC}"
    make build-tui
fi

# Launch TUI
echo -e "${GREEN}=== Launching TUI ===${NC}\n"
echo -e "${YELLOW}Quick Tips:${NC}"
echo "  • Press '?' for help"
echo "  • Press 'f' to load a file (try: tests/samples/hello.cpp)"
echo "  • Press 'l' to change language"
echo "  • Press Enter to compile"
echo "  • Press Tab to view history"
echo "  • Press 'q' or Ctrl+C to quit"
echo ""
sleep 2

./bin/will-it-compile-tui

# Cleanup message
echo -e "\n${GREEN}TUI closed.${NC}"
echo -e "${YELLOW}Note: API server is still running. Stop it with: pkill will-it-compile-api${NC}"
