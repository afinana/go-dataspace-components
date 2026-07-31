# ADR-0012: Observability Architecture with Prometheus, Grafana, and OpenTelemetry

* **Status**: Accepted
* **Date**: 2026-08-01

## Context

As the Sovereign Dataspace Connector expanded to include multiple microservices (Identity Hub, Control Plane, Data Plane, Catalog, Consumer Services, NATS JetStream, and Oxigraph), comprehensive observability became essential for monitoring system health, request latency, and transfer performance.

## Decision

We incorporate an integrated observability stack comprising:
1. **OpenTelemetry Integration**: Standardized telemetry initialization in `internal/pkg/telemetry`, establishing W3C TraceContext propagation, tracer spans, and duration histogram metrics across core services.
2. **Prometheus Metrics Aggregation**: Scrapes metrics from active dataspace components, JetStream brokers, and gateway endpoints on port `9090`.
3. **Grafana Executive Overview Dashboard**: Visualizes service status (up/down health metrics), request latencies, and system metrics on port `3000` (provisioned via `docker/grafana-datasources.yml` and `docker/grafana-dashboards.yml`).
4. **Terraform Monitoring Module**: Added a dedicated `monitoring` module under `terraform/modules/monitoring/` deploying Prometheus (`30090`) and Grafana (`30000`) Kubernetes resources.

## Consequences

* Provides immediate visual telemetry and metric scraping across local Docker Compose and Kubernetes k3s deployments.
* Ensures uniform metric collection adhering to OpenTelemetry standards without introducing heavy external service dependencies.
