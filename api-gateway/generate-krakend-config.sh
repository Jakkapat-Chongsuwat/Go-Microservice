#!/bin/bash

# Set default values for environment variables
export ORDER_SERVICE_URL=${ORDER_SERVICE_URL:-"http://order-service:60052"}
export USER_SERVICE_URL=${USER_SERVICE_URL:-"http://user-service:50052"}
export INVENTORY_SERVICE_URL=${INVENTORY_SERVICE_URL:-"http://inventory-service:30052"}

echo "Generating Kubernetes ConfigMap with these service URLs:"
echo "ORDER_SERVICE_URL: $ORDER_SERVICE_URL"
echo "USER_SERVICE_URL: $USER_SERVICE_URL"
echo "INVENTORY_SERVICE_URL: $INVENTORY_SERVICE_URL"

# Create k8s directory if it doesn't exist
mkdir -p k8s
echo "Created directory: k8s (if it didn't exist)"

# Create a temporary file for the processed template
echo "Processing template..."
cat krakend.json.tpl | sed "s|\$ORDER_SERVICE_URL|$ORDER_SERVICE_URL|g" | \
  sed "s|\$USER_SERVICE_URL|$USER_SERVICE_URL|g" | \
  sed "s|\$INVENTORY_SERVICE_URL|$INVENTORY_SERVICE_URL|g" > krakend.json.tmp

# Create the ConfigMap file
echo "Creating ConfigMap YAML..."
cat > k8s/krakend-configmap.yaml << EOF_YAML
apiVersion: v1
kind: ConfigMap
metadata:
  name: krakend-config
data:
  krakend.json: |
EOF_YAML

# Add the processed JSON to the ConfigMap with proper indentation
sed 's/^/    /' krakend.json.tmp >> k8s/krakend-configmap.yaml

# Clean up temporary file
rm krakend.json.tmp

echo "ConfigMap generated successfully at k8s/krakend-configmap.yaml"
