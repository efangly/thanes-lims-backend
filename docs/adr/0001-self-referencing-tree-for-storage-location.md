# Self-referencing tree for storage Locations, not fixed columns

`Sample.Location` was a freeform string (`"Fridge-A / R2-04"`). We're replacing it with a `locations` table modeling Cabinet → Shelf → Slot → Sub-slot as a self-referencing tree (`parent_id`, `level_type` enum, `name`), instead of the more obvious fixed-column approach (`cabinet_id`, `shelf_id`, `slot_id`, `sub_slot_id` on `samples`, or four separate typed tables).

We picked the tree because the lab wants to define an arbitrary number of shelves/slots/sub-slots per cabinet at data-entry time (auto-generated via a "generate N children" operation), not a schema with a fixed fan-out. A fixed-column design would need a schema migration every time the physical storage hierarchy changed shape; the tree absorbs that as data. The cost: hierarchy-order validation (cabinet→shelf→slot→sub_slot, no skipping) and Full Path resolution both move from the database's type system into application code and recursive queries.

## Status

accepted

## Considered Options

- **Four fixed FK columns on `samples`** (`cabinet_id`, `shelf_id`, `slot_id`, `sub_slot_id`) — simplest queries, but can't represent "cabinet with no shelves" as a storable location without nullable chains, and locks the depth at 4 forever.
- **Four separate typed tables** (`cabinets`, `shelves`, `slots`, `sub_slots`) with FK chains — type-safe per level, but adding a 5th level later means a new table + new FK column everywhere, and the "generate children" operation needs a different implementation per level.
- **Self-referencing tree (chosen)** — one table, one recursive "generate children" operation reusable at every level, depth changes are a data operation not a migration. Trades DB-level type safety for `level_type` order validated in the application layer.
