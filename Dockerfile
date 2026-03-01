# =============================================================================
# ClaraVerse Docker Image
# =============================================================================
# Multi-stage build: React frontend + Go backend in a single lightweight image.
# Databases (MySQL, MongoDB, Redis) run as separate containers via compose.
#
# Usage:
#   docker build -t claraverse .
#   docker run -d -p 3000:3000 -v claraverse-data:/app/data claraverse
# =============================================================================

# ---- Stage 1: Build Frontend ------------------------------------------------
FROM node:20-alpine AS frontend-build

WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --legacy-peer-deps
COPY frontend/ .

ENV VITE_API_BASE_URL="" \
    VITE_WS_URL="" \
    VITE_APP_NAME=ClaraVerse \
    VITE_APP_VERSION=1.0.0 \
    VITE_ENABLE_ANALYTICS=false

RUN npx vite build

# ---- Stage 2: Build Backend -------------------------------------------------
FROM golang:1.25.5-alpine AS backend-build

RUN apk add --no-cache gcc musl-dev

WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags "-s -w" -o claraverse ./cmd/server

# ---- Stage 3: Runtime Image -------------------------------------------------
FROM alpine:3.20

LABEL org.opencontainers.image.title="ClaraVerse" \
      org.opencontainers.image.description="Open-source AI platform — chat, agents, workflows, tools" \
      org.opencontainers.image.source="https://github.com/claraverse-space/ClaraVerse" \
      org.opencontainers.image.licenses="AGPL-3.0"

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget \
    bash \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ttf-freefont \
    font-noto \
    font-noto-emoji \
    font-noto-cjk \
    fontconfig && \
    fc-cache -f -v

WORKDIR /app

RUN mkdir -p /app/data /app/uploads /app/logs /app/public /app/config && \
    addgroup -g 1000 claraverse && \
    adduser -D -u 1000 -G claraverse claraverse && \
    chown -R claraverse:claraverse /app

ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_PATH=/usr/lib/chromium/

COPY --from=backend-build --chown=claraverse:claraverse /build/claraverse /app/claraverse
COPY --from=frontend-build --chown=claraverse:claraverse /frontend/dist /app/public
COPY --chown=claraverse:claraverse backend/migrations/ /app/migrations/
COPY --chown=claraverse:claraverse searxng/settings.yml /app/config/searxng-settings.yml
COPY --chown=claraverse:claraverse docker-entrypoint-unified.sh /app/docker-entrypoint.sh
RUN sed -i 's/\r$//' /app/docker-entrypoint.sh && chmod +x /app/docker-entrypoint.sh

ENV PORT=3000 \
    ENVIRONMENT=production \
    SERVE_FRONTEND=true \
    UPLOAD_DIR=/app/uploads \
    PROVIDERS_FILE=/app/data/providers.json \
    ALLOWED_ORIGINS=* \
    DOCKER=true

USER claraverse

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/claraverse"]
