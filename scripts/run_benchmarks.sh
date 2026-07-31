#!/bin/bash
# Sovereign Dataspace Connector - Grafana k6 Performance & Load Benchmark Runner
set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_DIR"

SUITE="${1:-all}"

run_k6() {
    local script="$1"
    if command -v k6 &> /dev/null; then
        echo "🚀 Running k6 locally: $script"
        k6 run "$script"
    else
        echo "ℹ️  k6 CLI not found locally; running k6 inside Docker..."
        docker run --rm -i --net=host -v "$PROJECT_DIR":/workspace -w /workspace grafana/k6:latest run "$script"
    fi
}

run_suite() {
    local target="$1"
    case "$target" in
        gateway)
            run_k6 "scripts/k6/gateway_benchmark.js"
            ;;
        web_gui|gui)
            run_k6 "scripts/k6/web_gui_benchmark.js"
            ;;
        full)
            run_k6 "scripts/k6/full_system_benchmark.js"
            ;;
        all)
            echo "================================================================="
            echo "   Executing Suite [1/3]: Gateway REST API Benchmark             "
            echo "================================================================="
            run_k6 "scripts/k6/gateway_benchmark.js"
            echo ""
            echo "================================================================="
            echo "   Executing Suite [2/3]: Web GUI Dashboard Benchmark           "
            echo "================================================================="
            run_k6 "scripts/k6/web_gui_benchmark.js"
            echo ""
            echo "================================================================="
            echo "   Executing Suite [3/3]: Full End-to-End System Benchmark        "
            echo "================================================================="
            run_k6 "scripts/k6/full_system_benchmark.js"
            ;;
        *)
            echo "Unknown target: $target"
            echo "Usage: $0 [gateway|web_gui|full|all]"
            exit 1
            ;;
    esac
}

run_suite "$SUITE"

echo ""
echo "================================================================="
echo "   Performance & Load Benchmarks Completed Successfully!          "
echo "================================================================="
