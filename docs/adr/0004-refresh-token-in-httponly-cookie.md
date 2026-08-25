# Refresh Token moves to an httpOnly cookie; Access Token stays bearer-in-body

Refresh Tokens were exchanged purely via JSON body (client resends them explicitly to `/auth/refresh` and `/auth/logout`), leaving them exposed to any script that can read the response — an XSS bug anywhere in the frontend could exfiltrate a 7-day credential. We're moving only the Refresh Token into an httpOnly, `SameSite=Strict` cookie scoped to `Path=/api/v1/auth`, because that's the long-lived, highest-value credential; the Access Token (15 min TTL) stays in the response body and `Authorization` header as before, since httpOnly cookies would force a bigger rework of the request path for comparatively little benefit.

Because the frontend and API are same-origin, standard CSRF defenses (SameSite alone) were judged insufficient against future changes to that assumption, so a custom `X-SMLIMS-CSRF` header is also required on cookie-authenticated auth requests — cheaper than a full double-submit CSRF token scheme and sufficient given the narrow (same-origin) deployment target.

`/auth/refresh` and `/auth/logout` read the Refresh Token from the cookie first, falling back to the JSON body if absent, since no non-browser client is confirmed to exist yet but one may show up (e.g. a mobile app) and cookies aren't available to it. `/auth/login` no longer echoes the Refresh Token in its JSON response for browser clients — cookie only.

Reuse Detection (a revoked Refresh Token presented again) revokes every Session the User holds, not just the affected Token Family, on the assumption that reuse means the token leaked and the whole account should be treated as compromised.

**Status**: accepted
