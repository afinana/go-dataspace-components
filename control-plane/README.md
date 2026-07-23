# Control Plane Module

The **Control Plane** drives the protocol handshakes, contract negotiation state machines, and data transfer coordination.

## Core Responsibilities

*   **Contract Negotiation State Machine**: Implements the validated state-transition progression for W3C DSP negotiations (Requested $\rightarrow$ Agreed $\rightarrow$ Finalized $\rightarrow$ Terminated).
*   **Transfer Process Machine**: Tracks data stream lifecycles, issuing active signal flows to target Data Planes.
*   **ODRL Policy Evaluator**: Coordinates with the Identity Hub to match participant claims against ODRL contract requirements prior to signing agreements.
*   **Outbound Signalers**: Dispatches signaling requests to partner control planes over the network.

## Architecture

*   `domain/`: Decoupled transition guards, evaluation contexts, and negotiation state types.
*   `ports/`: HTTP handlers matching DSP specifications, management API integration, and SQL store implementations.
