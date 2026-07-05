# Project Rules & Guidelines for Custom Dataspace Connector

This document defines coding standards, behavioral constraints, and architectural boundaries for AI agents contributing to this Go Dataspace Connector repository.

---

## 1. Architectural Integrity

*   **Hexagonal Architecture Boundaries**: 
    Keep `domain` packages (e.g., `control-plane/domain`, `catalog/domain`) strictly free of framework, router, or database engine dependencies. The domain must only define pure business models, state machines, and ports (interfaces).
*   **Decoupled Components**:
    Do not allow direct coupling between component implementations. Communication between the Control Plane, Data Plane, Catalog, and Identity Hub must occur via defined ports or standard network signaling APIs.
*   **Explicit Interface Definitions**:
    Always define explicit interface boundaries for repositories, crypto-helpers, and external protocol clients to enable testing and easy adapter swapping.

---

## 2. Coding Patterns & Constraints

*   **Go Version & Idioms**:
    Write idiomatic Go (targeting version 1.26+ for Docker builds, 1.22+ for local development). Use standard library capabilities where possible (e.g., standard HTTP routing features).
*   **Structured Logging**:
    Always log using the standard library's `log/slog` structured logging. Use context propagation (`logging.WithContext(ctx, logger)`) to associate correlation IDs and trace IDs with log records.
*   **OpenTelemetry Integration**:
    Wrap operations in telemetry spans where appropriate. Do not bypass the global Tracer and Meter providers initialized in `internal/pkg/telemetry`.
*   **Security & Input Validation**:
    *   **Path Traversal Prevention**: When accessing local file paths, always sanitize inputs using `filepath.Clean()`.
    *   **Credential Handling**: Never log, hardcode, or leak raw API keys, private keys, or tokens. Pull them dynamically from `DataAddress` properties or secure config vaults.

---

## 3. State Machine Operations

*   **Transition Constraints**:
    Any state modification on `ContractNegotiation` or `TransferProcess` models must pass through their respective `Transition()` checks. Do not write direct state assignments bypassing these rules.
*   **Error Logging**:
    When transitioning states to `StateTerminated` or `StateTransferTerminated`, always populate the `ErrorDetail` field with descriptive contexts for diagnostic tracing.

---

## 4. Data Plane Egress/Ingress Mechanics

*   **Constant Memory Footprint**:
    When streaming files or object storage assets, never load entire blocks or files into memory. Use buffered chunking via `io.CopyBuffer` with a fixed buffer size (e.g. `64KB`).
*   **Dynamic Proxying**:
    All API Proxy operations must validate incoming security tokens and dynamically inject authentication headers prior to dispatching queries to target provider backends.

---

## 5. Verification Policy

*   **Verification Commands**:
    Before completing any changes, ensure all packages compile and run tests successfully using:
    ```bash
    go build ./... && go test ./...
    ```

---

## 6. Agent Roles & Collaboration Profiles

When contributing to this workspace, AI agents should align themselves with one of the following personas depending on the tasks assigned:

### A. Architect
*   **Focus**: Architectural alignment, component decoupling, security controls, and standards conformance (W3C DCAT-AP, W3C DID, ODRL, DSP).
*   **Behavior**: Inspects system designs and reviews dependencies. Does not perform structural coding, but reviews interfaces and package borders.

### B. Developer
*   **Focus**: Component coding, state machine implementation, data stream structures, and port adapters.
*   **Behavior**: Writes type-safe Go, formats code according to standard idioms, implements proxy directors, and optimizes streaming buffer pipelines.

### C. QA Tester
*   **Focus**: Code testing, build validation, regression checking, and container topology health.
*   **Behavior**: Writes Go `_test.go` files, runs local tests, creates integration scripts, checks docker-compose port binding states, and tests endpoints.
