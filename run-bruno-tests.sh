#!/bin/bash
# Sovereign Dataspace Connector - Bruno Collection Test Runner
# This script compiles, deploys the updated connector stack, and executes the Bruno integration collection tests.
set -e

# Target workspace directory context
PROJECT_DIR="/home/afinana/development/github/go-dataspace-components"
cd "$PROJECT_DIR"

echo "================================================================="
echo "   Sovereign Dataspace - Bruno Test Suite Orchestrator          "
echo "================================================================="
echo ""

# 1. Build and boot the stack with latest endpoints
echo ">>> [1/3] Building and starting local container stack..."
./start.sh

# 2. Wait for health checks
echo ">>> [2/3] Waiting for services to become healthy..."
for i in {1..15}; do
  if curl -s http://localhost:8081/health >/dev/null && curl -s http://localhost:8082/health >/dev/null && curl -s http://localhost:8080/credentials >/dev/null; then
    echo "✔ All services are healthy!"
    break
  fi
  echo "Waiting for services... (attempt $i/15)"
  sleep 2
done

# 3. Run Bruno integration tests
echo ""
echo ">>> [3/3] Executing Bruno collection tests..."
cd Requests
if command -v bru &> /dev/null; then
  bru run --env local
else
  echo "bru CLI is not installed globally. Running via npx..."
  npx -y @usebruno/cli run --env local
fi

echo ""
echo "================================================================="
echo "   Bruno Collection Integration Tests Completed Successfully!    "
echo "================================================================="
