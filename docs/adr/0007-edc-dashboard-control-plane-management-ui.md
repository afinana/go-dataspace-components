# ADR-0007: EDC Data Dashboard Control Plane Wrapper & Workflow Management

## Status

Accepted

## Context

The Sovereign Dataspace Connector requires a standard administrative management UI to manage the lifecycle of data assets, usage policies, contract publishing definitions, federated catalog discovery, and data plane transfers. The UI must follow W3C Dataspace Protocol (DSP) and Eclipse Dataspace Components (EDC) Management API conventions while maintaining a lightweight footprint.

## Decision

We establish the **EDC Data Dashboard** as a 1:1 graphical management wrapper over the EDC Control Plane API with the following design patterns:

1. **Prerequisite Configuration Contract**:
   - `public/config/edc-connector-config.json` stores array definitions for connector instances.
   - `public/config/app-config.json` controls theme defaults, title parameters, polling intervals (`healthCheckIntervalSeconds`), and `enableUserConfig`.
2. **Standard State Machine Visualization**:
   - UI status badges strictly mirror Management API states: `INITIALIZED`, `REQUESTED`, `AGREED`, `STARTED`, and `TERMINATED`.
3. **Modal-Driven Workflow Operations**:
   - **Asset Setup**: Captures general metadata and Data Plane source address details (HttpData, AmazonS3, AzureStorage).
   - **Policy Definition**: Constructs ODRL rule sets (Permissions, Prohibitions, Obligations) and constraints (`eq`, `in`, `neq`).
   - **Contract Creation**: Links Access Policies, Contract Policies, and Asset Selector filters (`asset:prop:id = asset-01-sensor-data`).
   - **Federated Catalog & Data Transfer**: Queries DSP provider endpoints, tracks contract negotiations from `REQUESTED` to `AGREED`, and executes transfer streams with HTTP Pull or S3 Push destinations.

## Consequences

* Provides complete administrative control over local and remote connector nodes without external dependencies.
* Enables clear diagnostic tracing of negotiation and transfer states across dataspace participants.
* Guarantees fast server-side rendering while delivering interactive client-side wizards.
