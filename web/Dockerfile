# Build stage
FROM node:24-alpine AS builder

# Enable pnpm via corepack
RUN corepack enable

# Set working directory
WORKDIR /app

# Copy dependency files
COPY package.json pnpm-lock.yaml ./

# Install dependencies with optimization flags
RUN pnpm install --frozen-lockfile --prod=false \
    && pnpm store prune

# Copy source code
COPY . .

# Build the application with production optimizations
RUN pnpm run build

# Production stage - use smallest nginx image
FROM nginx:1.27-alpine-slim

# Install only wget for healthcheck (lighter than default packages)
RUN apk add --no-cache wget

# Remove default nginx configs and unnecessary files
RUN rm -rf /etc/nginx/conf.d/default.conf \
    && rm -rf /usr/share/nginx/html/* \
    && rm -rf /var/cache/apk/*

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built files from builder (static assets only)
COPY --from=builder /app/dist /usr/share/nginx/html

# Use non-root user for security
RUN chown -R nginx:nginx /usr/share/nginx/html \
    && chown -R nginx:nginx /var/cache/nginx \
    && chown -R nginx:nginx /var/log/nginx \
    && touch /var/run/nginx.pid \
    && chown -R nginx:nginx /var/run/nginx.pid

USER nginx

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/ || exit 1

# Expose port
EXPOSE 80

# Run nginx
CMD ["nginx", "-g", "daemon off;"]
