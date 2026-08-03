#!/bin/sh
set -e

# =============================================================================
# ClaraVerse Docker Entrypoint
# =============================================================================

echo "================================================"
echo "  ClaraVerse - Starting up..."
echo "================================================"

mkdir -p /app/data /app/uploads /app/logs

# Auto-generate secrets
if [ -z "$JWT_SECRET" ]; then
    KEY_FILE="/app/data/.jwt_secret"
    if [ ! -f "$KEY_FILE" ]; then
        echo "[init] Generating JWT secret key..."
        head -c 32 /dev/urandom | base64 | tr -d '\n' > "$KEY_FILE"
    else
        echo "[init] Loading existing JWT secret"
    fi
    export JWT_SECRET=$(cat "$KEY_FILE")
fi

if [ -z "$ENCRYPTION_MASTER_KEY" ]; then
    KEY_FILE="/app/data/.encryption_key"
    if [ ! -f "$KEY_FILE" ]; then
        echo "[init] Generating encryption master key..."
        head -c 32 /dev/urandom | od -A n -t x1 | tr -d ' \n' > "$KEY_FILE"
    else
        echo "[init] Loading existing encryption key"
    fi
    export ENCRYPTION_MASTER_KEY=$(cat "$KEY_FILE")
fi

# Generate runtime config.js for frontend
if [ "$SERVE_FRONTEND" = "true" ] && [ -d "/app/public" ]; then
    cat > /app/public/config.js << JSEOF
window.__CLARA_CONFIG__ = {
  API_BASE_URL: "${CLARA_API_BASE_URL:-}",
  WS_URL: "${CLARA_WS_URL:-}",
  APP_NAME: "${CLARA_APP_NAME:-ClaraVerse}",
};
JSEOF
    echo "[init] Frontend runtime config generated"
fi

# Copy SearXNG settings if needed
if [ -f "/app/config/searxng-settings.yml" ] && [ ! -f "/app/data/searxng-settings.yml" ]; then
    cp /app/config/searxng-settings.yml /app/data/searxng-settings.yml
fi

# Detect local AI providers.
#
# --tries=1 matters as much as the timeout here. wget retries 20 times by
# default, and --timeout bounds each attempt rather than the whole command, so
# on a host whose firewall DROPs these ports (rather than refusing them) the
# probe blocks for ~40s per provider and stalls container startup. One attempt
# with a short timeout is all a presence check needs.
probe_local_provider() {
    wget -q --spider --timeout=2 --tries=1 "$1" 2>/dev/null
}

OLLAMA_STATUS="not detected"
LMSTUDIO_STATUS="not detected"

OLLAMA_URL="${OLLAMA_BASE_URL:-http://host.docker.internal:11434}"
if probe_local_provider "$OLLAMA_URL"; then
    OLLAMA_STATUS="detected at $OLLAMA_URL"
    echo "[init] Found Ollama at $OLLAMA_URL — models will be auto-imported"
fi

LMSTUDIO_URL="${LMSTUDIO_BASE_URL:-http://host.docker.internal:1234}"
if probe_local_provider "$LMSTUDIO_URL/v1/models"; then
    LMSTUDIO_STATUS="detected at $LMSTUDIO_URL"
    echo "[init] Found LM Studio at $LMSTUDIO_URL — models will be auto-imported"
fi

echo ""
echo "  Port:        ${PORT:-3000}"
echo "  Database:    ${DATABASE_URL:+connected}"
echo "  MongoDB:     ${MONGODB_URI:+connected}"
echo "  Redis:       ${REDIS_URL:+connected}"
echo "  SearXNG:     ${SEARXNG_URL:+connected}"
echo "  Ollama:      $OLLAMA_STATUS"
echo "  LM Studio:   $LMSTUDIO_STATUS"
echo ""
echo "================================================"
echo ""

exec "$@"
