# 11. Oxigraph Triple Store and NATS JetStream Event Messaging

* Status: Accepted
* Date: 2026-08-01

## Context and Problem Statement

The Dataspace Connector previously relied solely on relational storage (PostgreSQL) for W3C DCAT-AP catalogs and synchronous HTTP/REST calls for inter-service signaling. W3C DCAT-AP datasets and ODRL policy definitions are natively RDF graphs. Additionally, as signaling volume grows, synchronous HTTP signaling creates tight coupling and lacks persistent message queues or event trajectory visibility.

## Decision Drivers

* **Native RDF Cataloging**: Ability to execute standard SPARQL 1.1 graph queries over W3C DCAT-AP dataset descriptors.
* **Asynchronous Event-Driven Signaling**: Persistent message queues for contract negotiation state transitions and transfer signaling.
* **Operational Observability**: Visual real-time management of message streams, consumer lag, and event payloads via a UI console.

## Considered Options

1. **PostgreSQL JSONB + Synchronous REST APIs** (Status Quo)
2. **Oxigraph RDF Triple Store + NATS JetStream & JetStream Console**

## Decision Outcome

Chosen option: **Oxigraph RDF Triple Store + NATS JetStream & JetStream Console**.

### Architecture Overview

1. **Oxigraph Triple Store Integration**:
   - Implemented `OxigraphCatalogStore` under `catalog/ports/oxigraph_repository.go` fulfilling `domain.AssetRegistry` and `domain.CatalogQueryService`.
   - Exposes standard SPARQL 1.1 endpoints (`/update` and `/query`) running on port `7878`.

2. **NATS JetStream & NATS Console Integration**:
   - Implemented `internal/pkg/events/nats.go` for event publishing over JetStream subject `dataspace.>`.
   - Configured `nats` (JetStream server on ports 4222/8222) and `nats-console` (NATS NUI on port 8085) services in `docker-compose.yml`.

## Consequences

### Positive
* Enables full SPARQL 1.1 graph querying across dataset catalogs.
* Provides high-throughput, persistent event stream delivery for negotiation and transfer processes.
* Offers real-time graphical monitoring of message streams via NATS JetStream Console on port 8085.

### Negative
* Introduces two additional lightweight container dependencies (`oxigraph` and `nats`).
