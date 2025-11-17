# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Copy source code and Makefile
COPY . .

# Build the application using Makefile (handles deps automatically)
RUN make build-api

# Runtime stage
FROM alpine:3.19

# Install Docker client for container orchestration
RUN apk add --no-cache docker-cli ca-certificates

# Create non-root user
RUN adduser -D -u 1001 apiuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/will-it-compile-api .

# Copy configuration files
COPY --from=builder /app/configs ./configs

# Change ownership
RUN chown -R apiuser:apiuser /app

# Switch to non-root user
USER apiuser

# Expose API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./will-it-compile-api"]
