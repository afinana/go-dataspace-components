# Identity Hub Module

The **Identity Hub** coordinates decentralized identity operations, verifiable credentials (VCs), and standard cryptographic proof signatures for the Sovereign Dataspace Connector node.

## Core Responsibilities

*   **Decentralized Identifiers (DIDs)**: Implements compliant resolution and validation schemas for `did:web` structures.
*   **Security Token Service (STS)**: Generates self-issued JSON Web Tokens (JWT) and presentation proofs matching requested criteria.
*   **Verifiable Credentials Storage**: Holds issued identity assertions (e.g. membership credentials, geographic region identifiers) in a secure database structure.
*   **Presentation Query API**: Coordinates validation of presented assertions from external partners resolving identity queries.

## Architecture

Following the hexagonal design limits, the module is decoupled into:
*   `domain/`: Defines DID document structures, VC claims models, and proof verification rules.
*   `ports/`: Contains HTTP signaling routes, JWT token encoders, and PostgreSQL verification storage.
