# Postman Collection

`docs/postman_collection.json` is generated from `docs/swagger.json` via `openapi-to-postmanv2`
(regenerate after `make swagger`: `npx openapi-to-postmanv2 -s docs/swagger.json -o docs/postman_collection.json -p`,
then reapply the auth/variable tweaks below if you hand-edit them again — or just re-run the
one-off postprocessing script from this session if you have it).

91 requests across 16 folders (one per module: audit, auth, samples, environment, partner-devices, ...).

## Setup

1. Import `docs/postman_collection.json` into Postman.
2. Import an environment — `docs/postman_environment.local.json` (`http://localhost:8080/api/v1`)
   or `docs/postman_environment.production.json` (`https://lims.siamatic.work/api/v1`) — and select it.
3. Run **auth › login** with a real email/password (see `SEED_CREDENTIALS.md` for the local seed
   users). Its test script auto-saves `data.access_token` into the `accessToken` collection
   variable — every other request already sends it as `Authorization: Bearer {{accessToken}}`.
4. When the access token expires (~15m), run **auth › refresh** (needs the httpOnly refresh
   cookie Postman received at login, so keep cookie jar enabled) — it auto-saves the new token
   the same way.

Endpoints the swagger docs didn't mark `@Security BearerAuth` (login, refresh, health) have no
auth block, matching what the API actually requires.
