# Data Dashboard Module

The **Data Dashboard** serves as the administrative control center for your Sovereign Dataspace Connector node.

## Core Responsibilities

*   **SSR Template Layouts**: Implements clean modular rendering via Go `html/template` and dark/light glassmorphic stylesheets.
*   **Asset Management GUI**: Registers, lists, and removes local DCAT-AP dataset catalog records.
*   **ODRL Policy Builder**: Interactive wizard allowing administrators to build compliant JSON policies.
*   **DSP Handshake Wizard**: Live visual progression execution of the contract negotiation and transfer mapping steps.
*   **Dynamic Node Switcher**: Re-targets client components to remote or adjacent nodes configured in `/config/edc-connector-config.json`.
*   **Telemetry Probes**: Periodically monitors local service status indicators.

## Architecture

*   `core/`: Model data structures, config parsers, and external client request handlers.
*   `ports/`: HTTP view handlers, API orchestration handlers, static style configurations, and template structures.
