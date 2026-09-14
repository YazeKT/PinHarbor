#!/usr/bin/env bash
set -euo pipefail

app_dir="$(cd "$(dirname "$0")" && pwd)"
cd "$app_dir"
mkdir -p bin webdata

echo "Building PinHarbor..."
go build -o bin/pinharbor .

if [[ ! -f webdata/.playwright-ready ]]; then
  echo "Installing the browser engine (first run only)..."
  PLAYWRIGHT_INSTALL_ONLY=1 bin/pinharbor
  touch webdata/.playwright-ready
fi

opened=0
while true; do
  echo "Starting PinHarbor at http://localhost:8080"
  bin/pinharbor -web -data-folder "$app_dir/webdata" &
  app_pid=$!

  if [[ "$opened" -eq 0 ]]; then
    for _ in $(seq 1 45); do
      if curl --silent --fail http://127.0.0.1:8080/health >/dev/null; then
        command -v xdg-open >/dev/null && xdg-open http://localhost:8080 >/dev/null 2>&1 || true
        opened=1
        break
      fi
      sleep 1
    done
  fi

  wait "$app_pid" || true
  echo "PinHarbor stopped unexpectedly; restarting in 3 seconds..."
  sleep 3
done
