FROM alpine:3.19

# Install GCC and G++
RUN apk add --no-cache \
    gcc \
    g++ \
    musl-dev \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN adduser -D -u 1000 compiler

# Copy compilation script
COPY compile.sh /usr/bin/compile.sh
RUN chmod +x /usr/bin/compile.sh

# Set up workspace
RUN mkdir -p /workspace && chown compiler:compiler /workspace

USER compiler
WORKDIR /workspace

# Default command
CMD ["/usr/bin/compile.sh"]
