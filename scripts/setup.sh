#!/bin/bash
# Setup script for will-it-compile

set -e

echo "Will-It-Compile Setup"
echo "====================="
echo ""

# Check prerequisites
echo "Checking prerequisites..."

# Check Go
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.25 or later."
    exit 1
fi
echo "✓ Go $(go version | awk '{print $3}')"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker."
    exit 1
fi
echo "✓ Docker $(docker --version | awk '{print $3}' | tr -d ',')"

# Check Docker daemon
if ! docker info &> /dev/null; then
    echo "❌ Docker daemon is not running. Please start Docker."
    exit 1
fi
echo "✓ Docker daemon is running"

echo ""
echo "Installing dependencies..."
go mod download

echo ""
echo "Pulling official compiler images..."
docker pull gcc:13
docker pull golang:1.22-alpine
docker pull rust:1.75-alpine

echo ""
echo "Building API server..."
go build -o bin/will-it-compile-api cmd/api/main.go

echo ""
echo "✓ Setup complete!"
echo ""
echo "To start the server, run:"
echo "  ./bin/will-it-compile-api"
echo ""
echo "Or use make:"
echo "  make run"
