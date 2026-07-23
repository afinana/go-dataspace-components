# Terraform Deployment for k3s

This directory contains the Terraform configuration to deploy the Sovereign Dataspace Connector on a **k3s** Kubernetes cluster.

The configuration has been refactored into reusable modules for each microservice under `modules/`.

## Architecture & Modules

*   **postgres**: Database initialization and credentials configuration.
*   **identity-hub**: W3C DID, Identity Hub, and credential verifier microservice.
*   **control-plane**: Control Plane coordinator, contract negotiator, and catalog manager.
*   **data-plane**: Data Plane asset emitter and egress/ingress routing service.
*   **data-dashboard**: Sovereign Data Space administrative dashboard (exposed as NodePort `30084`).

## Prerequisites

*   A running **k3s** cluster. The configuration pulls cluster credentials from `~/.kube/config` by default.
*   The deployment images (`dataspace-identity-hub:latest`, `dataspace-control-plane:latest`, etc.) must be imported or built in k3s. If running locally, import built Docker containers via:
    ```bash
    sudo k3s ctr images import <my-image-archive>.tar
    ```

## Usage

1.  **Configure environment variables**:
    Copy `env.example` to `.env` and fill in your custom values:
    ```bash
    cp env.example .env
    ```

2.  **Load the environment**:
    Prior to running Terraform, load variables from your `.env` file:
    ```bash
    export $(grep -v '^#' .env | xargs)
    ```

3.  **Initialize Terraform**:
    ```bash
    terraform init
    ```

4.  **Plan the deployment**:
    ```bash
    terraform plan
    ```

5.  **Apply the configuration**:
    ```bash
    terraform apply
    ```

6.  **Access the Sovereign Data Dashboard**:
    `http://<k3s-node-ip>:30084`

