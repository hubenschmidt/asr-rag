#!/bin/bash

WHISPER_BIN=~/.local/bin/whisper-server
WHISPER_MODEL=~/.local/share/whisper/ggml-medium.bin
WHISPER_PORT=8178

# Start Qdrant
docker compose up -d

# Start whisper-server in the background if not already running
if ! curl -s "http://localhost:$WHISPER_PORT/health" > /dev/null 2>&1; then
    echo "starting whisper-server on :$WHISPER_PORT..."
    $WHISPER_BIN -m $WHISPER_MODEL --port $WHISPER_PORT &
    WHISPER_PID=$!
    echo "whisper-server pid: $WHISPER_PID"
    # Wait for it to be ready
    for i in $(seq 1 30); do
        curl -s "http://localhost:$WHISPER_PORT/health" > /dev/null 2>&1 && break
        sleep 1
    done
    echo "whisper-server ready"
else
    echo "whisper-server already running on :$WHISPER_PORT"
fi

echo "all services ready"
