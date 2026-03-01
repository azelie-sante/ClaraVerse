#!/bin/bash

# Configuration
export TOKEN="eyJhbGciOiJIUzI1NiIsImtpZCI6Ik80ZEtMNTB1YityUmo4N2ciLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL29jcW9xamFmbWp1aXl3c3Bwd2t3LnN1cGFiYXNlLmNvL2F1dGgvdjEiLCJzdWIiOiJjNDFmZGM1YS01YzA0LTQxMmUtYWQ3Ny0xNzM0YjlkNTgyYmYiLCJhdWQiOiJhdXRoZW50aWNhdGVkIiwiZXhwIjoxNzY3MjI0NzY4LCJpYXQiOjE3NjcyMjExNjgsImVtYWlsIjoiYXJ1bnRlbW1lQGdtYWlsLmNvbSIsInBob25lIjoiIiwiYXBwX21ldGFkYXRhIjp7InByb3ZpZGVyIjoiZW1haWwiLCJwcm92aWRlcnMiOlsiZW1haWwiLCJnaXRodWIiXX0sInVzZXJfbWV0YWRhdGEiOnsiYXZhdGFyX3VybCI6Imh0dHBzOi8vYXZhdGFycy5naXRodWJ1c2VyY29udGVudC5jb20vdS80MTExODE4Nz92PTQiLCJkaXNwbGF5X25hbWUiOiJhcnVudGVtbWUiLCJlbWFpbCI6ImFydW5nYXV0aGFta0BnbWFpbC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZnVsbF9uYW1lIjoiQXJ1biIsImlzcyI6Imh0dHBzOi8vYXBpLmdpdGh1Yi5jb20iLCJuYW1lIjoiQXJ1biIsInBob25lX3ZlcmlmaWVkIjpmYWxzZSwicHJlZmVycmVkX3VzZXJuYW1lIjoiYXJ1bnRlbW1lIiwicHJvdmlkZXJfaWQiOiI0MTExODE4NyIsInN1YiI6IjQxMTE4MTg3IiwidXNlcl9uYW1lIjoiYXJ1bnRlbW1lIiwidXNlcm5hbWUiOiJhcnVudGVtbWUifSwicm9sZSI6ImF1dGhlbnRpY2F0ZWQiLCJhYWwiOiJhYWwxIiwiYW1yIjpbeyJtZXRob2QiOiJwYXNzd29yZCIsInRpbWVzdGFtcCI6MTc2NzIyMTE2OH1dLCJzZXNzaW9uX2lkIjoiZjUxMjE2ZmMtYTljOC00NGVmLWI5ZjEtNGVmOWU1N2NmODUwIiwiaXNfYW5vbnltb3VzIjpmYWxzZX0.NIfXpOQfHQ3_9C8ntfwMxybV75SppgT6HTZ4lJenZZU"
export TARGET_USER="30c8850f-bec2-47ef-8646-be4afbbfdb9e"
BASE_URL="http://localhost:3001/api/admin"

# echo "=== Testing Admin Endpoints ==="
# echo ""

# echo "1. Getting user details..."
# curl -s -X GET -H "Authorization: Bearer $TOKEN" "$BASE_URL/users/$TARGET_USER" | jq

echo ""
echo "2. Setting tier override to pro..."
curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"tier":"free","reason":"Test"}' "$BASE_URL/users/$TARGET_USER/overrides" | jq

# echo ""
# echo "3. Verifying tier override..."
# curl -s -X GET -H "Authorization: Bearer $TOKEN" "$BASE_URL/users/$TARGET_USER" | jq '.effective_tier, .has_tier_override'

# echo ""
# echo "4. Setting granular overrides..."
# curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
#   -d '{"limits":{"maxMessagesPerMonth":-1},"reason":"Unlimited messages test"}' \
#   "$BASE_URL/users/$TARGET_USER/overrides" | jq

# echo ""
# echo "5. Verifying granular overrides..."
# curl -s -X GET -H "Authorization: Bearer $TOKEN" "$BASE_URL/users/$TARGET_USER" | jq '.effective_limits.maxMessagesPerMonth, .has_limit_overrides'

# echo ""
# echo "6. Removing all overrides..."
# curl -s -X DELETE -H "Authorization: Bearer $TOKEN" "$BASE_URL/users/$TARGET_USER/overrides" | jq

# echo ""
# echo "7. Verifying overrides removed..."
# curl -s -X GET -H "Authorization: Bearer $TOKEN" "$BASE_URL/users/$TARGET_USER" | jq '.has_tier_override, .has_limit_overrides'

echo ""
echo "=== Tests Complete ==="
