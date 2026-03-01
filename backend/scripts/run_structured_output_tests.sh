#!/bin/bash

# Structured Output Testing Suite
# Tests all configured LLM providers for JSON schema support

set -e

echo "🧪 LLM Provider Structured Output Testing Suite"
echo "=============================================="
echo ""

# Load environment variables (filter out comments and empty lines)
if [ -f ../.env ]; then
    set -a
    source <(cat ../.env | grep -v '^#' | grep -v '^$' | sed 's/#.*$//' | sed 's/[[:space:]]*$//')
    set +a
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go to run this test suite."
    exit 1
fi

# Run the test suite
echo "📦 Compiling test suite..."
go run test_structured_outputs.go

echo ""
echo "✅ Testing complete! Check the report above for results."
echo ""
echo "📄 Results saved to: structured_output_test_results.json"
