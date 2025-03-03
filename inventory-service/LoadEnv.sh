#!/bin/bash
# Check if .env exists
if [ -f .env ]; then
  # Read .env line by line
  while IFS='=' read -r key value; do
    # Skip empty lines or comments
    if [[ -z "$key" || "$key" =~ ^# ]]; then
      continue
    fi
    # Trim whitespace from key and value
    key=$(echo "$key" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    value=$(echo "$value" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    # Export the variable
    export "$key=$value"
  done < .env
fi
