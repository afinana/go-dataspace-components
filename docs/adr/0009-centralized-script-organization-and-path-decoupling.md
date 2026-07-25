# ADR-0009: Centralized Script Organization, Dynamic Workspace Resolution, and Makefile Automation

## Status

Accepted

## Context

As the Sovereign Dataspace Connector expanded, multiple root-level shell scripts (`start.sh`, `test-dataspace.sh`, `run-bruno-tests.sh`) were created to manage local container orchestration, E2E integration testing, and Bruno API collection runner workflows. 

Having shell scripts located directly in the workspace root caused:
1. Root directory clutter and lack of clear script modularity.
2. Hardcoded absolute paths (e.g. `/home/afinana/...`), which caused failures when executed in different environments, developer machines, or CI/CD runners.
3. Inconsistent execution behavior depending on whether a user called a script from the project root or from a subdirectory.

## Decision

We establish the following script architecture and workflow standards:

### 1. Centralized Script Storage (`/scripts`)
All executable automation, integration testing, setup, and quality gate scripts are consolidated into the `/scripts` directory:
- `scripts/start.sh`: Compiles local Go binaries, executes unit test suites, cleans up legacy containers, and boots up the Docker Compose stack.
- `scripts/test-dataspace.sh`: Executes end-to-end integration testing across Identity Hub, Control Plane, Data Plane, and Data Dashboard endpoints.
- `scripts/run-bruno-tests.sh`: Orchestrates stack initialization, service health verification, and Bruno API collection testing via CLI.
- `scripts/check_coverage.sh`: Verifies global and core domain code coverage thresholds.
- `scripts/setup_hooks.sh`: Installs Git pre-commit quality gate hooks.

### 2. Portable Dynamic Path Resolution
All shell scripts must determine the project root dynamically using relative BASH source traversal:
```bash
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_DIR"
```
This guarantees that invoking any script (e.g. `./scripts/start.sh` or `cd scripts && ./start.sh`) reliably sets the current working directory to the monorepo root before executing relative commands.

### 3. Makefile Automation & Target Alignment
Makefile entrypoints are updated to wrap all script invocations:
- `make start` $\rightarrow$ `scripts/start.sh`
- `make test-dataspace` $\rightarrow$ `scripts/test-dataspace.sh`
- `make bruno-tests` $\rightarrow$ `scripts/run-bruno-tests.sh`
- `make coverage` $\rightarrow$ `scripts/check_coverage.sh`
- `make setup-hooks` $\rightarrow$ `scripts/setup_hooks.sh`

## Consequences

* Ensures a clean root workspace structure conforming to Go monorepo standards.
* Removes hardcoded path assumptions, enabling smooth execution across developer environments, Docker containers, and CI pipelines.
* Simplifies developer operations through standardized `make` targets.
