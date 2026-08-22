# Domain Glossary — Thanes LIMS Backend

## AI Chatbot (POC)

**Select AI** — Oracle Database 23ai feature (`DBMS_CLOUD_AI` package) that translates a natural-language question into SQL and executes it against tables in that same Oracle database, returning a narrated answer. Runs *inside* the Oracle DB, not as an external LLM API call from application code.

**AI Profile** — A named configuration stored in the Oracle DB (via `DBMS_CLOUD_AI.CREATE_PROFILE`) that tells Select AI which LLM provider to use (this project: **OCI Generative AI**), which tables/views are in scope, and what credentials to use for that provider.

**POC Oracle Instance** — A separate Oracle Autonomous Database (ADB), already provisioned, used *only* for this chatbot proof-of-concept. It is not the system of record — the SM-LIMS/Thanes LIMS backend's primary data store remains **Postgres** via GORM. The POC Oracle instance holds a synthetic, mirrored subset of the domain (Sample, TestResult, Inventory, PurchaseOrder) seeded independently — it is not synced from Postgres.

**Chatbot module** — New module in the existing hexagonal architecture (`internal/domain/chatbot`, etc.) exposing a single-turn `POST /chat`-style endpoint on the existing Go API, reusing the existing JWT auth. Calls into the POC Oracle instance via `godror` (Oracle wallet + Instant Client) to run Select AI queries and return the narrated answer.
