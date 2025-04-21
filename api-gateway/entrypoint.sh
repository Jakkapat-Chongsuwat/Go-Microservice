#!/bin/sh
set -e

# Print out the environment variables for debugging
echo "Service URLs:"
echo "ORDER_SERVICE_URL: $ORDER_SERVICE_URL"
echo "USER_SERVICE_URL: $USER_SERVICE_URL"
echo "INVENTORY_SERVICE_URL: $INVENTORY_SERVICE_URL"
echo "NOTIFICATION_SERVICE_URL: $NOTIFICATION_SERVICE_URL"

# If template exists, process it
if [ -f "/etc/krakend/krakend.json.tmpl" ]; then
    echo "Processing krakend.json.tmpl..."
    envsubst < /etc/krakend/krakend.json.tmpl > /tmp/krakend.json
    echo "Template processed successfully."
    
    # Optional: print the generated config for debugging
    if [ "${DEBUG:-false}" = "true" ]; then
        echo "Generated configuration:"
        cat /tmp/krakend.json
    fi
elif [ ! -f "/etc/krakend/krakend.json" ]; then
    echo "ERROR: No configuration file found at /etc/krakend/krakend.json"
    exit 1
else
    echo "Using provided configuration file."
    cp /etc/krakend/krakend.json /tmp/krakend.json
fi

# Start KrakenD with the configuration
exec krakend run -c /tmp/krakend.json "$@"
