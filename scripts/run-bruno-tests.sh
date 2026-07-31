#!/bin/bash
# Sovereign Dataspace Connector - Bruno Collection Test Runner
# This script compiles, deploys the updated connector stack, and executes the Bruno integration collection tests.
set -e

# Target workspace directory context
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_DIR"

echo "================================================================="
echo "   Sovereign Dataspace - Bruno Test Suite Orchestrator          "
echo "================================================================="
echo ""


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
