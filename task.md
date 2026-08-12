# Thanes LIMS Backend — งานที่เหลือ (Phase ถัดไป)

Phase 1 (Auth/Sample/TestResult) และ Phase 2 (Equipment/Inventory+PO/Documents/Environment/Notification) เสร็จสมบูรณ์แล้ว พร้อม unit test และ verify จริงกับ Postgres+MinIO ผ่านหมด รายการด้านล่างคือสิ่งที่ถูก defer ไว้ตอน grilling session ตั้งแต่ต้น หรือพบว่าควรทำเพิ่มระหว่างพัฒนา

## Real-time / Integration

- [ ] Environment module: เพิ่ม WebSocket broadcast layer สำหรับ alert แบบ real-time (ตอนนี้เป็น REST เท่านั้นตามที่ตกลงไว้ใน Phase 2) — ใช้ `contrib/websocket` ของ Fiber v3
- [ ] เชื่อม `Notifier` port (มี `AsNotifier` adapter พร้อมใช้ใน `internal/application/notification/create_notification.go` แล้ว) เข้ากับ use case ของ module อื่นที่ควร trigger notification จริง:
  - [ ] Environment: สร้าง alert ใหม่ (crit/warn) → แจ้งเตือน
  - [ ] Inventory: item ต่ำกว่า min → แจ้งเตือน (ตอนนี้ reorder ต้องกดเองผ่าน `POST /inventory/:id/reorder`)
  - [ ] TestResult: approve เสร็จ → แจ้งเตือนผู้เกี่ยวข้อง

## Reporting

- [ ] แทนที่ stub endpoint `GET /tests/:id/report` และ `GET /audit/export` (ตอนนี้คืน 501) ด้วย PDF generation จริงด้วย `gofpdf` เมื่อ requirement เรื่อง format ชัดเจน

## Testing

- [ ] เพิ่ม integration test ด้วย `testcontainers-go` (dependency ติดตั้งไว้แล้ว) เริ่มจาก `user`/`sample` repository ก่อน แล้วขยายไปทุก module — รันแยกจาก unit test เช่น `go test -tags=integration ./...`
- [ ] ตั้ง CI pipeline (GitHub Actions หรืออื่นๆ) รัน `go build`, `go vet`, `go test` อัตโนมัติทุก PR

## API Documentation

- [ ] เพิ่ม Swagger/OpenAPI docs ผ่าน `swaggo/swag` + `gofiber/swagger` (ถูก defer ไว้ตั้งแต่ grilling session)

## Equipment

- [ ] พิจารณาเพิ่ม `calibration_events` history table แยกจาก `equipment` เพื่อเก็บ audit trail การสอบเทียบแต่ละครั้ง (ตอนนี้เก็บแค่ `last_calibrated_at`/`next_calibration_due` ล่าสุดบน row เดียว)

## Inventory / PurchaseOrder

- [ ] พิจารณาทำ automated reorder (scheduled job ตรวจ item ต่ำกว่า min แล้วสร้าง PO อัตโนมัติ) แทนการกดเองผ่าน endpoint — ตอนนี้เป็น manual ตามที่ตกลงไว้สำหรับ MVP

## RBAC

- [ ] ทบทวนว่า permission ต้องแยกระดับ module หรือไม่ (ตอนนี้เป็น global-per-role ตามที่ตกลงไว้ — ถ้า requirement เปลี่ยนต้องปรับ `internal/domain/user/role.go`)

## Dev Ops

- [ ] `git commit` งานทั้งหมด (ปัจจุบัน repo มีแต่ untracked files ยังไม่ได้ commit)
- [ ] เชื่อมต่อ Postgres + MinIO cloud instance จริงสำหรับ demo (ตอนนี้ verify ผ่าน container ชั่วคราวที่ลบทิ้งแล้วเท่านั้น) แล้วรัน `make migrate-up` + `go run ./cmd/seed` กับของจริง
