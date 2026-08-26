#!/usr/bin/env bash

set -euo pipefail

cd /app

echo "waiting for mysql..."
until nc -z mysql 3306; do
  sleep 2
done

if [ ! -d /app/web/node_modules ]; then
  echo "installing frontend dependencies..."
  cd /app/web
  npm ci
  cd /app
fi

shutdown() {
  if [ -n "${BACKEND_PID:-}" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    kill "$BACKEND_PID" 2>/dev/null || true
  fi
  if [ -n "${FRONTEND_PID:-}" ] && kill -0 "$FRONTEND_PID" 2>/dev/null; then
    kill "$FRONTEND_PID" 2>/dev/null || true
  fi
  wait || true
}

trap shutdown EXIT INT TERM

go run ./cmd/crow -conf ./configs/docker-config.yaml &
BACKEND_PID=$!

cd /app/web
npm run dev -- --host 0.0.0.0 --port 5173 &
FRONTEND_PID=$!

wait -n "$BACKEND_PID" "$FRONTEND_PID"
