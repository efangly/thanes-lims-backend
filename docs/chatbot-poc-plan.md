# AI Chatbot (Oracle Select AI) — แผนงาน POC

มาจาก grilling session (2026-08-19) ดู `CONTEXT.md` สำหรับ glossary (Select AI, AI Profile, POC Oracle Instance, Chatbot module)

**สรุปแนวทาง**: เพิ่ม chatbot ถาม-ตอบข้อมูล Sample/TestResult และ Inventory/PurchaseOrder แบบ single-turn โดยใช้ **Oracle Select AI** (NL→SQL เกิดในตัว Oracle DB เอง ผ่าน `DBMS_CLOUD_AI`) บน **Oracle ADB แยกต่างหาก** ที่ provision ไว้แล้ว (ไม่แตะ Postgres ซึ่งยังเป็น production DB หลักของระบบ) เชื่อมจาก Go backend เดิมผ่าน `godror` (ใช้ wallet + Instant Client ที่ติดตั้งไว้แล้วในเครื่องนี้) เป็น module ใหม่ในสถาปัตยกรรม hexagonal เดิม ใช้ JWT auth เดียวกันกับ API หลัก LLM backend ของ Select AI คือ **OpenAI** ข้อมูลใน Oracle เป็น synthetic seed แยกต่างหาก ไม่ sync จาก Postgres เป้าหมาย POC คือ business demo ให้ stakeholder เห็นว่า chatbot ตอบคำถามจากข้อมูลจริงในระบบได้

> **Scope pivot (2026-08-20)**: เดิมตั้งใจใช้ OCI Generative AI แต่ region ของ ADB ที่ provision ไว้ (`ap-samutprakan-1`) ไม่มี OCI Generative AI service ให้ใช้เลย (ไม่มีเมนูใน Console) และ tenancy ไม่ได้ subscribe region อื่นที่มี service นี้ จึงเปลี่ยนไปใช้ **OpenAI** เป็น LLM backend ของ Select AI แทน (`DBMS_CLOUD_AI` รองรับ native อยู่แล้ว) — ดู Phase 3 ด้านล่าง

> **LLM pivot (2026-09-02)**: เลิกใช้ **Oracle Select AI** — เปลี่ยนมาให้ **Go backend เรียก Claude API** โดยตรง (`github.com/anthropics/anthropic-sdk-go`, model `claude-haiku-4-5-20251001`). เหตุผล: Phase 3 (Select AI) ติดค้างยาว เพราะไม่มี OpenAI API key, ไม่มี ADMIN password สำหรับ grant network ACL ให้ ADB เรียก `api.openai.com`, และ Select AI ผูกกับ provider ที่ Oracle รองรับเท่านั้น. สถาปัตยกรรมใหม่: **tool-use loop** — Go ส่ง schema + คำถามให้ Claude → Claude เรียก tool `run_sql` → Go รัน SELECT แบบ read-only (user `CHATBOT_RO` + Go guard) แล้วป้อนผลกลับ → Claude สรุปเป็นภาษาไทย. Oracle ADB เหลือหน้าที่เดียวคือรัน SQL คืนแถวข้อมูล. Phase 0–2 ใช้ได้เหมือนเดิม; Phase 3 เดิมถูกแทนที่ทั้งหมด (สคริปต์ Select AI เก็บไว้เป็นประวัติ). ดู `docs/adr` / โค้ด `internal/adapters/anthropic/chatbot`, `internal/adapters/oracle/chatbot`

> **DB pivot (2026-08-21)**: เปลี่ยนไปใช้ Oracle ADB instance ใหม่ (region `ap-samutprakan-1` เดิม, wallet re-issued ที่ path เดิม `~/oracle/wallet`) — service name เปลี่ยนจาก `dblims_high` เป็น **`limsdb_high`** (DB name `G6C725E35D8F9B9_LIMSDB`) จึงต้องรีรัน Phase 0–2 ทั้งหมดบน instance ใหม่ (สร้าง user, schema, seed ใหม่ — instance เก่าไม่มีข้อมูลติดมา) อัปเดต service name ในทุกไฟล์ (`.env`, `.env.example`, `scripts/oracle/*.sql`, `grant_openai_network_acl.sh`) เรียบร้อยแล้ว; รหัสผ่าน `CHATBOT_APP` และ credential/profile objects ของ Phase 3 (ยังไม่เริ่ม) ไม่กระทบเพราะยังไม่เคยสร้างบน instance ใหม่

---

## Phase 0 — Oracle connectivity (godror + wallet)

- [x] ตรวจสอบ wallet (`cwallet.sso`, `tnsnames.ora`, `sqlnet.ora`) และ Instant Client ที่ติดตั้งไว้ในเครื่องนี้ ว่าชี้ไปที่ POC ADB instance ที่ถูกต้อง และต่อผ่าน `sqlplus`/`sqlcl` ได้จริงก่อนเริ่มเขียนโค้ด — wallet ที่ใช้จริงคือ `~/oracle/wallet` (ตรงกับ `TNS_ADMIN`), Instant Client ที่ `~/oracle/instantclient/instantclient_23_26`; ต่อสำเร็จด้วย service **`limsdb_high`** (user `ADMIN`) ผ่าน `sqlcl` — instance ใหม่ (2026-08-21, ดู DB pivot ด้านบน), DB name `G6C725E35D8F9B9_LIMSDB`
- [x] เพิ่ม `github.com/godror/godror` ใน `go.mod` (`go get github.com/godror/godror@latest` — v0.51.4)
- [x] สร้าง dedicated DB user/schema สำหรับ chatbot module แยกจาก ADMIN (`scripts/oracle/000_create_app_user.sql`, รันครั้งเดียวด้วย ADMIN) — user `CHATBOT_APP` มีสิทธิ์ `CREATE SESSION`/`CREATE TABLE`/`CREATE VIEW`, quota unlimited บน `DATA`, และ `EXECUTE` บน `DBMS_CLOUD`/`DBMS_CLOUD_AI` (เตรียมไว้สำหรับ Phase 3) — รันซ้ำบน instance ใหม่แล้ว (2026-08-21)
- [x] ตั้งค่า env vars ใหม่ (`ORACLE_DSN`, `ORACLE_TNS_ADMIN`) ใน `internal/config` (`Config.OracleDSN`, `Config.OracleTNSAdmin`) — แยกจาก `DATABASE_URL` (Postgres) เดิมโดยสิ้นเชิง, ไม่ required เพราะ chatbot เป็น feature เสริม; DSN format คือ `CHATBOT_APP/"<password>"@limsdb_high` (service name อัปเดตตาม instance ใหม่)
- [x] เขียน connectivity smoke test เล็กๆ — `cmd/oracle-ping` ยืนยันว่า Go เชื่อม ADB ผ่าน godror+wallet ได้จริง (`go run ./cmd/oracle-ping`)
  - **macOS gotcha**: ต้อง export `DYLD_LIBRARY_PATH=/Users/tng-mac-01/oracle/instantclient/instantclient_23_26` ใน shell ก่อนรัน (`go run`/`go build`+exec) เสมอ — dyld resolve library search path ตอน process launch ก่อน Go/godror code จะรัน ดังนั้นตั้งผ่าน `.env`/`os.Setenv` ใน runtime **ไม่มีผล** ต้อง export ใน shell/launchd/systemd/Docker env จริงๆ (ดู `.env` comment)

## Phase 1 — Oracle schema (mirror จาก Postgres)

- [x] ออกแบบ DDL สำหรับตาราง `samples`, `test_results`, `inventory_items`, `purchase_orders` โดย mirror field จาก domain model เดิมใน `internal/domain/sample`, `internal/domain/testresult`, `internal/domain/inventory`, `internal/domain/purchaseorder` — ไม่มีตารางย่อยเพิ่ม (PurchaseOrder เดิมอ้าง item เดียวต่อ PO ไม่มี line items) และไม่ได้ mirror CoC step audit trail เพราะไม่อยู่ใน scope ที่ grill ไว้ (Sample/TestResult/Inventory/PurchaseOrder เท่านั้น)
- [x] เพิ่ม `COMMENT ON TABLE` / `COMMENT ON COLUMN` อธิบายความหมายเป็นภาษาไทยสำหรับ Select AI (ครบ 4 table comments + 29 column comments) — เช่น `samples.status`: "สถานะของตัวอย่าง: pending=รอดำเนินการ, testing=กำลังทดสอบ, completed=ทดสอบเสร็จสิ้น, transferred=ส่งต่อแผนกอื่น"
- [x] เก็บ DDL เป็นไฟล์ script แยก `scripts/oracle/001_schema.sql` — ไม่ปนกับ `migrations/` เดิมของ Postgres/golang-migrate
- [x] รัน DDL จริงกับ POC ADB (ด้วย user `CHATBOT_APP` ไม่ใช่ ADMIN) ยืนยันสร้างตารางสำเร็จครบ 4 ตาราง พร้อม FK (`test_results.sample_id → samples.id`, `purchase_orders.item_id → inventory_items.id`) — รันซ้ำบน instance ใหม่แล้ว (2026-08-21)

## Phase 2 — Synthetic seed data

- [x] ออกแบบชุดข้อมูลจำลองที่ครอบคลุมคำถาม demo ที่ตั้งใจจะถาม (เช่น sample ที่ค้าง/เกินกำหนด, test result ที่ fail, inventory ต่ำกว่า min, PO ที่ pending) — ระบุ scenario คำถาม-คำตอบตัวอย่างไว้ล่วงหน้าเพื่อ verify ทีหลัง (ดูตาราง demo scenarios ด้านล่าง)
- [x] เขียน seed script สำหรับ Oracle — `scripts/oracle/002_seed.sql` (ขยายจากเดิม 12/12/10/8 เป็น **40 samples, 39 test_results, 12 inventory_items, 15 purchase_orders**) — ข้อมูลสร้างใหม่ทั้งหมด ไม่ดึงจาก Postgres; วันที่ผูกกับ `SYSDATE`/`TRUNC(SYSDATE)` ตอนรัน เพื่อให้คำถามแนว "ค้างเกิน N วัน" ยังใช้ได้ทุกครั้งที่ seed ใหม่
- [x] รัน seed จริงกับ POC ADB (instance ใหม่, 2026-08-21) และตรวจนับ record ว่าครบตามที่ออกแบบ — ยืนยันแล้ว: `samples`=40, `test_results`=39, `inventory_items`=12, `purchase_orders`=15

### Demo Q&A scenarios (สำหรับ verify ใน Phase 3/5)

> อัปเดต 2026-08-21 ตามข้อมูลจริงบน instance ใหม่ (dataset ขยายจาก 12/12/10/8 เป็น 40/39/12/15 — คำตอบด้านล่างคือผลจริงจากการ query ตรง ไม่ใช่ค่าคาดเดา)

| # | คำถาม (ตัวอย่าง) | คำตอบที่คาดหวังจากข้อมูล seed |
|---|---|---|
| 1 | มี sample อะไรบ้างที่ยังค้างสถานะ pending เกิน 7 วัน? | SMP-2569-00021 (ค้าง 15 วัน), SMP-2569-00022 (14 วัน), SMP-2569-00027 (10 วัน) |
| 2 | test result อะไรบ้างที่ flag เป็น hi หรือ lo (ผลผิดปกติ)? | 15 รายการ เช่น TR-2569-00005 (IgG, hi), TR-2569-00007 (Coliform Count, hi), TR-2569-00009 (LDL Cholesterol, hi), TR-2569-00010 (HDL Cholesterol, lo), TR-2569-00030 (Lead (Pb), lo) |
| 3 | สารเคมี/วัสดุคงคลังอะไรบ้างที่ต่ำกว่าจุดสั่งซื้อขั้นต่ำ (min_qty)? | INV-0001 น้ำยา Glucose Reagent (5/20), INV-0002 ถุงมือไนไตรไซส์ M (3/15), INV-0003 หลอดเก็บซีรัม (8/30), INV-0006 Heavy Metal Kit (6/12), INV-0009 TSH Immunoassay Kit (14/15), INV-0011 HBsAg Test Kit (9/15) |
| 4 | มีใบสั่งซื้อ (PO) ที่ยังรออนุมัติหรือส่งให้ vendor แล้วกี่ใบ? | pending_approval: PO-2569-0012, PO-2569-0014, PO-2569-0015 (3 ใบ); sent_to_vendor: PO-2569-0009, PO-2569-0011, PO-2569-0013 (3 ใบ) |
| 5 | ตัวอย่างของ "วิภา สายใจ" มีอะไรบ้าง และสถานะเป็นอย่างไร? | 10 รายการ: SMP-2569-00003/00004/00010/00015/00018/00023/00024/00029 (completed), SMP-2569-00037/00038 (testing) |
| 6 | PO ของรายการ "ถุงมือไนไตรไซส์ M" (INV-0002) มีสถานะอะไรบ้าง? | PO-2569-0001 (received), PO-2569-0013 (sent_to_vendor) |

## Phase 3 — ~~Select AI setup~~ (superseded 2026-09-02)

> ทั้ง section ถูกแทนที่ด้วย "Phase 3′ — Claude API setup" ด้านล่าง. สคริปต์ `scripts/oracle/create_openai_credential.sh`, `grant_openai_network_acl.sh`, `create_ai_profile.sh`, `create_oci_credential.sh` เก็บไว้เป็นประวัติ ไม่ใช้แล้ว. ไม่ต้องมี OpenAI credential / network ACL / AI Profile / `DBMS_CLOUD_AI` อีกต่อไป

## Phase 3′ — Claude API setup

- [x] เพิ่ม `github.com/anthropics/anthropic-sdk-go` ใน `go.mod` (v1.69.0)
- [ ] สร้าง read-only DB user `CHATBOT_RO` — `scripts/oracle/003_create_readonly_user.sql` (รันครั้งเดียวด้วย **ADMIN**): grant `SELECT` เฉพาะ 4 ตาราง + synonym ให้เรียกชื่อตารางไม่ต้อง prefix schema
  - **รอ**: ต้องใช้ ADMIN password (ใส่ `ORACLE_ADMIN_DSN` ชั่วคราว). fallback: ใช้ `ORACLE_DSN` (CHATBOT_APP) ไปก่อน — Go guard + `SET TRANSACTION READ ONLY` ยังกันการเขียนได้
- [ ] ตั้ง env: `ANTHROPIC_API_KEY` (หรือ `ant auth login`), `CHATBOT_MODEL=claude-haiku-4-5-20251001`, `ORACLE_CHATBOT_DSN=CHATBOT_RO/"..."@limsdb_high`
- [ ] ทดสอบ `POST /chat` กับ scenario Phase 2 — ปรับ `systemPrompt` ใน `internal/adapters/anthropic/chatbot/assistant.go` (มี schema + Thai comments ฝังอยู่) ถ้าคำตอบยังไม่แม่น

## Phase 4 — Chatbot module (Go backend) — done

- [x] domain `internal/domain/chatbot` (`Question`, `Answer{Text, SQLQueries, Rows, ElapsedMS}`)
- [x] ports `internal/ports/chatbot` — `SQLRunner` (รัน SELECT read-only) + `Assistant` (tool-use loop)
- [x] adapter `internal/adapters/oracle/chatbot/runner.go` — `validateSelect` guard (single SELECT/WITH, block DML/DDL/comment) + `SET TRANSACTION READ ONLY` + timeout 15s + cap 200 แถว
- [x] adapter `internal/adapters/anthropic/chatbot/assistant.go` — manual tool-use loop (`client.Messages.New`), tool `run_sql`, สูงสุด 5 turn, parse tool input ด้วย `json.Unmarshal`
- [x] use case `internal/application/chatbot/ask.go` — validate คำถาม (ไม่ว่าง, ≤ 500 ตัวอักษร)
- [x] HTTP handler + route `POST /chat` ใน `internal/adapters/http/chatbot/` — `middleware.Auth` + `RequirePermission(rbac.ModuleChatbot, rbac.ActionView)`
- [x] RBAC: `ModuleChatbot` + migration `000037_add_chatbot_module_permissions` (grant `chatbot:view` ให้ทุก role)
- [x] Swagger annotation + regen `docs/` (endpoint `/chat` ขึ้นแล้ว)
- [x] wiring `cmd/api/{main,routes}.go` — สร้าง `chatbotDB` (RO DSN) ตอน startup, graceful fail, mount `/chat` เฉพาะเมื่อต่อ ADB สำเร็จ
- [x] config `internal/config` — `AnthropicAPIKey`, `ChatbotModel`, `OracleChatbotDSN`
- [x] unit test — `ask_test.go` (validate), `runner_test.go` (`validateSelect`)
- [x] prompt caching — `cache_control` breakpoint บน system block (schema+rules+tool byte-identical ทุก request). **หมายเหตุ**: system prompt ปัจจุบัน ~1.6k tokens ต่ำกว่า minimum cacheable prefix ของ Haiku 4.5 (4096 tokens) → marker เป็น no-op จนกว่าจะขยาย prompt เกิน 4k (เช่นเพิ่ม few-shot examples) หรือเปลี่ยนไปใช้ Opus/Sonnet (512/1024). `Answer.CacheReadTokens/CacheWriteTokens` ใช้ verify

### สถานะทดสอบจริง (2026-09-02, `cmd/chatbot-ask`)

รัน scenario 1/3/4 + negative test ผ่านครบ (Claude API + Oracle จริง). latency ~5–8s/คำถาม (1 query). scenario 3/4 ตรงกับ demo doc เป๊ะ. scenario 1 คืน 7 แถว (seed มี pending มากกว่า snapshot เดิม + junk rows `SMP-TEST-*` จาก `cmd/oracle-insert-test` — ควร re-run `002_seed.sql` ล้างก่อน demo). ยังต่อด้วย `CHATBOT_APP` (Go guard + read-only txn กันเขียนได้ — negative test ยืนยัน) — ควรสร้าง `CHATBOT_RO` ด้วย `003_*.sql`

## Phase 5 — Demo verification

- [ ] รันคำถาม demo scenario ทั้งหมดจาก Phase 2 ผ่าน `POST /chat` จริง (ผ่าน Swagger UI ตามที่ตกลงไว้) ยืนยันคำตอบตรงกับข้อมูลที่ seed ไว้
- [ ] จับเวลาตอบสนอง (latency) คร่าวๆ เผื่อต้องอธิบาย stakeholder เรื่อง response time
- [ ] เตรียมสคริปต์/ลำดับคำถามสำหรับ demo จริงให้ผู้บริหาร (ลำดับที่แสดงคุณค่าได้ชัดเจนที่สุด)
- [ ] อัปเดต `CONTEXT.md` / เอกสารนี้ถ้ามีการปรับ scope ระหว่างทำจริง

---

## นอกขอบเขต (out of scope สำหรับ POC รอบนี้)

รายการเหล่านี้ถูกตัดออกจาก POC โดยตั้งใจตาม grilling session — บันทึกไว้เผื่อพิจารณาต่อในอนาคต ไม่ใช่ checklist ที่ต้องทำ:

- Multi-turn conversation (จำ context ข้ามคำถาม)
- Chat UI/frontend (POC ใช้ Swagger/Postman demo เท่านั้น)
- Domain module อื่นนอกจาก Sample/TestResult/Inventory/PurchaseOrder (เช่น Equipment, Document, Environment, Notification, User, Audit)
- การ sync ข้อมูลจริงจาก Postgres ไปยัง Oracle (ข้อมูลใน POC เป็น synthetic ล้วน)
- Provider LLM อื่นนอกจาก OpenAI สำหรับ Select AI (เดิมตั้งใจ OCI Generative AI แต่เปลี่ยนเพราะ region ไม่รองรับ — ดู scope pivot ด้านบน)
