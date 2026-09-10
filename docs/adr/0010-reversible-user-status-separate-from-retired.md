# 0010 - Reversible user status (active/suspended) separate from Retired

## Status
Accepted (2026-09-10)

## Context
We needed a way for an Admin to stop a User from logging in without deleting
them — someone on extended leave, or a disciplinary pause. The codebase already
had **Retired** (`gorm.DeletedAt` soft-delete, wired on `users` but with no use
case) which hides the record, frees the unique email for reuse, and is treated
everywhere as irreversible.

## Decision
Add a distinct `users.status` column (`active` | `suspended`), NOT NULL
DEFAULT `'active'`, orthogonal to Retired and to Role. Suspending revokes
every Session immediately (reusing `RefreshTokenRepository.RevokeAllForUser`,
the same mechanism a Role change already uses) and is checked at both Login and
Refresh (fail-closed on Refresh). `reactivate` flips it back with no data loss.

We did **not** overload Retired for this: a Retired record is meant to be gone,
its email reusable and its listings hidden — none of which is true for a User
who is coming back. Nor did we model status as a third Role — Role drives the
JWT permission claim and the RBAC tables, and "suspended" is not a permission
set.

## Consequences
- The last-active-admin guard (previously `CountByRole`) now counts only
  `active` Admins, and covers three paths: role change, suspend, retire.
- Retiring a User is blocked while they are the Custodian of any non-Retired
  Sample or Inventory Item (checked via a new `CustodianChecker` port so
  `application/user` does not depend on the sample/inventory domains).
- Login can now fail *after* a correct password with a distinct "account
  suspended" error — acceptable, since the password check already passed.
