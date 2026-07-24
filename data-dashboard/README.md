# Eclipse Dataspace Components (EDC) Data Dashboard

The **EDC Data Dashboard** serves as a lightweight, open-source management UI designed to interface with the EDC Management API. It acts as a 1:1 graphical wrapper over the Control Plane to manage data assets, usage policies, contract definitions, catalog queries, and transfers.

---

## 1. Prerequisites & Getting Started

Before logging into the UI, ensure your local configuration or multi-environment instances are declared properly.

* **Connector Config (`public/config/edc-connector-config.json`):** Holds array definitions for known EDC instances, including target Control Plane, Data Plane, Catalog, and Identity Hub URLs and default API keys.
* **Application Settings (`public/config/app-config.json`):** Sets UI parameters like theme, titles, `healthCheckIntervalSeconds` (default: 30s), and whether individual users can append local connectors (`enableUserConfig`).

> **Security Warning:** Enabling `enableUserConfig: true` stores custom authorization keys inside browser `localStorage`. Do not use `localStorage` key persistence in production environments without strict security controls.

---

## 2. Design System & User Interface Map

The application features a modern flat 2D design system built with custom CSS variables, vector SVG icons, and a glassmorphism sidebar structure:

```
+--------------------------------------------------------------------------------+
|  EDC Data Dashboard                   [ Connector Selection Dropdown ]  [Theme]|
+------------------+-------------------------------------------------------------+
| (icon) Home      |  Catalog / Assets / Policies / Contracts / Transfers         |
| (icon) Assets    |                                                             |
| (icon) Policies  |  [ Search / Filter ]                     [ + Create / New ]  |
| (icon) Contracts |  +-------------------------------------------------------+  |
| (icon) Catalog   |  | Responsive Flat Card Grid (minmax 340px)               |  |
| (icon) Transfers |  |  [ID Tooltip] [Badge] ... [View] [Edit] [Delete]       |  |
+------------------+-------------------------------------------------------------+
```

### Key UI Features

* **Flat 2D Aesthetic**: Clean, high-contrast layouts across Dark, Light, and Cyber Custom themes without distracting 3D shifts or heavy drop shadows.
* **Responsive Card Grids (`minmax(340px, 1fr)`)**: Cards scale dynamically with equal row heights and single-line code snippet truncation (`max-width: 200px` with hover tooltips).
* **Sidebar SVG Navigation**: Theme-matching vector icons for all navigation items that automatically invert to white when active.
* **Health Probes**: Real-time polling indicator for Control Plane (CP), Data Plane (DP), Identity Hub (ID), and Catalog (CAT).

---

## 3. Core Functional Workflows

The EDC core workflow follows four essential operations: **Asset Setup → Policy Definition → Contract Creation → Catalog & Transfer Management**.

### Module A: Assets Management (`/assets`)

Assets represent the data sources or endpoints you want to make shareable in your dataspace.

1. **Create Asset:** Click **+ Create Asset**, specify metadata (ID, Title, Version, Content Type, Keywords) and Data Address parameters (`HttpData`, `AmazonS3`, `AzureStorage`).
2. **View & Edit:** Inspect raw JSON specifications with **View** or update existing asset configurations with **Edit**.
3. **Delete Safeguards:** Clicking **Delete** opens an explicit confirmation modal before submitting a POST `/assets/delete` request to the EDC Control Plane.

---

### Module B: Policy Definitions (`/policies`)

Policies specify legal or technical constraints under which data can be consumed (e.g., time boundaries, regional restrictions, or role requirements).

1. **Create Policy:** Click **+ Create Policy**, enter Policy ID, and construct ODRL constraints (`eq`, `in`, `neq`).
2. **Lifecycle Modals:** Full **View**, **Edit**, and **Delete Confirmation** modal support.

---

### Module C: Contract Definitions (`/contract-definitions`)

Contract Definitions link your **Assets** with your **Policies** to expose them in public catalogs.

* **Access Policy:** Determines *who* is allowed to discover and see the asset in the catalog.
* **Contract Policy:** Determines *under what conditions* a consumer can negotiate an agreement for the asset.
* **Asset Selector:** Select specific asset IDs or apply criteria (e.g., `asset:prop:id = asset-01-sensor-data`).
* **Lifecycle Modals:** Includes **View**, **Edit**, and **Delete Confirmation** workflows.

---

### Module D: Federated Catalog & Data Transfer (`/catalog`, `/transfers`)

This view handles the consumer side of the data-sharing process.

```
+---------------------------------------------------------------------------+
|                          CONSUMER WORKFLOW                                |
|                                                                           |
| [ Query Catalog ] ---> [ Select Asset ] ---> [ Negotiate Contract ]       |
|                                                     |                     |
|                                                     v                     |
| [ Trigger Transfer ] <--- [ Contract Agreement ] <---+                    |
+---------------------------------------------------------------------------+
```

1. **Browse/Fetch Catalog:** Enter Provider Counter-Party Address (e.g., `http://provider-connector:8282/api/v1/dsp`). The UI queries the remote catalog using the Dataspace Protocol (DSP) with automatic fallback for empty dataset slices.
2. **Negotiate Contract:** Select an available dataset card, click **Negotiate**, and track status via a visual progress stepper (Request &rarr; Verify &rarr; Agree).
3. **Initiate Transfer:** Go to **Transfers & Agreements**, select an active agreement, click **Initiate Transfer**, specify destination parameters (HTTP pull stream or S3/Azure Storage push), and monitor live status in Transfer History.

---

## 4. Key Management API Statuses

| State | Context | Meaning |
| --- | --- | --- |
| **INITIALIZED** | Negotiation / Transfer | Protocol handshake initiated with counter-party connector. |
| **REQUESTED** | Contract Negotiation | Offer sent to provider control plane for verification. |
| **AGREED** | Contract Agreement | Both parties accepted terms; legally binding agreement ID issued. |
| **STARTED** | Data Transfer | Data Plane actively routing packets between endpoints. |
| **TERMINATED** | Negotiation / Transfer | Transaction canceled due to policy violation or network failure. |

---

## 5. Verification & Testing

Run unit tests for all core models and HTTP handlers:
```bash
go build ./... && go test ./...
```
