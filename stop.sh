#!/bin/bash

echo "stopping whisper-server..."
pkill -f whisper-server 2>/dev/null || echo "whisper-server not running"

echo "stopping qdrant..."
docker compose down

echo "all services stopped"
