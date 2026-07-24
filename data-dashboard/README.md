# Eclipse Dataspace Components (EDC) Data Dashboard

The **EDC Data Dashboard** serves as a lightweight, open-source management UI designed to interface with the EDC Management API. It acts as a 1:1 graphical wrapper over the Control Plane to manage data assets, usage policies, contract definitions, and transfers.

---

## 1. Prerequisites & Getting Started

Before logging into the UI, ensure your local configuration or multi-environment instances are declared properly.

* **Connector Config (`public/config/edc-connector-config.json`):** Holds array definitions for known EDC instances, including target Control Plane, Data Plane, Catalog, and Identity Hub URLs and default API keys.
* **Application Settings (`public/config/app-config.json`):** Sets UI parameters like theme, titles, `healthCheckIntervalSeconds` (default: 30s), and whether individual users can append local connectors (`enableUserConfig`).

> **Security Warning:** Enabling `enableUserConfig: true` stores custom authorization keys inside browser `localStorage`. Do not use `localStorage` key persistence in production environments without strict security controls.

---

## 2. Navigation & User Interface Map

The application follows a clean Go SSR (`html/template`) + Tailwind/daisyUI inspired design with primary navigation handled via the main sidebar:

```
+-------------------------------------------------------------------------------+
|  EDC Data Dashboard                   [ Connector Selection Dropdown ]  [Theme]
+-----------------+-------------------------------------------------------------+
| [home] Home     |  Catalog / Assets / Policies / Contracts / Transfers        |
| [assets] Assets |                                                             |
| [policy] Policy |  [ Search / Filter ]                     [ + Create / New ] |
| [contract] ...  |  +-------------------------------------------------------+  |
|                 |  | Main Data Table / Visual Management Card Grid         |  |
+-----------------+-------------------------------------------------------------+
```

### Top Bar / Header

* **Connector Picker:** Choose which active EDC connector (e.g., *Provider EDC* vs. *Consumer EDC*) to target.
* **Health Status Indicator:** Displays connection validity based on the configured polling interval (`healthCheckIntervalSeconds`, default: `30s`).
* **Theme Switcher:** Toggles between Dark, Light, and Cyber Custom themes.

---

## 3. Core Functional Workflows

The EDC core workflow follows four essential operations: **Asset Setup → Policy Definition → Contract Creation → Catalog & Transfer Management**.

### Module A: Assets Management (`/assets`)

Assets represent the data sources or endpoints you want to make shareable in your dataspace.

1. **Open Assets View:** Select **Assets** from the left sidebar and click **+ Create Asset**.
2. **Specify Metadata:** Enter general asset descriptors (Asset ID, Title, Version, Content Type, Description, Keywords).
3. **Configure Data Address:** Define source details (Type e.g., `HttpData`, `AmazonS3`, `AzureStorage`, Base URL, Proxy Path, Auth Key, Bucket, Region).
4. **Save Asset:** Submit form to dispatch an HTTP `POST /v3/assets` request directly to the connected EDC Control Plane.

---

### Module B: Policy Definitions (`/policies`)

Policies specify legal or technical constraints under which data can be consumed (e.g., time boundaries, regional restrictions, or role requirements).

1. Click **Policies** on sidebar &rarr; Select **+ Create Policy**.
2. Provide a **Policy ID**.
3. Choose or construct **Rules** (Permissions, Prohibitions, Obligations).
4. Add **Constraints** using standard ODRL operators (`eq`, `in`, `neq`), e.g. Left Operand: `spatial`, Operator: `eq`, Right Operand: `https://w3id.org/idsa/code/EU`.

---

### Module C: Contract Definitions (`/contract-definitions`)

Contract Definitions link your **Assets** with your **Policies** to expose them in public catalogs.

* **Access Policy:** Determines *who* is allowed to discover and see the asset in the catalog.
* **Contract Policy:** Determines *under what conditions* a consumer can negotiate an agreement for the asset.
* **Asset Selector:** Select specific asset IDs or apply criteria (e.g., `asset:prop:id = asset-01-sensor-data`).

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

1. **Browse/Fetch Catalog:** Enter Provider Counter-Party Address (e.g., `http://provider-connector:8282/api/v1/dsp`). The UI queries the remote catalog using the Dataspace Protocol (DSP).
2. **Negotiate Contract:** Select an available asset item from the grid, review associated policy rules, and click **Negotiate**. Track negotiation progress from `REQUESTED` to `AGREED` or `FINALIZED`.
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
