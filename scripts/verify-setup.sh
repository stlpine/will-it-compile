#!/bin/bash
# verify-setup.sh - Verifies that the development environment is properly set up

set -e

echo "🔍 Verifying will-it-compile setup..."
echo

# Check if Docker is running
echo "✓ Checking Docker daemon..."
if ! docker info &> /dev/null; then
    echo "❌ ERROR: Docker daemon is not running"
    echo "   Please start Docker and try again"
    exit 1
fi
echo "  ✓ Docker daemon is running"
echo

# Check for required Docker images
echo "✓ Checking Docker images..."
MISSING_IMAGES=()

# List of required official images (Docker will pull automatically if missing)
REQUIRED_IMAGES=(
    "gcc:13"
    "golang:1.22-alpine"
    "rust:1.75-alpine"
)

for image in "${REQUIRED_IMAGES[@]}"; do
    if docker image inspect "$image" &> /dev/null; then
        echo "  ✓ Found: $image"
    else
        echo "  ℹ️  Not cached: $image (will be pulled on first use)"
        MISSING_IMAGES+=("$image")
    fi
done

echo

# If any images are missing, suggest pulling them
if [ ${#MISSING_IMAGES[@]} -gt 0 ]; then
    echo "ℹ️  Some Docker images are not cached locally."
    echo "   They will be pulled automatically on first use."
    echo
    echo "To pre-pull images for faster first compilation, run:"
    echo "  make docker-pull"
    echo
fi

# Check if Go is installed
echo "✓ Checking Go installation..."
if ! command -v go &> /dev/null; then
    echo "  ⚠️  WARNING: Go is not installed"
    echo "     Required for building the project"
else
    GO_VERSION=$(go version | awk '{print $3}')
    echo "  ✓ Go version: $GO_VERSION"
fi
echo

# Check Go modules
if [ -f "go.mod" ]; then
    echo "✓ Checking Go dependencies..."
    if go mod verify &> /dev/null; then
        echo "  ✓ All Go dependencies verified"
    else
        echo "  ⚠️  WARNING: Some Go dependencies may be outdated"
        echo "     Run: go mod download"
    fi
    echo
fi

# Try to build the project
echo "✓ Attempting to build project..."
if go build -o /dev/null ./cmd/api &> /dev/null; then
    echo "  ✓ Project builds successfully"
else
    echo "  ❌ ERROR: Project failed to build"
    echo "     Run: go build ./cmd/api for details"
    exit 1
fi
echo

# Success!
echo "✅ All checks passed!"
echo
echo "You can now run the server with:"
echo "  make run"
echo "  or: ./bin/will-it-compile-api"
echo
