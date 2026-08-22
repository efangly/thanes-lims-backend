# AI Chatbot (Oracle Select AI) — แผนงาน POC

มาจาก grilling session (2026-08-19) ดู `CONTEXT.md` สำหรับ glossary (Select AI, AI Profile, POC Oracle Instance, Chatbot module)

**สรุปแนวทาง**: เพิ่ม chatbot ถาม-ตอบข้อมูล Sample/TestResult และ Inventory/PurchaseOrder แบบ single-turn โดยใช้ **Oracle Select AI** (NL→SQL เกิดในตัว Oracle DB เอง ผ่าน `DBMS_CLOUD_AI`) บน **Oracle ADB แยกต่างหาก** ที่ provision ไว้แล้ว (ไม่แตะ Postgres ซึ่งยังเป็น production DB หลักของระบบ) เชื่อมจาก Go backend เดิมผ่าน `godror` (ใช้ wallet + Instant Client ที่ติดตั้งไว้แล้วในเครื่องนี้) เป็น module ใหม่ในสถาปัตยกรรม hexagonal เดิม ใช้ JWT auth เดียวกันกับ API หลัก LLM backend ของ Select AI คือ **OpenAI** ข้อมูลใน Oracle เป็น synthetic seed แยกต่างหาก ไม่ sync จาก Postgres เป้าหมาย POC คือ business demo ให้ stakeholder เห็นว่า chatbot ตอบคำถามจากข้อมูลจริงในระบบได้

> **Scope pivot (2026-08-20)**: เดิมตั้งใจใช้ OCI Generative AI แต่ region ของ ADB ที่ provision ไว้ (`ap-samutprakan-1`) ไม่มี OCI Generative AI service ให้ใช้เลย (ไม่มีเมนูใน Console) และ tenancy ไม่ได้ subscribe region อื่นที่มี service นี้ จึงเปลี่ยนไปใช้ **OpenAI** เป็น LLM backend ของ Select AI แทน (`DBMS_CLOUD_AI` รองรับ native อยู่แล้ว) — ดู Phase 3 ด้านล่าง

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

## Phase 3 — Select AI setup

- [x] ~~สร้าง Credential object สำหรับ OCI Generative AI~~ — ลองแล้ว (`scripts/oracle/create_oci_credential.sh`, credential `OCI_GENAI_CRED` สร้างสำเร็จ) แต่ล้มเหลวตอนทดสอบจริง (ดู scope pivot ด้านบน) — เปลี่ยนไปใช้ OpenAI แทน; `create_ai_profile.sh` จะ drop `OCI_GENAI_CRED` ทิ้งตอนสร้าง profile ใหม่
- [ ] สร้าง Credential object สำหรับ OpenAI (`DBMS_CLOUD.CREATE_CREDENTIAL`, username=`OPENAI`) — `scripts/oracle/create_openai_credential.sh` อ่าน `OPENAI_API_KEY` จาก `.env` ตอน execute เท่านั้น ไม่ฝัง secret ไว้ในไฟล์ที่ commit; credential name = `OPENAI_CRED`
  - **รอ**: ยังไม่มี OpenAI API key ต้องสร้างที่ platform.openai.com ก่อน แล้วใส่ `OPENAI_API_KEY` ใน `.env`
- [ ] Grant network ACL ให้ `CHATBOT_APP` เรียก `api.openai.com` ได้ (`DBMS_NETWORK_ACL_ADMIN.APPEND_HOST_ACE`) — `scripts/oracle/grant_openai_network_acl.sh`, ต้องรันด้วย user `ADMIN` (ครั้งเดียว) เพราะ `CHATBOT_APP` ไม่มีสิทธิ์นี้
  - **รอ**: ยังไม่มี ADMIN password ของ ADB นี้ ต้องขอจากผู้ดูแล ใส่ `ORACLE_ADMIN_DSN` ใน `.env` ชั่วคราวตอนรันสคริปต์นี้ครั้งเดียว
- [ ] สร้าง AI Profile ด้วย `DBMS_CLOUD_AI.CREATE_PROFILE` ระบุ provider = OpenAI, object_list จำกัดเฉพาะ 4 ตารางใน Phase 1 (`samples`, `test_results`, `inventory_items`, `purchase_orders`), model = `gpt-5.4`, ไม่เปิด `conversation` (multi-turn อยู่นอกขอบเขต) — `scripts/oracle/create_ai_profile.sh` (profile name เดิม `CHATBOT_AI_PROFILE` เพื่อไม่ต้องแก้ Phase 4)
- [ ] ทดสอบ `SELECT AI narrate <คำถามภาษาไทย/อังกฤษ>` ตรงผ่าน SQL client (นอก Go) ก่อน ยืนยันว่า AI Profile ตอบคำถาม scenario ที่ออกแบบใน Phase 2 ได้ถูกต้อง — ปรับ comment/schema ใน Phase 1 ถ้าคำตอบยังไม่แม่น

## Phase 4 — Chatbot module (Go backend)

- [ ] เพิ่ม domain `internal/domain/chatbot` (เช่น `ChatQuestion`, `ChatAnswer`)
- [ ] เพิ่ม port `internal/ports/chatbot` (interface สำหรับเรียก Select AI)
- [ ] เพิ่ม adapter `internal/adapters/oracle/chatbot` ที่เปิด connection ผ่าน godror และรัน `SELECT AI narrate ...` กับ AI Profile จาก Phase 3
- [ ] เพิ่ม use case `internal/application/chatbot/ask.go`
- [ ] เพิ่ม HTTP handler + route `POST /chat` ใน `internal/adapters/http/chatbot/` ผูก JWT auth เดียวกับ route อื่น (พิจารณาจำกัด role ที่เข้าถึงได้ตาม RBAC เดิม)
- [ ] เพิ่ม Swagger annotation (`@Summary`/`@Router`) แล้วรัน `make swagger` ให้ endpoint ใหม่ขึ้นใน `docs/`
- [ ] เพิ่ม `ORACLE_DSN`/wallet path ใน `cmd/api/main.go` wiring (สร้าง oracle client ตอน startup, graceful fail ถ้าต่อไม่ได้แทนที่จะ crash ทั้ง API เพราะ chatbot เป็น POC feature เสริม)

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
