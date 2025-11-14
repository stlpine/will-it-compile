#!/bin/bash
set -e

IMAGE_NAME="will-it-compile/cpp-gcc"
IMAGE_TAG="13-alpine"

echo "Building Docker image: ${IMAGE_NAME}:${IMAGE_TAG}"
docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" .

echo "Image built successfully!"
echo "To test: docker run --rm ${IMAGE_NAME}:${IMAGE_TAG}"
