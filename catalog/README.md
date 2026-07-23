# Catalog Module

The **Catalog** component administers DCAT-AP dataset metadata, ODRL usage policy constraints, and asset discovery operations.

## Core Responsibilities

*   **W3C DCAT-AP Model Compliance**: Maps structural attributes of local files or database assets to valid metadata specifications.
*   **Asset Registry**: Exposes local datasets available for external querying and handles registration actions.
*   **Federated Search**: Allows querying remote nodes in the dataspace network to fetch datasets and construct unified catalogs.
*   **ODRL Policy Rules**: Manages target access permission mappings, constraint predicates (geographic, temporal), and prohibition terms.

## Architecture

*   `domain/`: Standard catalog domain classes, distributions, and dataset validation.
*   `ports/`: HTTP handlers matching DSP standards, SQL catalog loaders, and serialization layers.
