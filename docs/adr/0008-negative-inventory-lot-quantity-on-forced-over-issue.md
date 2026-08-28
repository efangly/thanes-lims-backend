# Stock Issue allows a Lot's Quantity to go negative

Stock Issue always draws against one Lot the user explicitly picks (never an auto-FEFO selection). When the requested quantity exceeds that Lot's recorded `Quantity`, the system warns and shows the recorded remaining balance, then asks whether to continue from a second Lot. If the user instead confirms withdrawing the excess from the same Lot anyway, we record it: `Quantity` is allowed to go below zero rather than the request being rejected.

We chose this over hard-blocking the over-issue because the trigger for it is a real-world physical count disagreeing with the system's recorded count (e.g. someone already took stock without logging it) — blocking the withdrawal wouldn't make the physical shortage go away, it would just stop the log from reflecting what actually happened. A negative `Quantity` is therefore a deliberate signal of a count discrepancy to investigate, not a bug or a data integrity violation to prevent.

## Status

accepted

## Considered Options

- **Hard-block any withdrawal exceeding the Lot's recorded Quantity** — rejected: forces the user to either under-report the real withdrawal or go log a separate "adjustment" first, adding a step for something that will keep happening as long as physical counts drift from the system.
- **Auto-split the excess across other Lots (FEFO) instead of asking** — rejected earlier in the same discussion: Lot selection must stay an explicit user choice (see `[[inventory-lot-and-stock-issue]]` in CONTEXT.md), since knowing exactly which Lot was drawn from is the point of tracking Lots at all.
