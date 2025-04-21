#!/bin/bash

# Function to log messages
log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Check if Minikube is running
if ! minikube status | grep -q "host: Running"; then
  log "Error: Minikube is not running. Please start Minikube first."
  exit 1
fi

# List of microservice images to load
IMAGES=(
  "go-microservice-notification-service:latest"
  "go-microservice-order-service:latest"
  "go-microservice-user-service:latest"
  "go-microservice-api-gateway:latest"
  "go-microservice-inventory-service:latest"
)

# Load each image into Minikube
for image in "${IMAGES[@]}"; do
  log "Loading image: $image into Minikube..."
  minikube image load "$image"
done

log "All images loaded successfully!"
log "You can verify images in Minikube with:"
log "  minikube ssh 'docker images'"
