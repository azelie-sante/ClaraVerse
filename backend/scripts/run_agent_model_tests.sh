#!/bin/bash

# Agent Model Testing Script
# Tests all models with "agents": true and measures their performance

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "🧪 Running Agent Model Performance Tests"
echo "========================================"
echo ""
echo "This script will:"
echo "  1. Test all models with 'agents: true'"
echo "  2. Measure average response time (3 tests per model)"
echo "  3. Provide suggestions for providers.json updates"
echo ""
echo "⚠️  This may take several minutes depending on the number of models"
echo ""

read -p "Press Enter to continue or Ctrl+C to cancel..."

# Check if providers.json exists
if [ ! -f "$BACKEND_DIR/providers.json" ]; then
    echo "❌ Error: providers.json not found in $BACKEND_DIR"
    exit 1
fi

# Build and run the test script
echo ""
echo "📦 Building test script..."
cd "$SCRIPT_DIR"
go build -o test_agent_models test_agent_models.go

echo "🚀 Running tests..."
echo ""

# Run from backend directory so it can find providers.json
cd "$BACKEND_DIR"
"$SCRIPT_DIR/test_agent_models"

# Cleanup
rm -f "$SCRIPT_DIR/test_agent_models"

echo ""
echo "✅ Testing complete!"
echo ""
echo "📝 Next steps:"
echo "  1. Review the test results above"
echo "  2. Copy the suggested updates to providers.json"
echo "  3. Restart the backend server"
echo ""
