# ADR-0006: High-Performance L1 Key-Value Store Caching Layer

## Status

Accepted

## Context

Repeated catalog queries, dataset metadata requests, and remote W3C DID document resolutions over HTTP introduce latency and database load across the Sovereign Dataspace Connector stack. An in-memory L1 key-value cache layer with automatic TTL eviction and thread safety is needed to boost read throughput and eliminate redundant database/network operations.

## Decision

We will implement a thread-safe `MemoryKVStore` under `internal/pkg/kvstore` featuring TTL expiration and background eviction. We will integrate this KV store into key high-frequency access components:

1.  **PostgresCatalogStore**: Caches dataset lookups (`dataset:<id>`), dataset lists (`datasets:all`), and catalog definitions (`catalog:main`). Write or delete operations trigger targeted cache updates and invalidation.
2.  **DIDWebResolver**: Caches resolved W3C DID Documents (`did:<id>`) to minimize remote HTTPS network round-trips during capability verification.

## Consequences

*   **Sub-millisecond Latency**: Cached dataset and DID queries bypass PostgreSQL and HTTP network round-trips entirely.
*   **Reduced Database Load**: Eliminates redundant database reads for high-frequency catalog catalog requests.
*   **Automatic Cache Invalidation**: Write/Delete operations automatically invalidate affected cache keys to maintain data consistency.
