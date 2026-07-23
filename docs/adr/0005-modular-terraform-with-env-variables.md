# ADR-0005: Modular Terraform Architecture and Environment Configuration

## Status

Accepted

## Context

Initially, the Terraform files for the Sovereign Dataspace Connector were defined in a monolithic `main.tf` file. As the number of microservices (Identity Hub, Control Plane, Data Plane, Data Dashboard, and PostgreSQL) grew, managing resource definitions, variables, and cross-service configurations in a single file became complex and difficult to maintain. Additionally, managing deployment configurations across different environments (development, staging, production) required a standard mechanism for defining and loading environment variables.

## Decision

We will refactor the monolithic Terraform configuration into separate, reusable modules under `terraform/modules/` (e.g., `postgres`, `identity-hub`, `control-plane`, `data-plane`, `data-dashboard`).

To manage environment variables, we will adopt a standard `.env` and `env.example` setup mapping variables directly to Terraform's environment inputs (`TF_VAR_*`).

## Consequences

*   **Modular Architecture**: Each microservice configuration is self-contained in its own module directory, making maintenance and updates much cleaner.
*   **Variable Isolation**: Modules explicitly define inputs and outputs, eliminating side-effects and providing strong interfaces.
*   **Environment Management**: Configurations can be quickly adapted for multiple environments by copying `env.example` to `.env` and sourcing it prior to running `terraform` commands.
*   **Clean Outputs**: Root configuration leverages module outputs to compute high-level endpoint mappings (e.g., NodePort dashboard endpoints).
