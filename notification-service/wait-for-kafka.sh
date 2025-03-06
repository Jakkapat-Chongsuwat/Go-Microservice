#!/bin/sh
TIMEOUT=60
while ! nc -z kafka $KAFKA_PORT; do
    echo "Waiting for Kafka to be ready..."
    sleep 2
    TIMEOUT=$((TIMEOUT-2))
    if [ $TIMEOUT -le 0 ]; then
        echo "Kafka did not become ready in time."
        exit 1
    fi
done

echo "Kafka is ready. Starting notification service..."
exec ./notification-service
