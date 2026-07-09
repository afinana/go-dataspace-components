#!/bin/bash
# Script to compile, test, and boot up the Sovereign Dataspace Connector stack using Docker.
set -e

# Target workspace directory context
PROJECT_DIR="/home/afinana/development/github/go-dataspace-components"
cd "$PROJECT_DIR"

echo "================================================================="
echo "   Sovereign Dataspace Connector - Build & Start Orchestrator    "
echo "================================================================="
echo ""

echo ">>> [1/4] Running Go unit tests..."
go test ./...
echo "✔ All tests passed successfully."
echo ""

echo ">>> [2/4] Verifying local Go package builds..."
go build ./cmd/...
echo "✔ Codebase compiled cleanly."
echo ""

echo ">>> [3/4] Building and launching containers in detached mode..."
# Stop and clean up any pre-existing stack instances
docker compose down --remove-orphans

# Boot the Docker stack
docker compose up --build -d
echo "✔ Containers constructed and booted successfully."
echo ""

echo ">>> [4/4] Verifying container states and port bindings..."
# Pause briefly to allow health checks to complete
sleep 4
echo ""
docker compose ps
echo ""

echo "================================================================="
echo "   Connector Stack endpoints are active:                         "
echo "   - Identity Hub:   http://localhost:8080                       "
echo "   - Control Plane:  http://localhost:8081 (incl. DCAT Catalog)   "
echo "   - Data Plane:     http://localhost:8082                       "
echo "   - Data Dashboard: http://localhost:8084                       "
echo "   - PostgreSQL:     localhost:5432                              "
echo "================================================================="
echo "   To view logs in real-time, execute:                           "
echo "   $ docker compose logs -f                                      "
echo "================================================================="
