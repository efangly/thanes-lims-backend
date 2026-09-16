# Chatbot Migration — Acceptance Checklist

ใช้ตรวจ regression ตอน Phase 5.2 (verification) ของแผนย้าย chatbot ไป NestJS + LangGraph.js
(ดู `/Users/tng-mac-01/.claude/plans/ai-chatbot-groovy-spindle.md`). แปลงมาจาก demo Q&A
scenarios เดิมใน `docs/chatbot-poc-plan.md` (Phase 2, seed data บน instance `limsdb_high`,
40 samples / 39 test_results / 12 inventory_items / 15 purchase_orders) ให้เป็น assertion
ที่ทดสอบซ้ำได้กับ Postgres (ต้อง seed fixture เทียบเท่าใน Postgres ก่อนรัน — ดูหมายเหตุท้ายไฟล์)

สถานะ: `[ ]` ยังไม่ verify กับ NestJS/MCP ตัวใหม่ — ติ๊กเมื่อรันผ่านจริงใน Phase 5.2

---

## Scenario 1 — Sample ค้าง pending เกิน 7 วัน

- **Given**: seed state ที่มี sample สถานะ `pending` ค้างนานกว่า 7 วันนับจากวันที่รัน (วันที่ผูกกับเวลาที่ seed จริง ไม่ fix ตายตัว)
- **When**: ถาม `"มี sample อะไรบ้างที่ยังค้างสถานะ pending เกิน 7 วัน?"`
- **Then**: คำตอบต้องระบุ sample ID ที่ตรงเงื่อนไขทั้งหมด พร้อมจำนวนวันที่ค้าง (ตัวอย่างผลจาก Oracle seed เดิม: SMP-2569-00021 ค้าง 15 วัน, SMP-2569-00022 ค้าง 14 วัน, SMP-2569-00027 ค้าง 10 วัน — ตัวเลขจริงจะต่างกันเมื่อ seed ใหม่ใน Postgres แต่ต้องตรงกับ fixture ที่เตรียมไว้)
- [ ] Verify ผ่าน MCP tool `listSamplesByStatus(status="pending", olderThanDays=7)`
- [ ] Verify ผ่าน NestJS endpoint เต็ม flow (คำถามภาษาไทย → คำตอบภาษาไทยที่ถูกต้อง)

## Scenario 2 — Test result ที่ flag hi/lo

- **Given**: seed state ที่มี test result ผิดปกติ (flag `hi` หรือ `lo`)
- **When**: ถาม `"test result อะไรบ้างที่ flag เป็น hi หรือ lo (ผลผิดปกติ)?"`
- **Then**: คำตอบต้องระบุจำนวนรวม + ตัวอย่างรายการ (ผลจาก Oracle seed เดิม: 15 รายการ เช่น TR-2569-00005 IgG hi, TR-2569-00007 Coliform Count hi, TR-2569-00009 LDL Cholesterol hi, TR-2569-00010 HDL Cholesterol lo, TR-2569-00030 Lead (Pb) lo)
- [ ] Verify ผ่าน MCP tool `searchTestResults(flag="hi")` และ `searchTestResults(flag="lo")` (หรือ combined)
- [ ] Verify ผ่าน NestJS endpoint เต็ม flow

## Scenario 3 — Inventory ต่ำกว่า min_qty

- **Given**: seed state ที่มี inventory item ที่ quantity ต่ำกว่า min_qty
- **When**: ถาม `"สารเคมี/วัสดุคงคลังอะไรบ้างที่ต่ำกว่าจุดสั่งซื้อขั้นต่ำ (min_qty)?"`
- **Then**: คำตอบต้องระบุรายการที่ต่ำกว่า min_qty ทั้งหมดพร้อมตัวเลข quantity/min_qty (ผลจาก Oracle seed เดิม: INV-0001 Glucose Reagent 5/20, INV-0002 ถุงมือไนไตรไซส์ M 3/15, INV-0003 หลอดเก็บซีรัม 8/30, INV-0006 Heavy Metal Kit 6/12, INV-0009 TSH Immunoassay Kit 14/15, INV-0011 HBsAg Test Kit 9/15)
- [ ] Verify ผ่าน MCP tool `listInventoryLowStock()`
- [ ] Verify ผ่าน NestJS endpoint เต็ม flow

## Scenario 4 — PO ที่ pending approval / sent to vendor

- **Given**: seed state ที่มี PO สถานะ `pending_approval` และ `sent_to_vendor`
- **When**: ถาม `"มีใบสั่งซื้อ (PO) ที่ยังรออนุมัติหรือส่งให้ vendor แล้วกี่ใบ?"`
- **Then**: คำตอบต้องแยกนับ+รายการตามสถานะ (ผลจาก Oracle seed เดิม: pending_approval 3 ใบ PO-2569-0012/0014/0015, sent_to_vendor 3 ใบ PO-2569-0009/0011/0013)
- [ ] Verify ผ่าน MCP tool `listPurchaseOrders(status="pending_approval")` และ `listPurchaseOrders(status="sent_to_vendor")`
- [ ] Verify ผ่าน NestJS endpoint เต็ม flow

## Scenario 5 — Sample ตาม custodian name

- **Given**: seed state ที่มี custodian ชื่อ "วิภา สายใจ" ถือ sample หลายรายการ
- **When**: ถาม `"ตัวอย่างของ \"วิภา สายใจ\" มีอะไรบ้าง และสถานะเป็นอย่างไร?"`
- **Then**: คำตอบต้องระบุรายการ sample ทั้งหมดของ custodian คนนั้นพร้อมสถานะ (ผลจาก Oracle seed เดิม: 10 รายการ — SMP-2569-00003/00004/00010/00015/00018/00023/00024/00029 completed, SMP-2569-00037/00038 testing)
- [ ] Verify ผ่าน MCP tool `listSamplesByCustodianName(name="วิภา สายใจ")` (รวม name-resolution step)
- [ ] Verify ผ่าน NestJS endpoint เต็ม flow

## Scenario 6 — PO ตาม inventory item

- **Given**: seed state ที่มี PO อ้างถึง inventory item "ถุงมือไนไตรไซส์ M" (INV-0002)
- **When**: ถาม `"PO ของรายการ \"ถุงมือไนไตรไซส์ M\" (INV-0002) มีสถานะอะไรบ้าง?"`
- **Then**: คำตอบต้องระบุ PO ทั้งหมดที่อ้างถึง item นั้นพร้อมสถานะ (ผลจาก Oracle seed เดิม: PO-2569-0001 received, PO-2569-0013 sent_to_vendor)
- [ ] Verify ผ่าน MCP tool `listPurchaseOrdersByItem(inventoryItemId="INV-0002")`
- [ ] Verify ผ่าน NestJS endpoint เต็ม flow

---

## Cross-cutting checks (นอกเหนือจาก 6 scenario ข้างบน)

- [ ] **RBAC**: user role ที่ไม่มี `chatbot:view` ต้องถูก reject 403 ที่ NestJS layer ก่อนถึง MCP call เลย
- [ ] **Out-of-scope question**: ถามเรื่องนอกขอบเขต (เช่น Equipment, User, การตั้งค่า) โมเดลต้องตอบว่าตอบไม่ได้ ไม่ hallucinate
- [ ] **Read-only guarantee**: สั่งให้ "ลบ"/"แก้ไข" ข้อมูลผ่านคำถาม ต้องถูกปฏิเสธ (ไม่มี MCP tool ใดเขียนข้อมูลได้ตั้งแต่ต้น เพราะเป็น typed read-only tools)
- [ ] **Response shape parity**: response ยังมี field ที่ frontend ใช้แสดงผลจริง — `answer` (Markdown), รายการ "สิ่งที่ใช้ตอบ" แบบ human-readable (ทดแทน `sql_queries` เดิมที่โชว์ใน accordion "ดู SQL ที่ใช้" — ดูหมายเหตุด้านล่าง), `elapsed_ms`
- [ ] **Latency**: เฉลี่ยไม่แย่ไปกว่า baseline เดิมมากนัก (baseline เดิม ~5-8 วินาที/คำถาม, ผู้ใช้ยอมรับได้ถึง ~30 วินาทีในบางเคสตาม `docs/chatbot-frontend-integration.md`)
- [ ] **Empty/invalid question**: คำถามว่างหรือเกิน 500 ตัวอักษร ต้องถูก reject ด้วย validation error (400) ไม่ใช่ error 500

---

## หมายเหตุสำคัญที่พบระหว่างทำ checklist นี้

1. **`sql_queries` เป็น UI-facing ไม่ใช่แค่ debug aid** — `docs/chatbot-frontend-integration.md` ข้อ 4 ระบุชัดว่า frontend โชว์ `sql_queries` ใน accordion "ดู SQL ที่ใช้" เพื่อสร้างความน่าเชื่อถือให้ stakeholder เห็นว่าตอบจากข้อมูลจริง — ดังนั้น response DTO ใหม่ (Phase 4) ต้องมี field ทดแทนที่ยัง **โชว์ให้ผู้ใช้เห็นได้** (เช่น `toolCalls: [{tool, args}]` หรือ human-readable summary ของ query ที่ใช้) ไม่ใช่แค่เก็บไว้เป็น internal log
2. **`cache_read_tokens`/`cache_write_tokens` เป็น telemetry ล้วน** — เอกสารเดิมระบุ "ไม่ต้องแสดงให้ผู้ใช้ทั่วไป" ยืนยันว่าตัดออกจาก response DTO ใหม่ได้โดยไม่กระทบ UI
3. **404 มีความหมายพิเศษ**: ปัจจุบัน 404 = "feature ทั้งก้อนไม่พร้อม" (route ไม่ mount เพราะ Oracle ต่อไม่ได้ตอน start) ไม่ใช่ "ไม่พบข้อมูล" — เมื่อย้ายไป NestJS ที่ไม่มี Oracle แล้ว ต้องนิยาม 404/503 ใหม่ให้เหมาะกับ failure mode ของ service ใหม่ (เช่น MCP server ต่อไม่ได้) และอัปเดตพฤติกรรม "frontend ซ่อน UI chatbot เมื่อเจอ 404" ให้ตรงกับ error code ใหม่ ใน Phase 4/5
4. **ข้อมูล seed สำหรับเทียบผล**: scenario ด้านบนอ้างอิงผลลัพธ์จริงจาก Oracle POC seed (`scripts/oracle/002_seed.sql`, instance `limsdb_high`, 40/39/12/15 records) ซึ่งเป็น synthetic data แยกจาก Postgres — ก่อนรัน Phase 5.2 ต้อง **เตรียม fixture ใน Postgres ที่ให้ผลลัพธ์เทียบเท่ากัน** (ไม่จำเป็นต้องมี sample/PO ID ตรงเป๊ะ แต่เงื่อนไข given/then ต้องคงอยู่) แล้วอัปเดตตัวเลขในไฟล์นี้ให้ตรงกับ Postgres fixture จริงก่อนใช้ verify

## Business scope confirmation (ตรงกับที่ระบุใน `docs/chatbot-poc-plan.md` "นอกขอบเขต")

ตอบได้เฉพาะ 4 โดเมน: **Sample, TestResult, Inventory (+InventoryLot), PurchaseOrder**
ไม่รวม: Equipment, Document, Environment, Notification, User, Audit — ขอบเขตนี้คงเดิมหลังย้ายไป NestJS (ไม่ได้ขยาย scope ในรอบ migration นี้)

## RBAC decision (ตรงกับ Phase 2 ของแผน migration)

Permission `chatbot:view` (migration `000037_add_chatbot_module_permissions`) ปัจจุบัน grant ให้
**ทุก role** (Admin, Lab Manager, Scientist, QA, General) — **ตัดสินใจคงไว้แบบเดิมหลังย้าย**
ไม่เพิ่ม fine-grained permission ต่อ domain (เช่น `sample:view` แยก) เพราะไม่มี requirement ใหม่
ที่ต้องจำกัดสิทธิ์ละเอียดกว่านี้ — ทุก role ที่ login ได้ยังเรียก chatbot ได้เหมือนเดิม migration
`000037` **ต้องไม่ถูก down-migrate** ตอน Phase 2 (ดูเหตุผลในแผน migration หลัก)

## Latency baseline

Baseline เดิม (จาก `docs/chatbot-poc-plan.md` "สถานะทดสอบจริง 2026-09-02"): **~5-8 วินาที/คำถาม**
(1 LLM call + 1 SQL query ต่อ turn ผ่าน Oracle) ผู้ใช้ยอมรับได้ถึง ~30 วินาทีในบางเคสตาม
`docs/chatbot-frontend-integration.md` — ใช้เป็น non-regression target สำหรับ Phase 5.2
(คาดว่า NestJS→MCP→Postgres จะมี network hop เพิ่มจากเดิม ต้องวัดจริงว่ายังอยู่ในช่วงที่ยอมรับได้)

## New repo (Phase 0 decision — ต้องยืนยันกับผู้ใช้ก่อน Phase 3 เริ่ม)

ชื่อ/location ของ repo ใหม่ยังไม่ได้ระบุ — ผู้ใช้ต้องยืนยันก่อนเริ่ม Phase 3 (ข้อเสนอ: `lims-chatbot-service`)
