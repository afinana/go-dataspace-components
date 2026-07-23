# ADR-0004: Constant Memory Data Plane Egress Streaming

## Status

Accepted

## Context

The data plane serves as the secure egress gate for streaming data flows between provider backends and consumers. Loading large file payloads or asset binaries directly into application heap memory can lead to resource exhaustion (Out Of Memory crashes) and limits throughput.

## Decision

We will stream data transfers through the Data Plane using buffered chunking (`io.CopyBuffer`) with a strict, fixed memory buffer limit (typically 32KB-64KB). Whole payloads must never be buffered in memory.

## Consequences

*   **Stable Footprint**: RAM usage remains constant regardless of the file size (from few bytes to multi-gigabyte datasets).
*   **Backpressure Handling**: Data is streamed asynchronously, letting network sockets manage backpressure naturally.
*   **Security compliance**: Eliminates temporary local file creation, preventing path traversal vulnerabilities and disk exhaustion exploits.
