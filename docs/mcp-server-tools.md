# LIMS MCP server — tool catalog

This is the contract for the MCP (Model Context Protocol) server in `cmd/mcp-server`
(`internal/adapters/mcp`), built for the external NestJS + LangGraph.js chatbot service to
call instead of the old Oracle ADB mirror (see `docs/adr/0013-mcp-server-transport.md` and
`/Users/tng-mac-01/.claude/plans/ai-chatbot-groovy-spindle.md`). It reads Postgres directly
through this backend's existing Sample/TestResult/Inventory/PurchaseOrder repositories -
read-only, no new SQL.

## Connecting

- **Transport**: MCP Streamable HTTP, per the MCP spec.
- **URL**: `http://<host>:<MCP_SERVER_PORT>/` (default port `8090`; there is no separate
  path prefix, the handler is mounted at the process root).
- **Headers required on every request** (both mandatory, checked before any MCP method
  including `tools/list`):
  - `X-Service-Api-Key: <MCP_SERVICE_API_KEY>` — static service-to-service secret shared
    with the NestJS deployment out of band.
  - `Authorization: Bearer <jwt>` — the **same access token** the end user already
    authenticated with against this backend's own `/api/v1/auth/login`. Forward it
    unmodified; do not mint a separate service-level token.
- **RBAC**: every tool requires the caller's JWT `permissions` claim to include **both**
  `chatbot:view` (`rbac.ModuleChatbot`/`rbac.ActionView` - "may use the AI assistant at
  all") **and** the tool's own domain permission (see the table below) - a caller can only
  pull a domain's data through the chatbot if their Role could already see that data
  through the ordinary REST API. `chatbot:view` is granted to every Role today (Admin, Lab
  Manager, Scientist, QA, General) - see `docs/chatbot-acceptance-checklist.md`; the
  per-domain permissions follow each Role's existing grants (same `sample:view`/
  `testresult:view`/`inventory:view`/`purchaseorder:view` keys the regular endpoints already
  check - no new permission or migration was added for this). A caller missing either gets a
  tool-level error (`CallToolResult.IsError = true`), not a transport rejection - the
  missing/invalid service key or JWT itself is what returns HTTP 401.

## Tools

| Tool | Input | Output | Domain permission required (in addition to `chatbot:view`) | Scenario covered |
|---|---|---|---|---|
| `getSampleById` | `id: string` | one `Sample` object | `sample:view` | Sample detail lookup |
| `listSamplesByStatus` | `status: string`, `olderThanDays?: int` | `{ samples: Sample[], count }` | `sample:view` | Acceptance checklist scenario 1 (pending > 7 days) |
| `listSamplesByCustodianName` | `name: string` | `{ samples: Sample[], count }` | `sample:view` | Acceptance checklist scenario 5 |
| `searchTestResults` | `sampleId?: string`, `status?: string`, `flag?: string` | `{ results: TestResult[], count }` | `testresult:view` | Acceptance checklist scenario 2 (hi/lo flags) |
| `listInventoryLowStock` | *(none)* | `{ items: InventoryItem[], count }` | `inventory:view` | Acceptance checklist scenario 3 |
| `getInventoryItemById` | `id: string` | one `InventoryItem` (with `lots`) | `inventory:view` | Inventory detail lookup |
| `listPurchaseOrders` | `status?: string` | `{ orders: PurchaseOrder[], count }` | `purchaseorder:view` | Acceptance checklist scenario 4 |
| `listPurchaseOrdersByItem` | `inventoryItemId: string` | `{ orders: PurchaseOrder[], count }` | `purchaseorder:view` | Acceptance checklist scenario 6 |

All string enum fields below are spelled out because the LLM only ever sees these JSON
schema descriptions, never the Go domain types.

### `Sample`

```jsonc
{
  "id": "SMP-2569-00021",
  "name": "string",
  "type": "blood | urine | water | tissue | food | serum",
  "status": "pending | testing | completed | transferred",
  "custodianUserId": 7,
  "custodianName": "วิภา สายใจ",       // resolved from the user directory when possible
  "locationId": "string | omitted",
  "receivedAt": "RFC3339 timestamp",
  "pendingDays": 15,                    // only present when olderThanDays was set on listSamplesByStatus
  "barcodeId": "string | omitted",
  "description": "string | omitted"
}
```

- `status` meanings: `pending` = received, awaiting testing; `testing` = in progress;
  `completed` = testing finished, terminal (a sample is never reopened); `transferred` =
  moved to another department.

### `TestResult`

```jsonc
{
  "id": "TR-2569-00005",
  "sampleId": "SMP-2569-00003",
  "testName": "IgG",
  "analyst": "string",
  "result": "string",
  "flag": "hi | lo | ok",
  "refRange": "string",
  "status": "analyzing | pending_verification | approved"
}
```

- `flag`: `hi` = above reference range, `lo` = below reference range, `ok` = normal.
  `searchTestResults` filters on `flag` in the Go handler (not at the database layer) - see
  the "Design notes" section below.

### `InventoryItem`

```jsonc
{
  "id": "INV-0002",
  "name": "ถุงมือไนไตรไซส์ M",
  "category": "string",
  "quantity": 3,
  "unit": "string",
  "minQuantity": 15,
  "maxQuantity": 100,
  "belowMin": true,               // quantity <= minQuantity
  "defaultVendor": "string | omitted",
  "manufacturer": "string | omitted",
  "custodianUserId": 4,
  "earliestExpireDate": "RFC3339 date | omitted",
  "lotCount": 2,
  "lots": [                        // only populated by getInventoryItemById
    { "id": "LOT-1", "lotNo": "A100", "quantity": 3, "expireDate": "RFC3339 | omitted" }
  ]
}
```

### `PurchaseOrder`

```jsonc
{
  "id": "PO-2569-0012",
  "itemId": "INV-0002",
  "quantity": 10,
  "vendor": "string",
  "orderDate": "RFC3339 timestamp",
  "status": "pending_approval | sent_to_vendor | received | cancelled"
}
```

## Design notes / deviations from repository interfaces

These are deliberate choices to keep the MCP adapter thin and avoid widening core
repository interfaces for narrow, low-traffic chatbot queries:

- **`listSamplesByStatus`'s `olderThanDays`**: `SampleRepository.List` has no date filter.
  The handler calls `List` with just the status filter, then filters by `ReceivedAt` in Go.
  Sample volumes are lab-scale (not high-cardinality), so this is cheap and avoids adding a
  bespoke filter field to `sample.ListFilter` for one query shape.
- **`listSamplesByCustodianName`'s name resolution**: `ports/user.UserRepository` has no
  find-by-name method. The handler calls `Users.List()` and matches case-insensitively in
  Go, rather than adding a new repository method. The user table is lab staff, not
  customers - small and low-traffic - so this tradeoff was chosen over widening a core port.
- **`searchTestResults`'s `flag` filter**: `testresult.ListFilter` only supports
  `SampleID`/`Status`; `flag` is filtered client-side in the handler after a broader `List`.
- **`listPurchaseOrders`'s `status` filter** and **`listPurchaseOrdersByItem`**:
  `purchaseorder.Repository.List` takes no filter at all; both are filtered in the handler
  after a full `List` call. The `purchase_orders` table is small (tens of rows in the seed
  data), so this is not a performance concern.

None of these required changing any repository interface or its Postgres implementation.

## Error handling

- **Bad/missing required input** (empty `id`, invalid enum value): the tool call returns
  `CallToolResult{ IsError: true }` with a human-readable message in `Content`, so the
  calling LLM sees the failure and can self-correct (e.g. re-ask with a valid status).
- **Not found** (`getSampleById`, `getInventoryItemById`): same shape, `IsError: true`,
  `"sample \"X\" not found"` style message.
- **Missing `chatbot:view` permission**: same shape, `"permission denied: missing
  chatbot:view"`.
- **Missing the tool's domain permission** (e.g. `inventory:view` for `listInventoryLowStock`),
  even with `chatbot:view` present: same shape, `"permission denied: missing
  <module>:view"`.
- **Missing/invalid `X-Service-Api-Key` or `Authorization` header**: HTTP 401, before any
  MCP message is processed at all (transport-level, not a tool error).
- **Unexpected repository error**: returned as a Go error from the handler, which the SDK
  surfaces as an MCP protocol-level error (distinct from `IsError` tool-level failures).
