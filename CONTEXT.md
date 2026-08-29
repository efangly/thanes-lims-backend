# Domain Glossary — Thanes LIMS Backend

## AI Chatbot (POC)

**Select AI** — Oracle Database 23ai feature (`DBMS_CLOUD_AI` package) that translates a natural-language question into SQL and executes it against tables in that same Oracle database, returning a narrated answer. Runs *inside* the Oracle DB, not as an external LLM API call from application code.

**AI Profile** — A named configuration stored in the Oracle DB (via `DBMS_CLOUD_AI.CREATE_PROFILE`) that tells Select AI which LLM provider to use (this project: **OCI Generative AI**), which tables/views are in scope, and what credentials to use for that provider.

**POC Oracle Instance** — A separate Oracle Autonomous Database (ADB), already provisioned, used *only* for this chatbot proof-of-concept. It is not the system of record — the SM-LIMS/Thanes LIMS backend's primary data store remains **Postgres** via GORM. The POC Oracle instance holds a synthetic, mirrored subset of the domain (Sample, TestResult, Inventory, PurchaseOrder) seeded independently — it is not synced from Postgres.

**Chatbot module** — New module in the existing hexagonal architecture (`internal/domain/chatbot`, etc.) exposing a single-turn `POST /chat`-style endpoint on the existing Go API, reusing the existing JWT auth. Calls into the POC Oracle instance via `godror` (Oracle wallet + Instant Client) to run Select AI queries and return the narrated answer.

## Sample Registry

**Barcode ID** — An optional code on a Sample (`BarcodeID`), either typed in by the user or system-generated (`SMP-BC-{seq5}`, its own `id_sequences` scope), separate from the Sample's `ID` (which stays the auto-generated sequence, e.g. `SMP-2569-00001`) and unique across non-Retired Samples when set. Used to scan-filter the registry (`GET /samples?barcode_id=` exact match) and to print a physical sticker (`GET /samples/{id}/sticker?template=&symbology=`). The backend renders the sticker itself (reusing `internal/adapters/pdf`) rather than handing the frontend a raw code to lay out — different use cases already print different physical label sizes, so the sticker renderer supports more than one label template/size (`cap` 9.5×6.4mm, `stem` 20.5×6.5mm, `small` 40×20mm, `medium` 60×30mm) and both `code128` and `qr` symbologies, not one fixed layout. If a Sample has no Barcode ID, its sticker encodes the Sample `ID` instead so it is always scannable.
_Avoid_: Sample ID, code (Sample already has an `ID` — Barcode ID is an additional identifier for physical/scan use, not a replacement for it)

**Custodian** — The User responsible for a Sample, piece of Equipment, or Inventory Item, referenced by FK to User (not free text). Chosen from a dropdown sourced from the User list.
_Avoid_: Owner, responsible person, assignee

## Storage Location

**Location** — A node in the physical storage hierarchy where samples are kept. Forms a tree: a Location may have a parent Location and child Locations. Every Location has a `Level Type` and a `Name`, and belongs to exactly one Cabinet's tree (its root ancestor).
_Avoid_: Place, position, spot (use Location for the entity; "leaf location" for the node a sample is actually stored at)

**Level Type** — The rung of the storage hierarchy a Location occupies: `cabinet` → `shelf` → `slot` → `sub_slot`, in that fixed order. A parent's Level Type must be the immediate predecessor of its child's — levels cannot be skipped.
_Avoid_: Level, type, tier

**Cabinet** — A root Location (`parent` is none, `level_type` is `cabinet`). Its Name is unique across the whole tree, not just among siblings.

**Leaf Location** — A Location with no children **and not a Box**. Only a Leaf Location can be assigned a Sample directly, one at a time; any Level Type can be a leaf if the operator chooses not to subdivide it further (e.g. a Cabinet with no Shelves is itself a valid storage spot).

**Box** — A Location with `level_type = 'box'` that holds many Samples at once, each in an addressable Cell, instead of one Sample outright (see `docs/adr/0009`). It carries a **Grid** (`rows`, `cols` columns on `locations`, non-null only for boxes; `rows` ≤ 26, `cols` ≤ 99), hangs off a Shelf, Slot, or Sub-slot (not a fixed depth), never has child Locations, and carries a Location Barcode like any node. `level_type` is therefore no longer a pure depth indicator: `box` is a terminal marker at depth 2, 3, or 4.
_Avoid_: Compartment, container, rack, plate

**Cell** — One position inside a Box's Grid, named microplate-style (row letter `A`–`Z`, then column number: `A1`, `H12`; `A1` top-left). A Cell is **not** a Location node — it is the `position` string stored on the Sample (`samples.position`, null for every Sample not in a box). Occupancy for a box is "one active Sample per `(location_id, position)`"; a Sample in a box always has a position. Enforced by the app layer plus a partial unique index `uq_samples_box_cell_active`.
_Avoid_: Well, slot (Slot is a Level Type), position (the field name; Cell is the thing)

**Move-within-box** — Rearranging Samples among the Cells of one Box via `POST /locations/:boxId/moves` with `[{sample_id, position}]`, applied as one transaction so a multi-select drag or a two-Cell swap is atomic; a resulting position clash fails the whole batch with 409. Moving a Sample in from elsewhere is ordinary Put-away (`PATCH /samples/:id/location` with `position`), not a grid move. Boxes only grow — enlarge with `PATCH /locations/:id/grid`; shrinking is not supported (make a new box and move).

**Full Path** — The human-readable chain of a Location's ancestors down to itself (e.g. "Fridge-A / Shelf-2 / Slot-4"), derived on read from the tree — never stored.

## Vendors

**Vendor** — Master-data entity for a supplier: `Name` (unique), `ContactName`, `ContactPhone`, `ContactEmail`, `Address` (optional). Referenced by FK (`VendorID`) from Equipment, Inventory Item, and Purchase Order — never duplicated as free text. Can be created either from a dedicated Vendor master page or inline while filling in an Equipment/Inventory/Purchase Order form (a quick-add that still writes to the same master record, not a local copy).
_Avoid_: Supplier, manufacturer (Manufacturer is a separate, plain descriptive field on Equipment/Inventory Item — it does not carry contact details and is not master data)

## Storage Location (generalized)

**Location Kind** — Discriminator on `Location` distinguishing which subsystem a Location tree belongs to. `sample_storage` is the original tree (`cabinet`→`shelf`→`slot`→`sub_slot`, leaf-only assignment, occupancy-checked — see below). `equipment_storage` is a second tree (`building`→`room`→`zone`→`cabinet`→`shelf`, no occupancy constraint) shared by both Equipment and Inventory Item — the same physical storage room legitimately holds both, so they use one tree rather than two. `cabinet` and `shelf` are reused as level names in both trees at different depths, always disambiguated by Kind. Which kind of thing sits at a given node is never a field on the node itself: it falls out of whichever FK (`EquipmentID` or `InventoryItemID`) references that Location.
_Avoid_: Location type (Level Type already names the rung within one tree; Kind names which tree)

**Location Barcode** — A `BarcodeCode` on any `Location` node (format `LOC-BC-{seq5}`, its own `id_sequences` counter), auto-generated when the node is created and unique across non-Retired Locations, letting a scan resolve directly to that Location (`GET /locations/by-barcode/{code}`) — e.g. scanning a cabinet as the destination when moving a Sample, instead of navigating the tree by hand. Every node that existed before Phase 2 was backfilled with a code by migration `000027`.

## Equipment

**Equipment Asset Fields** (Phase 5) — Beyond `Name` and `TypeCode`, an Equipment carries: `SerialNumber` (unique across active Equipment when set), `Category` (a human-facing classification kept *separate* from `TypeCode` — `TypeCode` stays a short code and drives the `EQ-{TYPE}-{seq}` ID sequence), `Manufacturer` (plain descriptive string — *not* the Vendor), `Model`, `InstallationDate`, `VendorID` (FK → Vendor, optional), and `LocationID` (FK → a `equipment_storage` Location, optional — ADR 0007, Kind is validated on write). All editable after creation via `PATCH /equipment/{id}` (partial update; calibration dates move only through `/calibration`).
_Avoid_: putting supplier contact details in `Manufacturer` (that is what the Vendor master record is for)

**Equipment Document link** (Phase 5) — A `Document` may optionally link to one Equipment via a nullable `Document.EquipmentID` FK (single-owner, not a join table — a warranty/manual belongs to exactly one machine). Set at upload time (`equipment_id` form field); filter with `GET /documents?equipment_id={id}`. New document `Type`: `warranty`.

**Calibration Schedule** (Phase 6) — A record on one Equipment: a free-text `Label` (e.g. "สอบเทียบภายใน", "สอบเทียบภายนอก"), a `NextDueDate`, and an optional `IntervalMonths`. An Equipment may have several Schedules at once, added manually via a "+" control — each tracked independently. When `IntervalMonths` is set, logging a `CalibrationEvent` whose `CalibrationType` matches the Schedule's `Label` (case-insensitive, trimmed) auto-advances `NextDueDate` by that interval (counted from the later of the current due date and the calibration date, so a late calibration doesn't leave the schedule in the past); when left unset, the user sets the next `NextDueDate` by hand each time. CRUD under `/equipment/{id}/calibration-schedules[/{scheduleId}]` (list/create need Equipment `view`/`edit`; update/delete need `edit`). Delete is a hard delete — a schedule is a live commitment, not audited history. Schedules are **not** a `Module` — they audit under Equipment.
_Avoid_: Due date (ambiguous now that one Equipment can carry several — Due date belongs to a specific Calibration Schedule)

**Calibration Event measurements** (Phase 6) — Beyond dates, a `CalibrationEvent` records `CalibrationType` (free text; drives Schedule auto-advance), `CalibrateValue` and `AcceptanceValue` (free text so units/tolerances like "±0.1 g" survive verbatim), and `Result` (`pass` / `fail` / empty). Still append-only, still logged via `PATCH /equipment/{id}/calibration` (Equipment `approve`).

**Calibration Results page** (Phase 6, requirement 2.2.1) — `GET /calibration-results` (Equipment `view`) is a flat, cross-Equipment list of every logged `CalibrationEvent`, newest first, with a search bar: `q` (ILIKE over equipment id/name, performed_by, calibration_type, notes), `equipment_id`, `result`, `from`/`to` (RFC3339, filters `CalibratedAt`).

**Calibration Certificate link** (Phase 6) — A `Document` may optionally link to one `CalibrationEvent` via a nullable `Document.CalibrationEventID` FK (single-owner, same pattern as the Phase 5 `EquipmentID` link). Set at upload time (`calibration_event_id` form field); filter with `GET /documents?calibration_event_id={id}`. New document `Type`: `certificate`.

## Inventory

**Inventory Item Asset Fields** (Phase 7) — Beyond `Name`, `Category`, `Unit`, `Min`/`Max` and the derived `Quantity`, an InventoryItem carries: `CustodianUserID` (FK → users, **required** — the User responsible for the item, see "Custodian"; same shape as `Sample.CustodianUserID` from Phase 3), `Manufacturer` (plain descriptive string — *not* the Vendor), `VendorID` (FK → Vendor, optional), and `LocationID` (FK → an `equipment_storage` Location, optional — the same shared tree Equipment uses, ADR 0007, Kind validated on write). The pre-existing free-text `DefaultVendor` is kept as-is (it drives auto-reorder) *alongside* the new `VendorID`. All editable after creation via `PATCH /inventory/{id}` (partial update; `Quantity` still moves only through `/quantity`, `DefaultVendor` only through `/default-vendor`).
_Avoid_: putting supplier contact details in `Manufacturer` (that is what the Vendor master record is for); treating `DefaultVendor` and `VendorID` as the same thing

**Inventory Lot** — One batch of stock for an Inventory Item: `LotNo`, `ExpireDate` (optional), `Quantity`. An Item's on-hand quantity is the sum of its Lots' Quantities, not a field set directly.
_Implemented (Phase 8, 2026-08-27)_: `internal/domain/inventory/inventory_lot.go`; table `inventory_lots` + migration `000035` which **drops `inventory_items.quantity`** (now derived — `postgres/inventory.Repository` sums lots on `FindByID`/`List`; down migration re-adds and backfills the column). `(item_id, lot_no)` unique — receiving a duplicate lot number tops up that lot. Receiving: `POST /inventory/{id}/receive` (`lot_no` required, `expire_date` optional, `quantity` > 0) via `ReceiveStockUseCase`; `GET /inventory/{id}/lots` lists an item's lots (`ListLotsUseCase`). **Removed**: `PATCH /inventory/{id}/quantity`, `UpdateQuantityUseCase`, `Repository.UpdateQuantity`, and the `quantity` field on create-item — stock only enters through lots. `PurchaseOrder` receipt now books an `InventoryLot` instead of bumping quantity: `MarkReceivedUseCase` takes `MarkReceivedInput{ID, LotNo, ExpireDate}` and `PATCH /purchase-orders/{id}/receive` now requires a `lot_no` body. New lot id `LOT-{seq5}` (idgen scope `inventory_lot`). Seed: opening stock enters as an `OPENING` lot per item.

**Stock Issue** — A withdrawal recorded against one specific Inventory Lot, chosen explicitly by the user (never auto-selected/FEFO). If the requested quantity exceeds that Lot's recorded Quantity, the system warns, shows the Lot's remaining balance, and asks whether to continue the withdrawal from a second Lot. If the user chooses to withdraw the excess from the same Lot anyway (e.g. a physical count disagrees with the system), the Lot's Quantity is allowed to go negative — a negative Quantity is the signal of a count discrepancy, not a validation error.
_Avoid_: Auto-allocation, FEFO pick (the system may suggest, but Lot selection is always a manual, explicit user choice)
_Implemented (Phase 9, 2026-08-27)_: `IssueStockUseCase` (`internal/application/inventory/issue_stock.go`) → `POST /inventory/{id}/issue` (permission `inventory:edit`). Body `{lines: [{lot_id, quantity>0}], force}` — one line per Lot, multi-line for a split withdrawal. Validates ≥1 line, each Lot exists and belongs to the item, no duplicate `lot_id`. If a line's quantity exceeds its Lot's balance and `force` is not set, **nothing is applied** and the response is `{applied: false, shortfalls: [{lot_id, lot_no, requested, available}], item, lots}` (real balances — the caller then adds another lot line or re-submits with `force: true`). With `force: true` every line is drawn down and the Lot Quantity may go negative (ADR 0008). Reuses `LotRepository.FindByID` + `UpdateQuantity`; **no ledger table or migration** — the audit middleware records the item field-diff when applied. Seed unchanged.

## Access Control

**Role** — A named set of Permissions assigned to a User (exactly one Role per User): Admin, Lab Manager, QA, Scientist, or General. Determines what the User is allowed to do; carries no other meaning (not a job title or org-chart position).
_Avoid_: Group, level, tier

**Permission** — A grant of one Action on one Module to a Role (e.g. "Scientist may edit Sample"). Permissions are never assigned to a User directly, only via their Role.
_Avoid_: Right, scope, claim

**Action** — The verb half of a Permission: `view`, `create`, `edit`, `delete`, or `approve`. The same five Actions apply uniformly across every Module, even where one doesn't obviously make sense yet (e.g. `approve` on Document) — an unused Action is simply never granted to any Role, rather than the Module having a bespoke Action set. One exception: `export` exists only on the Audit Module, distinct from `view` — a Role can browse the Audit Trail (`view`) without being able to pull its PDF compliance export (`export`).

**Module** — The noun half of a Permission and of an Audit Entry's Resource: the business area being acted on (Sample, Location, Equipment, Inventory Item, Purchase Order, Document, Test Result, User, Vendor, Environment, Notification, Audit). Corresponds to one hexagonal-architecture module in the codebase.

**Approve** — The Action that marks a record as reviewed and accepted by someone other than its author (e.g. QA approving a Test Result, or a Purchase Order). Distinct from `edit`: approving does not imply the ability to change the record's content.

## Audit Trail

**Audit Entry** — An automatically-recorded fact that a specific Actor performed a specific Action on a specific Module (and, for most Modules, a specific record) at a specific time. Written for every write, never edited or deleted after the fact.
_Avoid_: Audit log line, activity record

**Actor** — The authenticated User whose credentials were used to perform the Action an Audit Entry records. Every Audit Entry has exactly one Actor.

**Change Set** — The list of a record's fields whose values differed before and after a write, each as an old→new pair, attached to an Audit Entry. Only fields that actually changed appear — an Audit Entry for a write that touched one field carries a Change Set of one.
_Avoid_: Diff, delta

**Chain of Custody (CoC) Step** — A single handoff event in a Sample's custody history (who had it, when, what happened). Audited as its own Module, separate from the Sample it belongs to, because custody handoffs are themselves a compliance record independent of the Sample's other fields.

**Retired** — A record that has been deleted by a User but is kept in storage (never physically removed) so its Audit Trail and any references to it stay intact. Applies to every audited Module. A Retired record does not appear in normal listings and does not block reuse of values that were unique to it (e.g. a Retired User's email can be reused by a new User).
_Avoid_: Soft-deleted, archived

## Authentication & Sessions

**Session** — One continuous login on one device/browser, from initial Login until it is revoked or reaches its Absolute Session Lifetime. Represented in the system as a Token Family. A User may hold several Sessions at once (e.g. lab workstation and phone).
_Avoid_: Login, device (a Session is the login itself, not the hardware)

**Token Family** — The chain of Refresh Tokens produced by repeated Rolling Refresh of a single Session: each rotation revokes the presented token and issues its successor within the same family. Identified by a family id shared across the chain, with a `family_created_at` fixed at the first Login — used to enforce the Absolute Session Lifetime.
_Avoid_: Chain, token lineage

**Rolling Refresh** — The rule that using a valid Refresh Token immediately revokes it and issues a new one (same Token Family) with a fresh 7-day expiry — so an active Session never expires from inactivity alone, only by hitting the Absolute Session Lifetime or being revoked.

**Absolute Session Lifetime** — The hard cap of 30 days from a Token Family's `family_created_at`, after which Rolling Refresh stops renewing it regardless of activity and the User must Login again. Exists so no Session can persist forever purely by staying active.

**Reuse Detection** — The check that a Refresh Token already revoked by Rolling Refresh is being presented again — the signal that a Refresh Token has leaked. Triggers revocation of every Session the User holds, not just the offending Token Family, since a leaked token implies the whole account may be compromised.

## Caching

**Cache** — A Redis-backed, key-value read-through layer sitting in front of Postgres, accessed only through the `internal/ports/cache` port. Postgres remains the source of truth for every cached value; the Cache exists purely to reduce read load and is never the only place a value lives.
_Avoid_: Store, session store (Redis here is a Cache, not a data store of record)

**Refresh Token Cache Entry** — The cached copy of a Refresh Token row (family id, expiry, revoked flag, family_created_at), keyed by token hash, kept in sync with Postgres on every Rolling Refresh and revocation. Read fail-closed: if the Cache is unreachable, the request is rejected rather than risking a stale "not revoked" answer.
_Avoid_: Token blacklist (the cache holds live rows, not just a revoked-token denylist)

**Full Path Cache Entry** — The cached rendering of a Location's Full Path, keyed by Location id, expiring after a fixed TTL rather than being invalidated on Location changes. Read fail-open: if the Cache is unreachable or the entry has expired, Full Path is recomputed from Postgres directly.

**Refresh Cookie** — The httpOnly cookie that carries the current Refresh Token from browser to backend, scoped to the auth endpoints only. Distinct from the Access Token, which continues to travel as a bearer value in the response body / `Authorization` header, never as a cookie.
