# Equipment and Inventory share one Location tree, separate from Sample's

`Location` (the self-referencing tree from ADR 0001) gains a `Kind` discriminator: `sample_storage` is the existing tree (`cabinet`→`shelf`→`slot`→`sub_slot`, leaf-only assignment, occupancy-checked). A second kind, `equipment_storage`, is a new tree with its own level types and no occupancy constraint, and it is shared by both Equipment and Inventory Item rather than split into a third `inventory_storage` kind — the lab's actual storage rooms hold instruments and reagents side by side, so one tree per physical space matches reality better than one tree per business entity. Which kind of thing occupies a given node is never a column on `Location` itself: it's derived from whichever FK (`EquipmentID` or `InventoryItemID`) points at it, so adding a third occupant type later needs only a new FK, not a schema change to `Location`.

## Status

accepted

## Considered Options

- **One Location kind for everything** (Sample, Equipment, Inventory all in the same tree) — rejected: Sample's leaf-only, occupancy-checked assignment rules don't apply to Equipment/Inventory, and forcing them into the same tree would mean bolting exceptions onto Sample's invariants.
- **Three separate kinds** (`sample_storage`, `equipment_storage`, `inventory_storage`) — rejected: Equipment and Inventory don't need different tree shapes from each other, and the physical storage areas overlap in practice, so splitting them would just require cross-referencing two trees for what is one room.
