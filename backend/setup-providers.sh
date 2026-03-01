#!/bin/bash

# ClaraVerse Provider Setup Script
# This script helps you quickly set up providers and filters

echo "🚀 ClaraVerse Provider Setup"
echo "=============================="
echo ""

BASE_URL="http://localhost:3001"

# Function to add a provider
add_provider() {
    local name=$1
    local base_url=$2
    local api_key=$3

    echo "📡 Adding provider: $name"

    response=$(curl -s -X POST "$BASE_URL/api/providers" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"$name\",
            \"base_url\": \"$base_url\",
            \"api_key\": \"$api_key\"
        }")

    provider_id=$(echo $response | grep -o '"id":[0-9]*' | grep -o '[0-9]*')

    if [ -n "$provider_id" ]; then
        echo "   ✅ Provider added with ID: $provider_id"
        echo "$provider_id"
    else
        echo "   ❌ Failed to add provider"
        echo "   Response: $response"
        echo "0"
    fi
}

# Function to add a filter
add_filter() {
    local provider_id=$1
    local pattern=$2
    local action=$3
    local priority=$4

    echo "   🔧 Adding filter: $pattern ($action)"

    curl -s -X POST "$BASE_URL/api/providers/$provider_id/filters" \
        -H "Content-Type: application/json" \
        -d "{
            \"pattern\": \"$pattern\",
            \"action\": \"$action\",
            \"priority\": $priority
        }" > /dev/null
}

# Function to refresh models
refresh_models() {
    local provider_id=$1

    echo "   🔄 Refreshing models for provider $provider_id..."

    curl -s -X POST "$BASE_URL/api/models/refresh/$provider_id" > /dev/null

    echo "   ✅ Models refreshed"
}

echo "This script will set up example providers."
echo "You will need to provide your API keys."
echo ""

# Check if server is running
if ! curl -s "$BASE_URL/health" > /dev/null 2>&1; then
    echo "❌ Error: Server is not running at $BASE_URL"
    echo "Please start the server first with: ./claraverse-server"
    exit 1
fi

echo "✅ Server is running"
echo ""

# Setup OpenAI
echo "Would you like to add OpenAI? (y/n)"
read -r add_openai

if [ "$add_openai" = "y" ] || [ "$add_openai" = "Y" ]; then
    echo "Enter your OpenAI API key:"
    read -r openai_key

    if [ -n "$openai_key" ]; then
        provider_id=$(add_provider "OpenAI" "https://api.openai.com/v1" "$openai_key")

        if [ "$provider_id" != "0" ]; then
            echo "   Adding filters to show only GPT-4 and GPT-3.5-turbo models..."
            add_filter "$provider_id" "gpt-4o" "include" 20
            add_filter "$provider_id" "gpt-4o-mini" "include" 15
            add_filter "$provider_id" "gpt-3.5-turbo" "include" 10
            refresh_models "$provider_id"
        fi
    fi
    echo ""
fi

# Setup Anthropic
echo "Would you like to add Anthropic? (y/n)"
read -r add_anthropic

if [ "$add_anthropic" = "y" ] || [ "$add_anthropic" = "Y" ]; then
    echo "Enter your Anthropic API key:"
    read -r anthropic_key

    if [ -n "$anthropic_key" ]; then
        provider_id=$(add_provider "Anthropic" "https://api.anthropic.com/v1" "$anthropic_key")

        if [ "$provider_id" != "0" ]; then
            echo "   Adding filters to show Claude 3 models..."
            add_filter "$provider_id" "claude-3-*" "include" 10
            refresh_models "$provider_id"
        fi
    fi
    echo ""
fi

# Setup Z.AI
echo "Would you like to add Z.AI? (y/n)"
read -r add_zai

if [ "$add_zai" = "y" ] || [ "$add_zai" = "Y" ]; then
    echo "Enter your Z.AI API key:"
    read -r zai_key

    if [ -n "$zai_key" ]; then
        provider_id=$(add_provider "Z.AI" "https://api.z.ai/api/coding/paas/v4" "$zai_key")

        if [ "$provider_id" != "0" ]; then
            echo "   Adding filters to show GLM-4 models..."
            add_filter "$provider_id" "glm-4*" "include" 10
            refresh_models "$provider_id"
        fi
    fi
    echo ""
fi

# Custom provider
echo "Would you like to add a custom OpenAI-compatible provider? (y/n)"
read -r add_custom

if [ "$add_custom" = "y" ] || [ "$add_custom" = "Y" ]; then
    echo "Enter provider name:"
    read -r custom_name

    echo "Enter base URL (e.g., https://api.example.com/v1):"
    read -r custom_url

    echo "Enter API key:"
    read -r custom_key

    if [ -n "$custom_name" ] && [ -n "$custom_url" ] && [ -n "$custom_key" ]; then
        provider_id=$(add_provider "$custom_name" "$custom_url" "$custom_key")

        if [ "$provider_id" != "0" ]; then
            refresh_models "$provider_id"
        fi
    fi
    echo ""
fi

echo "=============================="
echo "✅ Setup Complete!"
echo ""
echo "You can now:"
echo "  • List all models: curl $BASE_URL/api/models"
echo "  • List providers: curl $BASE_URL/api/providers"
echo "  • View API documentation: cat API_DOCUMENTATION.md"
echo ""
