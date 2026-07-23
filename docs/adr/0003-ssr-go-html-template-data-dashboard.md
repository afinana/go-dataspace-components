# ADR-0003: SSR Go HTML Template Data Dashboard

## Status

Accepted

## Context

The monorepo contains a containerised dashboard service (`data-dashboard`) to administer local connector instances. Modern web dashboards are often constructed as React/Vue/Svelte Single Page Applications (SPAs). However, these require Node.js, package builders (Webpack, Vite), complex state synchronization, and multiple client-side bundle distributions, which increases repository size and dependencies.

## Decision

We will use Server-Side Rendered (SSR) Go `html/template` packages combined with native Vanilla CSS/JS and dynamic AJAX handlers for the `data-dashboard` service. 

## Consequences

*   **Zero Node/Build overhead**: The entire dashboard service builds as a single static Go binary. No node_modules, webpack, npm build or compile stages are required inside Docker files.
*   **Constant memory footprints**: Pages are compiled server-side and served instantly. 
*   **Simple state modeling**: Active connections are stored in memory or JSON files and rendered directly.
*   **AJAX Enhancement**: Interactivity (like health checks, policy updates, and progress wizards) is enhanced client-side using native modern JS (`fetch`, DOM queries) without requiring state management frameworks.
