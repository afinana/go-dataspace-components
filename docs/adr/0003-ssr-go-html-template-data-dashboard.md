# ADR-0003: SSR Go HTML Template Data Dashboard & Control Plane Management UI

## Status

Accepted

## Context

The monorepo contains a containerised dashboard service (`data-dashboard`) to administer local connector instances. Modern web dashboards are often constructed as React/Vue/Svelte Single Page Applications (SPAs). However, these require Node.js, package builders (Webpack, Vite), complex state synchronization, and multiple client-side bundle distributions, which increases repository size and dependencies.

Furthermore, the EDC Management UI must interface cleanly with the EDC Management API, acting as a 1:1 graphical wrapper over the Control Plane to manage data assets, usage policies, contract definitions, catalog queries, and transfer execution.

## Decision

1. We use Server-Side Rendered (SSR) Go `html/template` packages combined with native Vanilla CSS/JS and dynamic AJAX handlers for the `data-dashboard` service.
2. Configuration descriptors are decoupled into static JSON files served via `/public/config/*`:
   - `public/config/edc-connector-config.json`: Array definitions for known EDC instances (target Control Plane, Data Plane, Catalog, Identity Hub URLs, and API keys).
   - `public/config/app-config.json`: UI parameters including theme defaults, titles, `healthCheckIntervalSeconds` (default: 30s), and `enableUserConfig`.
3. The dashboard UI implements four core EDC functional workflows:
   - **Assets Management (`/assets`)**: Asset metadata registration + Data Address (HttpData, AmazonS3, AzureStorage) sending `POST /v3/assets`.
   - **Policy Definitions (`/policies`)**: ODRL permissions, prohibitions, obligations, and constraint building (`eq`, `in`, `neq`).
   - **Contract Definitions (`/contract-definitions`)**: Publishing rules linking Access Policies, Contract Policies, and Asset Selector criteria (`asset:prop:id = asset-01-sensor-data`).
   - **Federated Catalog (`/catalog`) & Transfers (`/transfers`)**: Remote DSP catalog querying, Contract Negotiation state machine execution (`REQUESTED` &rarr; `AGREED`), and Data Plane Transfer initiation (HTTP Pull stream vs S3/Azure Storage push).
4. Top-bar controls support active node picker context switching, dynamic health probes (CP, DP, ID, CAT) polling every `healthCheckIntervalSeconds`, theme toggling (Dark, Light, Custom Cyber), and security warnings when `enableUserConfig` is active.

## Consequences

*   **Zero Node/Build overhead**: The entire dashboard service builds as a single static Go binary. No node_modules, webpack, npm build or compile stages are required inside Docker files.
*   **Constant memory footprints**: Pages are compiled server-side and served instantly.
*   **Simple state modeling**: Active connections are loaded from JSON files and rendered directly with client-side fallback handling.
*   **Security awareness**: Explicit UI warning chips alert administrators when local `localStorage` authorization key persistence is enabled.
*   **AJAX Enhancement**: Interactivity (like health checks, policy updates, DSP negotiations, and progress wizards) is enhanced client-side using native modern JS (`fetch`, DOM queries) without requiring heavy SPA frameworks.
