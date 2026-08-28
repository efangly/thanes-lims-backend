# รายงานผลการรีวิวโปรเจค (Thanes LIMS Backend)

รีวิวทั้งโปรเจค เน้นส่วน auth / refresh token / Redis caching layer ที่เพิ่งเพิ่มเข้ามา
(commit `5619a53`, `38f901a`, `0f236fd`) รวมถึง middleware, config และ data layer

- `go build ./...` ผ่าน
- `go vet ./...` ไม่มี warning
- `go test ./...` ผ่าน (integration test ที่ต้องใช้ DB ถูก skip)

สถานะ: ยังไม่มีบั๊กที่ทำให้ระบบพังทันที แต่มีช่องโหว่ด้านความปลอดภัย/ความถูกต้องของ cache หลายจุดที่ควรแก้

---

## 🔴 ความสำคัญสูง (ควรแก้ก่อน)

- [x] **Race condition: refresh token ถูกใช้ซ้ำได้ (double-spend)**
  แก้แล้ว: `Revoke` เปลี่ยนเป็น compare-and-swap (`WHERE id = ? AND revoked = false`)
  คืน `RowsAffected`; ถ้าได้ 0 แถวใน `refresh.go` ถือเป็น token leak → `RevokeAllForUser`

  `internal/application/user/refresh.go:43-66` — ระหว่าง `FindByTokenHash` กับ `Revoke`
  ไม่มีการล็อก/compare-and-swap แบบ atomic ทำให้ 2 request ที่ถือ refresh token เดียวกัน
  ยิงพร้อมกันผ่านเงื่อนไข `!stored.Revoked` ได้ทั้งคู่ และได้ token pair ใหม่ทั้งคู่
  → ควรให้ `Revoke` คืนจำนวนแถวที่อัปเดต และถ้าเป็น 0 ให้ถือว่าเป็น reuse (revoke ทั้ง family)

- [x] **Cache invalidation ล้มเหลว = การ revoke ไม่มีผล (reuse detection ถูกปิด)**
  แก้แล้ว: `RevokeAllForUser` พยายามลบทุก key แต่ถ้ามีอันไหน fail จะ return error
  (พร้อมนับจำนวนที่ fail) ให้ caller alert ได้

  `internal/adapters/cacheduser/refresh_token_repository.go:99-113` —
  ใน `RevokeAllForUser` ถ้า `cache.Delete` ล้มเหลว จะแค่ `log.Printf` แล้วไปต่อ
  entry ใน Redis ยังเก็บ `Revoked=false` ไว้ได้นานถึง `JWT_REFRESH_TTL` (default 168h = 7 วัน)
  ในช่วงนี้ token ที่ถูก revoke ไปแล้วยัง refresh สำเร็จ และ reuse detection จะไม่ทำงานเลย
  → พิจารณา: ให้ `RevokeAllForUser` คืน error เมื่อ cache delete ล้มเหลว, ลด TTL ของ cache entry,
  หรือเก็บ user→token index ใน Redis เพื่อลบเป็นชุด

- [~] **Redis ล่ม = ผู้ใช้ทุกคนถูกบังคับ logout ภายใน 15 นาที**
  แก้บางส่วน: `/api/v1/health` ตอนนี้ ping Redis ด้วย ถ้า Redis ล่มจะคืน HTTP 503
  `{"status":"degraded"}` ให้ monitoring alert ได้
  ยังไม่ทำ: degraded mode fallback ไป Postgres — ขัดกับ ADR 0005 (fail-closed) ต้องให้ทีมตัดสินใจก่อน

  `internal/adapters/cacheduser/refresh_token_repository.go:56-77` — path นี้ fail-closed
  (ตาม ADR 0005) ถ้า Redis unreachable การ refresh ทั้งหมดจะล้มเหลว
  access token หมดอายุใน `JWT_ACCESS_TTL` (15m) แล้วผู้ใช้จะเข้าระบบต่อไม่ได้
  → ยืนยันว่านี่คือ trade-off ที่ตั้งใจ และเพิ่ม health check / alert สำหรับ Redis + พิจารณา
  degraded mode ที่ fallback ไป Postgres ชั่วคราวเมื่อ Redis ล่มทั้งคลัสเตอร์

- [x] **error จาก reuse-detection revoke ถูกกลืน**
  แก้แล้ว: ทั้ง reuse path และ double-spend path log เป็น `ERROR refresh: ...` พร้อม user id

  `internal/application/user/refresh.go:49` — `_ = uc.refresh.RevokeAllForUser(...)`
  ถ้า revoke-all ล้มเหลว family ของ attacker จะไม่ถูกฆ่า และไม่มี log/alert
  → อย่างน้อยควร `log` เป็น level error และส่ง metric

---

## 🟠 ความสำคัญปานกลาง

- [x] **CORS เปิดกว้าง (`*`) เมื่อไม่ตั้งค่า `CORS_ALLOW_ORIGINS`**
  แก้แล้ว: ถ้า `APP_ENV != local` และไม่ตั้ง `CORS_ALLOW_ORIGINS` จะ `log.Fatalf` ไม่ยอม start

- [x] **access token ถูกส่งผ่าน query string (`?token=`)**
  แก้แล้ว: `AuthQuery` เช็ค header `Connection: upgrade` + `Upgrade: websocket` ก่อน
  ถ้าไม่ใช่ WS handshake คืน 426 Upgrade Required — ใช้ query token ได้เฉพาะ WS
  (ticket token อายุสั้น = follow-up, ใส่ note ไว้ในโค้ดแล้ว)

- [x] **`cachedlocation` ไม่ invalidate cache ตอน `Delete`**
  แก้แล้ว: `Delete` evict `location:fullpath:<id>` แบบ best-effort หลังลบสำเร็จ

- [x] **audit write ใช้ request-scoped context หลัง handler จบแล้ว**
  แก้แล้ว: ใช้ `context.WithTimeout(context.WithoutCancel(c.Context()), 5s)` สำหรับเขียน audit

- [x] **`c.IP()` ไม่ได้ตั้ง trusted proxy**
  แก้แล้ว: เพิ่ม env `TRUSTED_PROXIES` + `PROXY_HEADER` → set `TrustProxy` /
  `TrustProxyConfig.Proxies` / `ProxyHeader` ใน Fiber config เมื่อมีค่า

---

## 🟡 ความสำคัญต่ำ / ปรับปรุงคุณภาพโค้ด

- [x] **`cachedrbac`: gob decode error ถูกกลืนโดยไม่ log**
  แก้แล้ว: เพิ่ม `log.Printf` เมื่อ decode cache ล้มเหลวก่อน fallback ไป Postgres

- [x] **role_permissions ยังไม่มี invalidation path**
  ใส่ `TODO(rbac-invalidation)` / `TODO(location-invalidation)` note ไว้ในโค้ดแล้ว

- [ ] **โค้ดซ้ำในชั้น cache decorator (Duplicated Code)**
  ยังไม่แก้ — เป็น refactor ที่มีความเสี่ยง และ policy แต่ละชั้นต่างกัน (fail-open vs fail-closed,
  json vs gob) แนะนำทำแยกเป็นงาน refactor ต่างหากพร้อม test coverage

- [x] **`jwt.parse` ไม่ได้ระบุ `WithValidMethods`**
  แก้แล้ว: เพิ่ม `jwtlib.WithValidMethods([]string{"HS256"})`

- [x] **config Oracle ไม่ validate ตอน boot**
  แก้แล้ว: เพิ่ม env `ORACLE_ENABLED` (explicit); `config.validate()` fail ตอน boot ถ้า
  `ORACLE_ENABLED=true` แต่ไม่มี `ORACLE_DSN`

- [x] **audit `ResourceID` เก็บได้แค่ `:id` ของ parent**
  แก้แล้ว: `auditResourceID` ใช้ route param ตัวสุดท้าย (ลึกสุด) ก่อน แล้ว fallback เป็น `:id`

---

## ✅ จุดที่ตรวจแล้วไม่พบปัญหา

- ไม่พบ SQL injection — ทุก `.Order()` เป็น string คงที่, `.Where()` ใช้ placeholder `?` ทั้งหมด
- password hash ไม่หลุดลง audit metadata (`toUserResponse` ตัด `PasswordHash` ออกก่อน Snapshot/ChangeSet)
- refresh token hashing (SHA-256) เหมาะสมกับ use case, มี `jti` กัน token ชนกันใน constraint
- login ไม่ leak ว่า user มีอยู่จริงหรือไม่ (คืน `ErrUnauthorized` เหมือนกันทั้ง 2 กรณี)
- CSRF header guard + SameSite=Strict + httpOnly cookie + path-scoped cookie ตั้งค่าถูกต้อง
- graceful shutdown ของ auto-reorder job จัดการ context/channel ถูกต้อง
