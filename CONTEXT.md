# Domain Glossary — Thanes LIMS Backend

## AI Chatbot (POC)

**Select AI** — Oracle Database 23ai feature (`DBMS_CLOUD_AI` package) that translates a natural-language question into SQL and executes it against tables in that same Oracle database, returning a narrated answer. Runs *inside* the Oracle DB, not as an external LLM API call from application code.

**AI Profile** — A named configuration stored in the Oracle DB (via `DBMS_CLOUD_AI.CREATE_PROFILE`) that tells Select AI which LLM provider to use (this project: **OCI Generative AI**), which tables/views are in scope, and what credentials to use for that provider.

**POC Oracle Instance** — A separate Oracle Autonomous Database (ADB), already provisioned, used *only* for this chatbot proof-of-concept. It is not the system of record — the SM-LIMS/Thanes LIMS backend's primary data store remains **Postgres** via GORM. The POC Oracle instance holds a synthetic, mirrored subset of the domain (Sample, TestResult, Inventory, PurchaseOrder) seeded independently — it is not synced from Postgres.

**Chatbot module** — New module in the existing hexagonal architecture (`internal/domain/chatbot`, etc.) exposing a single-turn `POST /chat`-style endpoint on the existing Go API, reusing the existing JWT auth. Calls into the POC Oracle instance via `godror` (Oracle wallet + Instant Client) to run Select AI queries and return the narrated answer.

## Storage Location

**Location** — A node in the physical storage hierarchy where samples are kept. Forms a tree: a Location may have a parent Location and child Locations. Every Location has a `Level Type` and a `Name`, and belongs to exactly one Cabinet's tree (its root ancestor).
_Avoid_: Place, position, spot (use Location for the entity; "leaf location" for the node a sample is actually stored at)

**Level Type** — The rung of the storage hierarchy a Location occupies: `cabinet` → `shelf` → `slot` → `sub_slot`, in that fixed order. A parent's Level Type must be the immediate predecessor of its child's — levels cannot be skipped.
_Avoid_: Level, type, tier

**Cabinet** — A root Location (`parent` is none, `level_type` is `cabinet`). Its Name is unique across the whole tree, not just among siblings.

**Leaf Location** — A Location with no children. Only a Leaf Location can be assigned to a Sample; any Level Type can be a leaf if the operator chooses not to subdivide it further (e.g. a Cabinet with no Shelves is itself a valid storage spot).

**Full Path** — The human-readable chain of a Location's ancestors down to itself (e.g. "Fridge-A / Shelf-2 / Slot-4"), derived on read from the tree — never stored.
