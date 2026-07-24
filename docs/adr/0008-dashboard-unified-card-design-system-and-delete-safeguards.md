# ADR-0008: Dashboard Unified Flat Card Design System, Lifecycle Modals, and DSP Fallback Handling

## Status

Accepted

## Context

The EDC Data Dashboard acts as a graphical management console over the EDC Control Plane API for data assets, ODRL policy definitions, contract publishing rules, federated catalog discovery, and transfer processes. As these management interfaces grew, the UI required:
1. A standardized responsive card design system that prevents mid-word text breaking and scrollbar clutter.
2. Safe lifecycle workflows (View specs, Edit configurations, and Confirm Deletions).
3. Resilience when querying remote Dataspace Protocol (DSP) catalogs across heterogeneous container topologies.
4. Cross-platform font compatibility for headless Linux server environments.

## Decision

We establish the following design and architectural guidelines for the EDC Data Dashboard:

### 1. Unified Flat Card Design System
- **Responsive Card Grid**: Standardized responsive CSS Grid layout (`minmax(340px, 1fr)`) with equal-height card containers (`catalog-card-fixed`, `min-height: 340px`).
- **Text Truncation & Tooltips**: Code snippet badges (`.code-snippet`) and URL endpoints use single-line text truncation (`max-width: 200px`, `text-overflow: ellipsis`) accompanied by HTML hover `title` tooltips to preserve readability without awkward line wrapping.
- **Flat 2D Aesthetic**: Eliminated 3D offset translations (`transform: translateY/X`) and heavy drop shadows in favor of a clean, modern flat aesthetic utilizing themed border highlights (`var(--border-highlight)`) and flat background transitions (`var(--bg-card-hover)`).

### 2. Full Lifecycle Operations (View, Edit, Delete Safeguards)
- **View Modal**: Displays formatted JSON payload definitions (`Asset`, `PolicyDefinition`, `ContractDefinition`) directly from Control Plane models.
- **Edit Modal**: Pre-populates management forms with existing asset IDs (read-only key) and properties to dispatch update requests.
- **Delete Confirmation Modal**: Intercepts delete actions with an explicit confirmation dialog displaying target resource IDs before dispatching destructive HTTP POST requests (`/assets/delete`, `/policies/delete`, `/contract-definitions/delete`).

### 3. DSP Catalog Resilience & Cross-Platform Styling
- **Remote Catalog Querying**: Queries counter-party DSP endpoints (`/api/catalog/query`). When an empty catalog slice is returned from an unpopulated target node, the system gracefully supplies counter-party bound sample datasets so UI exploration is never blocked.
- **CSS-Only Status Icons**: Status indicators, spinners (`@keyframes spin`), and negotiation success/failure badges (`✓`, `✕`) use CSS-rendered vector shapes and standard ASCII/SVG glyphs, eliminating broken font box rendering (`[🅇]`) on headless Linux environments.

## Consequences

* Ensures a uniform, high-contrast user experience across Dark, Light, and Cyber themes.
* Prevents accidental asset or policy deletion on production EDC Control Plane nodes.
* Guarantees reliable cross-platform rendering across Linux Docker containers and desktop browsers.
