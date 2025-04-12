#!/bin/sh
set -e

# Check if we're using environment variables or direct config
if [ -f "/etc/krakend/krakend.json.tpl" ] && [ ! -f "/etc/krakend/krakend.json" ]; then
    echo "Processing template configuration..."
    envsubst < /etc/krakend/krakend.json.tpl > /etc/krakend/krakend.json
    echo "Template processed successfully."
elif [ ! -f "/etc/krakend/krakend.json" ]; then
    echo "ERROR: No configuration file found at /etc/krakend/krakend.json"
    exit 1
else
    echo "Using provided configuration file."
fi

# Start KrakenD with the configuration
exec krakend run -c /etc/krakend/krakend.json "$@"
