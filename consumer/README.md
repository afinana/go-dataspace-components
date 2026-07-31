# Go Dataspace Components - Consumer Module

This directory contains the consumer-specific modules for the Go Dataspace Connector.
The consumer module is responsible for orchestrating the client-side flows of standard Eclipse Dataspace Components protocols (DSP 2025-1).

## Architecture

Following the hexagonal architecture pattern, the consumer module is split into ports and adapters:
- **ports**: Interfaces, DB repositories (using PostgreSQL), HTTP handlers for Management API and DSP Callbacks, and DSP client.
- **cmd/consumer**: The main entry point to run the consumer service.

## Core Features
- DSP Client to send protocol messages to Provider Connectors (Catalog Request, Contract Request, Transfer Request).
- Callback Handlers to receive DSP messages from Provider Connectors (Offer, Agreement, Events, Termination).
- Management API for Consumer Operator to initiate and track negotiations and transfers.

## Configuration
Requires the following environment variables:
- `PROVIDER_DSP_URL`: URL of the provider's DSP protocol endpoint.
- `CONSUMER_CALLBACK_URL`: URL where the consumer receives DSP callbacks.
- `DATABASE_URL`: Connection string to PostgreSQL.
