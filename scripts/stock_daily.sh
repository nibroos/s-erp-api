#!/bin/bash

echo "Running stock_daily.sh"

# Detect OS
OS="$(uname)"
if [[ "$OS" != "Linux" && "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OS" == "Windows_NT" ]]; then
    IS_WINDOWS=true
else
    IS_WINDOWS=false
fi

# Get the script's directory path (works on both Windows and Linux)
if [[ "$IS_WINDOWS" == true ]]; then
    SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -W)
else
    SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
fi

# Calculate the path to the .env file
ENV_FILE="$SCRIPT_DIR/../docker/.env"

# Check if .env file exists
if [ ! -f "$ENV_FILE" ]; then
    echo "Error: .env file not found at $ENV_FILE"
    exit 1
fi

# Source the .env file
set -a
source "$ENV_FILE"
set +a

# Login and get JWT token
LOGIN_RESPONSE=$(curl -s -X POST "${APP_HOST}/api/v1/auth/login" \
     -H "Content-Type: application/json" \
     -d "{
        \"email\": \"${schedule_credentials}\",
        \"password\": \"${schedule_password}\"
     }")

# echo "APP_HOST: $APP_HOST"
# echo "Schedule credentials: $schedule_credentials"
# echo "Schedule password: $schedule_password"
# echo "Login request: ${APP_HOST}/api/v1/auth/login"
# echo "Login data: {\"email\": \"${schedule_credentials}\", \"password\": \"${schedule_password}\"}"
# echo "Login response: $LOGIN_RESPONSE"

# Extract token - Handle both Windows and Linux
if [[ "$IS_WINDOWS" == true ]]; then
    # For Windows, first check for jq in Git Bash or manually installed
    if command -v jq &> /dev/null || [ -f "/c/Program Files/Git/usr/bin/jq.exe" ]; then
        TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.optional.token')
    else
        # PowerShell fallback for Windows
        TOKEN=$(echo $LOGIN_RESPONSE | powershell -command "$input | ConvertFrom-Json | Select-Object -ExpandProperty optional | Select-Object -ExpandProperty token")
    fi
else
    # Linux path
    if ! command -v jq &> /dev/null; then
        echo "Warning: jq is not installed. Using grep fallback method."
        TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    else
        TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.optional.token')
    fi
fi


if [ -z "$TOKEN" ]; then
    echo "Error: Failed to get authentication token"
    echo "Full response was: $LOGIN_RESPONSE"
    exit 1
fi

# API endpoint for stock closing
API_URL="${APP_HOST}/api/v1/inventories/stocks/create-stock-closing"

# body, end_closing_at, password

# Set closing_at as current date
END_CLOSING_AT=$(date +"%Y-%m-%d")

echo "END_CLOSING_AT: $END_CLOSING_AT"

# Make POST request with JWT token
curl -X POST "$API_URL" \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer ${TOKEN}" \
     -d "{\"end_closing_at\": \"$END_CLOSING_AT\", \"password\": \"${RABBITMQ_PASSWORD}\"}" \

echo ""
echo "Stock daily script executed successfully."