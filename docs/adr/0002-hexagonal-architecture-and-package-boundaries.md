# ADR-0002: Hexagonal Architecture and Package Boundaries

## Status

Accepted

## Context

A core architectural principle of dataspace connectors is decoupled domain design to enable modular service configuration, multi-database support, and protocol flexibility. Dependencies between system services (Control Plane, Data Plane, Catalog, and Identity Hub) must be minimised to enable testing and easy adapters swap.

## Decision

We will strictly enforce Hexagonal Architecture boundaries in all codebase components:
1. **Domain package** (e.g. `/control-plane/domain`, `/catalog/domain`): Must contain only pure business entities, core states, constraints, transition logic, and repository interfaces (ports). They must have **zero** dependencies on HTTP routers, SQL drivers, ORMs, or other external components.
2. **Ports/Adapters package** (e.g. `/control-plane/ports`, `/catalog/ports`): Implements repository interfaces, network signaling endpoints, and framework adapters (like SQL queries, standard library routers, or JSON-LD serializers).

## Consequences

*   High level of testability: Domain business logic can be tested in isolation using mock structures without launching database engines or servers.
*   Framework independence: We can replace PostgreSQL repositories with In-Memory storage or other databases without altering core domain code.
*   Explicit dependencies: Package imports remain unidirectional from adapters/ports to domain packages.
