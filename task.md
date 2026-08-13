# Thanes LIMS Backend — งานที่เหลือ (Phase ถัดไป)

Phase 1 (Auth/Sample/TestResult) และ Phase 2 (Equipment/Inventory+PO/Documents/Environment/Notification) เสร็จสมบูรณ์แล้ว พร้อม unit test และ verify จริงกับ Postgres+MinIO ผ่านหมด รายการด้านล่างคือสิ่งที่ถูก defer ไว้ตอน grilling session ตั้งแต่ต้น หรือพบว่าควรทำเพิ่มระหว่างพัฒนา

## Real-time / Integration

- [x] Environment module: เพิ่ม WebSocket broadcast layer สำหรับ alert แบบ real-time (ตอนนี้เป็น REST เท่านั้นตามที่ตกลงไว้ใน Phase 2) — ใช้ `contrib/websocket` ของ Fiber v3 (หมายเหตุ: `gofiber/contrib/websocket` ยังไม่รองรับ Fiber v3 จริง — ใช้ `github.com/fasthttp/websocket` ตรงกับ `c.RequestCtx()` แทน ดู `internal/adapters/http/environment/ws_hub.go` + route `GET /environment/alerts/ws` (auth ผ่าน query param `?token=`))
- [x] เชื่อม `Notifier` port (มี `AsNotifier` adapter พร้อมใช้ใน `internal/application/notification/create_notification.go` แล้ว) เข้ากับ use case ของ module อื่นที่ควร trigger notification จริง:
  - [x] Environment: สร้าง alert ใหม่ (crit/warn) → แจ้งเตือน
  - [x] Inventory: item ต่ำกว่า min → แจ้งเตือน (ตอนนี้ reorder ต้องกดเองผ่าน `POST /inventory/:id/reorder`)
  - [x] TestResult: approve เสร็จ → แจ้งเตือนผู้เกี่ยวข้อง

## Reporting

- [x] แทนที่ stub endpoint `GET /tests/:id/report` และ `GET /audit/export` (ตอนนี้คืน 501) ด้วย PDF generation จริงด้วย `gofpdf` เมื่อ requirement เรื่อง format ชัดเจน (หมายเหตุ: format ยังไม่มีสเปกจากธุรกิจ จึงออกแบบ layout เองแบบเรียบง่าย — เพิ่ม field/ปรับ branding ได้ภายหลังถ้ามี requirement ชัดเจนกว่านี้)
  - `GET /tests/:id/report`: PDF ผลการทดสอบ + ข้อมูลตัวอย่าง + chain-of-custody trail (`internal/application/testresult/generate_report.go`, `internal/adapters/pdf/testresult_report.go`)
  - `GET /audit/export`: PDF ตาราง audit log พร้อม filter ช่วงเวลา `?from=&to=` (RFC3339 หรือ `YYYY-MM-DD`), จำกัดสิทธิ์เฉพาะ admin/qa (`internal/adapters/http/audit/`, เพิ่ม `AuditLogger.List` ใน `internal/ports/audit/audit.go` ที่แต่ก่อนมีแค่ `Log`)
  - ใช้ `github.com/jung-kurt/gofpdf` + ฟอนต์ Sarabun (SIL OFL, embed ผ่าน `go:embed` ใน `internal/adapters/pdf/fonts/`) เพราะฟอนต์ core ของ PDF ไม่รองรับภาษาไทย

## Testing

- [x] เพิ่ม integration test ด้วย `testcontainers-go` เริ่มจาก `user`/`sample` repository (หมายเหตุ: dependency ยังไม่ได้ติดตั้งจริงตอนเริ่มงานนี้ ต้อง `go get` เพิ่มเอง — เพิ่ม `testcontainers-go` + `.../modules/postgres` และ `golang-migrate/migrate/v4` เป็น Go library สำหรับรัน migration ใส่ container) รันแยกจาก unit test ด้วย `go test -tags=integration ./...` หรือ `make test-integration` (ดู `internal/adapters/postgres/pgtest/container.go` สำหรับ helper spin-up container + apply migrations, `internal/adapters/postgres/user/repository_integration_test.go`, `internal/adapters/postgres/sample/repository_integration_test.go`) — ยังไม่ได้ขยายไปโมดูลอื่น (equipment/inventory/purchaseorder/document/environment/notification/testresult/audit) เหลือเป็นงานต่อไป
- [x] ตั้ง CI pipeline (GitHub Actions หรืออื่นๆ) รัน `go build`, `go vet`, `go test` อัตโนมัติทุก PR (`.github/workflows/ci.yml` — รันบน push/PR เข้า `main`; ไม่รวม `test-integration` เพราะต้องใช้ Docker daemon รัน testcontainers)

## API Documentation

- [x] เพิ่ม Swagger/OpenAPI docs ผ่าน `swaggo/swag` + `gofiber/swagger` (ถูก defer ไว้ตั้งแต่ grilling session) — ใช้ `github.com/gofiber/contrib/v3/swaggo` แทน `gofiber/swagger` เพราะตัวหลัง deprecated แล้วและ maintainer ย้ายไป contrib/v3 (ดู `internal/adapters/http/*/handler.go` สำหรับ `@Summary`/`@Router` annotations ของทุก endpoint, `cmd/api/main.go` สำหรับ general API info + `BearerAuth` security scheme, `cmd/api/routes.go` สำหรับ route `GET /api/v1/swagger/*`) รัน `make swagger` เพื่อ regenerate `docs/` หลังแก้ annotation (ต้อง commit `docs/` เพราะ `cmd/api` import ตรงๆ)

## Equipment

- [x] พิจารณาเพิ่ม `calibration_events` history table แยกจาก `equipment` เพื่อเก็บ audit trail การสอบเทียบแต่ละครั้ง (ตอนนี้เก็บแค่ `last_calibrated_at`/`next_calibration_due` ล่าสุดบน row เดียว) — เพิ่ม migration `000014_create_calibration_events_table` (`equipment_id` FK, `calibrated_at`, `next_calibration_due`, `performed_by`, `notes`), domain `CalibrationEvent` + port `CalibrationRepository` (append-only, ตาม pattern ของ `coc_steps`/`doc_history`), `RecordCalibrationUseCase` แก้ให้ append event ทุกครั้งที่บันทึกการสอบเทียบ, เพิ่ม `ListCalibrationEventsUseCase` + endpoint `GET /equipment/:id/calibration-events` (ยังไม่มี integration test สำหรับ equipment module — เหลือเป็นงานต่อไปตามหัวข้อ Testing ด้านบน)

## Inventory / PurchaseOrder

- [ ] พิจารณาทำ automated reorder (scheduled job ตรวจ item ต่ำกว่า min แล้วสร้าง PO อัตโนมัติ) แทนการกดเองผ่าน endpoint — ตอนนี้เป็น manual ตามที่ตกลงไว้สำหรับ MVP

## RBAC

- [ ] ทบทวนว่า permission ต้องแยกระดับ module หรือไม่ (ตอนนี้เป็น global-per-role ตามที่ตกลงไว้ — ถ้า requirement เปลี่ยนต้องปรับ `internal/domain/user/role.go`)

## Dev Ops

- [x] `git commit` งานทั้งหมด (commit ครบแล้ว, ยังไม่ได้ `git push` ขึ้น origin)
- [ ] เชื่อมต่อ Postgres + MinIO cloud instance จริงสำหรับ demo (ตอนนี้ verify ผ่าน container ชั่วคราวที่ลบทิ้งแล้วเท่านั้น) แล้วรัน `make migrate-up` + `go run ./cmd/seed` กับของจริง
