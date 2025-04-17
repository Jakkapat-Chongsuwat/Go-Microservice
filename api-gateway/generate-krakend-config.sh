#!/bin/bash

# Default service URLs - these use fully qualified Kubernetes service names
export ORDER_SERVICE_URL=${ORDER_SERVICE_URL:-"http://order-service.order-service.svc.cluster.local:60052"}
export USER_SERVICE_URL=${USER_SERVICE_URL:-"http://user-service.user-service.svc.cluster.local:50052"}
export INVENTORY_SERVICE_URL=${INVENTORY_SERVICE_URL:-"http://inventory-service.inventory-service.svc.cluster.local:30052"}

echo "Generating configuration using:"
echo "- ORDER_SERVICE_URL: $ORDER_SERVICE_URL"
echo "- USER_SERVICE_URL: $USER_SERVICE_URL"
echo "- INVENTORY_SERVICE_URL: $INVENTORY_SERVICE_URL"

# Create k8s directory if it doesn't exist
mkdir -p k8s

# Process the template
envsubst < krakend.json.tpl > krakend.json

# Create the ConfigMap YAML (for reference purposes only)
cat > k8s/krakend-configmap.yaml << EOF_YAML
apiVersion: v1
kind: ConfigMap
metadata:
  name: krakend-config
  namespace: api-gateway
data:
  krakend.json.tpl: |
EOF_YAML

# Add the template to the ConfigMap
sed 's/^/    /' krakend.json.tpl >> k8s/krakend-configmap.yaml

echo "Configuration generated successfully"
