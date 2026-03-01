#!/bin/sh
set -e

# ClaraVerse Backend Docker Entrypoint Script
# This script handles initialization before starting the application

echo "=== ClaraVerse Backend Starting ==="

# Check if .env file exists
if [ ! -f "/app/.env" ]; then
    echo "WARNING: .env file not found. Using example configuration."
    echo "Please create .env file from .env.example before running in production."

    # Create .env from example if available
    if [ -f "/app/.env.example" ]; then
        cp /app/.env.example /app/.env
        echo "Created .env from .env.example"
    fi
fi

# Check if providers.json exists at configured location
PROVIDERS_PATH="${PROVIDERS_FILE:-/app/providers.json}"
if [ ! -f "$PROVIDERS_PATH" ]; then
    echo "WARNING: providers.json not found at $PROVIDERS_PATH. Using example configuration."

    # Create providers.json from example if available
    if [ -f "/app/providers.example.json" ]; then
        cp /app/providers.example.json "$PROVIDERS_PATH"
        echo "Created providers.json from providers.example.json"
    else
        echo "ERROR: No providers configuration found!"
        echo "Please mount providers.json or create one from providers.example.json"
        exit 1
    fi
fi

# Ensure data directories exist with proper permissions
mkdir -p /app/data /app/uploads /app/logs

# Display configuration summary
echo ""
echo "Configuration:"
echo "  - Database: ${DATABASE_PATH:-model_capabilities.db}"
echo "  - Providers: ${PROVIDERS_FILE:-providers.json}"
echo "  - Upload Dir: ${UPLOAD_DIR:-./uploads}"
echo "  - Environment: ${ENVIRONMENT:-development}"
echo "  - Port: ${PORT:-3001}"
echo ""

# Check for required environment variables in production
if [ "$ENVIRONMENT" = "production" ]; then
    echo "Running in PRODUCTION mode"

    if [ -z "$SUPABASE_URL" ] || [ -z "$SUPABASE_KEY" ]; then
        echo "WARNING: Supabase configuration is missing!"
        echo "SUPABASE_URL and SUPABASE_KEY are required in production mode."
        echo "The server will terminate if authentication is not properly configured."
    fi
fi

echo ""
echo "Starting ClaraVerse backend..."
echo "==================================="
echo ""

# Execute the main application
exec "$@"
