# 10. Consumer-Side Control & Data Plane Architecture & DSP 2025-1 Protocol Compliance

*   **Status**: Accepted
*   **Date**: 2026-07-31
*   **Deciders**: Dataspace Architecture Team

---

## Context

The Go Dataspace Connector repository initially provided a provider-centric implementation. To enable true peer-to-peer dataspace capabilities, contract negotiation state machines, and complete end-to-end testing, the connector needed a dedicated, persistent **Consumer Control Plane** and **Consumer Data Plane**, as well as an upgrade to the latest **W3C Dataspace Protocol (DSP) 2025-1** specification (`https://w3id.org/dspace/2025/1/context/`).

Prior to this decision:
1. Consumer functionality existed only as transient client test scripts (`cmd/consumer/main.go`).
2. Provider protocol endpoints used non-standard paths (`/protocol/...`) and lacked consistent DSP 2025-1 JSON-LD context payloads.
3. State machines for contract negotiation lacked explicit support for `OFFERED` and `ACCEPTED` states, while transfer processes lacked `SUSPENDED` states.

---

## Decision

We decided to implement a full-featured, persistent Consumer Module and upgrade protocol signaling across all components to DSP 2025-1 compliance:

1. **Persistent Consumer Control Plane (`cmd/consumer`, port 8091)**:
   - Exposes DSP Callback Endpoints (`/consumer/negotiations/...`, `/consumer/transfers/...`) to receive async protocol state events from provider connectors.
   - Exposes a dedicated Consumer Management API (`/api/consumer/v4/...`) for orchestrating catalog queries, initiating contract negotiations, tracking agreement lifecycles, and managing data transfers.
   - Reuses hexagonal domain models with dedicated consumer persistence stores (`consumer_contract_negotiations` and `consumer_transfer_processes`).

2. **Consumer Data Plane & Ingress Proxy (`docker/consumer-data-plane.Dockerfile`, port 8092)**:
   - Provides a separate consumer-side data plane instance capable of receiving data transfers and serving consumer-side proxy requests.

3. **DSP 2025-1 Domain Upgrade**:
   - Expanded `ContractNegotiation` state machine with `OFFERED` and `ACCEPTED` states.
   - Expanded `TransferProcess` state machine with `SUSPENDED` state.
   - Provider DSP endpoint routing normalized under `/api/dsp/2025-1/...` with JSON-LD `@context` (`https://w3id.org/dspace/2025/1/context/`) added across all protocol messages.

4. **Multi-Service Topology & Infrastructure**:
   - Updated `docker-compose.yml` to launch a 7-service topology (Identity Hub, Provider CP/DP, Consumer CP/DP, Dashboard, Postgres).
   - Expanded Terraform configuration with dedicated `consumer` and `consumer_data_plane` modules for Kubernetes deployments.
   - Reorganized Bruno API requests into distinct `Consumer Management/` and `Provider Management/` collections.

---

## Consequences

### Positive
* Enables automated peer-to-peer end-to-end contract negotiations and data transfers without third-party mock tools.
* Full alignment with DSP 2025-1 protocol specifications and JSON-LD contexts.
* Clean separation of concerns between provider-side catalog/egress logic and consumer-side callback/ingress logic.
* High container and Kubernetes deployment parity via modular Dockerfiles and Terraform declarations.

### Negative / Trade-offs
* Higher resource consumption during Docker Compose integration tests due to 7 concurrently running services.
