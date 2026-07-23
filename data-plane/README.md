# Data Plane Module

The **Data Plane** operates as the high-throughput, secure egress/ingress gate. It manages physical data flows based on Control Plane instructions.

## Core Responsibilities

*   **API Reverse Proxying**: Sanitizes incoming consumer headers and dynamically injects target provider backend authentication credentials.
*   **Constant Memory Chunking**: Implements zero-copy streaming of datasets using standard Go readers/writers and fixed 32KB circular buffers to minimize heap usage.
*   **Signaling Channel Receiver**: Validates initialization requests from the local Control Plane to authorize specific transfer process IDs.
*   **EDR Token Validation**: Decodes and verifies secure Endpoint Data Reference (EDR) bearer tokens before dispatching target streams.

## Architecture

*   `domain/`: Stream context mappings and proxy parameters.
*   `ports/`: HTTP proxy pipelines, file stream writers, and signaling receivers.
