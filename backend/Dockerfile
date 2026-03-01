# Stage 1: Builder
FROM golang:1.25.5-alpine AS builder

# Install build dependencies for CGO (required for SQLite)
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /build

# Copy dependency files first (for better layer caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Run critical tests before building (fail fast if tests fail)
# This ensures no broken code makes it into production
ARG SKIP_TESTS=false
RUN if [ "$SKIP_TESTS" = "false" ]; then \
    echo "=== Running test suite ===" && \
    echo "--- Testing database ---" && \
    CGO_ENABLED=1 go test -race -timeout 120s ./internal/database/... && \
    echo "--- Testing models ---" && \
    CGO_ENABLED=1 go test -race -timeout 120s ./internal/models/... && \
    echo "--- Testing services ---" && \
    CGO_ENABLED=1 go test -race -timeout 120s ./internal/services/... && \
    echo "--- Testing tools (file I/O) ---" && \
    CGO_ENABLED=1 go test -race -timeout 120s ./internal/tools/... && \
    echo "--- Testing execution (workflows) ---" && \
    CGO_ENABLED=1 go test -race -timeout 120s ./internal/execution/... && \
    echo "--- Testing filecache ---" && \
    CGO_ENABLED=1 go test -race -timeout 120s ./internal/filecache/... && \
    echo "--- Testing audio service ---" && \
    CGO_ENABLED=1 go test -race -timeout 120s ./internal/audio/... && \
    echo "--- Testing vision service ---" && \
    CGO_ENABLED=1 go test -race -timeout 120s ./internal/vision/... && \
    echo "--- Testing securefile service ---" && \
    CGO_ENABLED=1 go test -race -timeout 120s ./internal/securefile/... && \
    echo "--- Testing preflight ---" && \
    CGO_ENABLED=1 go test -race -timeout 120s ./internal/preflight/... && \
    echo "--- Testing integration ---" && \
    CGO_ENABLED=1 go test -race -timeout 120s ./tests/... && \
    echo "=== All tests passed ==="; \
    else echo "=== Skipping tests ==="; fi

# Build the application with CGO enabled (required for modernc.org/sqlite)
# -ldflags "-s -w" strips debug info to reduce binary size
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags "-s -w" -o claraverse ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies including Chromium for PDF generation
# Includes emoji fonts and extended Unicode symbol support for PDF rendering
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ttf-freefont \
    font-noto \
    font-noto-emoji \
    font-noto-cjk \
    fontconfig

# Rebuild font cache to register all installed fonts
RUN fc-cache -f -v

# Create non-root user and group
RUN addgroup -g 1000 claraverse && \
    adduser -D -u 1000 -G claraverse claraverse

# Set working directory
WORKDIR /app

# Create necessary directories with proper permissions
RUN mkdir -p /app/data /app/config /app/uploads /app/logs /app/generated /app/secure_files && \
    chown -R claraverse:claraverse /app

# Set environment variables for Chromium
ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_PATH=/usr/lib/chromium/

# Copy binary from builder stage
COPY --from=builder --chown=claraverse:claraverse /build/claraverse /app/claraverse

# Copy example configuration files (if they exist)
COPY --chown=claraverse:claraverse providers.example.json /app/providers.example.json

# Copy and set permissions for entrypoint script
COPY --chown=claraverse:claraverse docker-entrypoint.sh /app/docker-entrypoint.sh
RUN sed -i 's/\r$//' /app/docker-entrypoint.sh && chmod +x /app/docker-entrypoint.sh

# Switch to non-root user
USER claraverse

# Expose port
EXPOSE 3001

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3001/health || exit 1

# Set entrypoint and command
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/claraverse"]
