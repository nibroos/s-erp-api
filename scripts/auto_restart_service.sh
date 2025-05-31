#!/bin/bash

HEALTH_URL="http://localhost:4001/api/v1/health"
CONSUMER_HEALTH_URL="http://localhost:4010/health"
CONTAINER_NAME="s-erp-api-service-1"
CONSUMER_CONTAINER_NAME="s-erp-api-consumer-service-1"
INTERVAL=5 # seconds

while true; do
  rest_response=$(curl -s -o /dev/null -w "%{http_code}" "$HEALTH_URL")
  consumer_response=$(curl -s -o /dev/null -w "%{http_code}" "$CONSUMER_HEALTH_URL")

  if [ "$rest_response" != "200" ]; then
    echo "$(date) REST API health check failed (HTTP $rest_response). Restarting container $CONTAINER_NAME..."
    docker restart "$CONTAINER_NAME"
  else
    echo "$(date) REST API health check passed (HTTP 200)."
  fi

  if [ "$consumer_response" != "200" ]; then
    echo "$(date) Consumer health check failed (HTTP $consumer_response). Restarting container $CONSUMER_CONTAINER_NAME..."
    docker restart "$CONSUMER_CONTAINER_NAME"
  else
    echo "$(date) Consumer health check passed (HTTP 200)."
  fi

  sleep $INTERVAL
done