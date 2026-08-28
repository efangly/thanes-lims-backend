# Storage Location API — Frontend Integration Guide

เอกสารนี้สรุป API และ business rules ของฟีเจอร์ "Storage Location" (ตู้/ชั้น/ช่อง/sub-ช่อง) ที่เพิ่งเพิ่มเข้ามาแทนที่ `Sample.Location` แบบ text เดิม สำหรับทีม frontend เอาไปสร้างหน้าจัดการตำแหน่งจัดเก็บตัวอย่าง

อ้างอิงเพิ่มเติม: `CONTEXT.md#storage-location` (glossary), `docs/adr/0001-self-referencing-tree-for-storage-location.md` (เหตุผลการออกแบบ), Swagger ที่ `GET /api/v1/swagger/*` (ทุก endpoint มี annotation ครบ รวมถึง schema แบบเต็ม)

## แนวคิดหลัก

Location เป็น**ต้นไม้** (tree) 4 ระดับตายตัว ไล่จากบนลงล่าง:

```
cabinet (ตู้)  →  shelf (ชั้น)  →  slot (ช่อง)  →  sub_slot (sub-ช่อง)
```

- แต่ละ node มี `level_type` กำกับ (`cabinet` / `shelf` / `slot` / `sub_slot`) — ห้ามข้ามระดับ (เช่น cabinet มีลูกเป็น slot ตรงๆ ไม่ได้)
- **node ระดับกลางเป็น "จุดจัดเก็บจริง" (leaf) ได้เอง** ถ้าไม่ถูกแบ่งย่อยต่อ — เช่น ตู้เย็นเล็กที่ไม่มีชั้นย่อยเลย ก็ผูก sample เข้าที่ตู้นั้นได้ตรงๆ นิยาม "leaf" คือ **node ที่ไม่มีลูก** ไม่ใช่ node ที่ level_type เป็น sub_slot เท่านั้น
- **Sample ผูกได้เฉพาะ leaf node เท่านั้น** และ 1 leaf ผูกกับ sample ที่ยัง "active" ได้แค่ 1 ตัวในเวลาเดียวกัน (ดู [กติกา active/leaf](#กติกา-active-กับ-leaf) ด้านล่าง)
- **Full Path** (เช่น `"Fridge-A / Shelf-2 / Slot-4"`) คำนวณจาก tree สดทุกครั้งที่เรียก ไม่ได้เก็บไว้เป็น column แยก — ถ้าแก้ชื่อ node กลางทาง full path ของลูกทุกตัวจะเปลี่ยนตามทันที

## Data model

```jsonc
// LocationResponse
{
  "id": "LOC-00001",
  "parent_id": null,          // null = เป็น root (cabinet)
  "name": "Fridge-A",
  "level_type": "cabinet"     // "cabinet" | "shelf" | "slot" | "sub_slot"
}
```

`id` เป็น human-readable string รูปแบบ `LOC-00001` (auto-generate, ไม่ต้องส่งตอนสร้าง)

## Endpoints

ทุก endpoint อยู่ใต้ `/api/v1` และต้องแนบ `Authorization: Bearer <token>` เหมือน endpoint อื่นในระบบ ตอบกลับเป็น envelope มาตรฐาน `{ success, data, error: { code, message } }`

### สร้างตู้ใหม่ (root)

```
POST /locations
Body: { "name": "Fridge-A" }
→ 201 { data: LocationResponse }
```

- ชื่อตู้ต้อง **unique ทั้งระบบ** (ไม่ใช่แค่ในกลุ่มพี่น้อง) — ซ้ำแล้วได้ `409 conflict`
- สร้างได้แค่ระดับ `cabinet` เท่านั้นผ่าน endpoint นี้ (root เสมอ ไม่มี `parent_id`)

### สร้างลูกอัตโนมัติ (generate children)

```
POST /locations/:id/children
Body: { "prefix": "Shelf", "count": 5 }
→ 201 { data: [LocationResponse, ...] }
```

- สร้างลูกของ `:id` จำนวน `count` ตัว ชื่อ `"{prefix}-1"` .. `"{prefix}-{count}"` เรียงตามเลข
- `level_type` ของลูกถูกกำหนดอัตโนมัติเป็น **ระดับถัดไป** จาก parent (parent เป็น `cabinet` → ลูกเป็น `shelf` เสมอ ฯลฯ) — ไม่ต้องส่ง `level_type` มาเอง และส่งมาก็ไม่มีผลเพราะ endpoint ไม่รับ field นี้
- endpoint เดียวใช้ซ้ำได้ทุกระดับ (เรียกกับ cabinet ก็ได้ shelf, เรียกกับ shelf ก็ได้ slot, เรียกกับ slot ก็ได้ sub_slot)
- เรียกกับ `sub_slot` (ระดับลึกสุด) จะได้ `400 validation_failed` เพราะแบ่งย่อยต่อไม่ได้แล้ว
- ถ้าชื่อที่จะ generate ไปชนกับ sibling ที่มีอยู่แล้ว (เช่น generate "Shelf-1".."Shelf-5" ซ้ำรอบสอง) จะได้ `409 conflict` และ **ไม่สร้างอะไรเลยทั้ง batch** (all-or-nothing)

### ดูรายการลูกโดยตรง

```
GET /locations?parent_id=LOC-00001
→ 200 { data: [LocationResponse, ...] }

GET /locations                      // ไม่ระบุ parent_id = list ตู้ (root) ทั้งหมด
→ 200 { data: [LocationResponse, ...] }
```

- คืนเฉพาะ**ลูกโดยตรง** (1 ชั้น) ไม่ recursive — หน้า UI ที่จะโชว์ tree ต้อง drill-down เรียกทีละระดับเอง (เช่น เรียก `?parent_id=` ของ cabinet ที่เลือก เพื่อได้ shelf, แล้วเรียกอีกครั้งด้วย shelf ที่เลือกเพื่อได้ slot)
- เรียงผลลัพธ์ตามชื่อ (alphabetical) — ถ้าตั้งชื่อผ่าน generate-children (`Shelf-1`..`Shelf-10`) การเรียงแบบ string จะทำให้ `Shelf-10` มาก่อน `Shelf-2` (lexicographic ไม่ใช่ natural sort) — ฝั่ง frontend อาจต้อง sort เพิ่มเองถ้าต้องการลำดับเลขที่ถูกต้องเมื่อจำนวนเกิน 9

### ดู Full Path

```
GET /locations/:id/full-path
→ 200 { data: { "full_path": "Fridge-A / Shelf-2 / Slot-4" } }
```

- ใช้แสดงผลตำแหน่งแบบอ่านง่ายในหน้ารายการ sample / ใบรายงาน — ไม่ต้อง fetch ancestor ทีละ node เอง

### ลบ Location

```
DELETE /locations/:id
→ 204 (no content)
```

- **ลบไม่ได้ถ้ายังมีลูกอยู่** หรือ **มี sample (ไม่ว่า status ไหน) อ้างอิงอยู่** → `409 conflict` — ต้องย้าย/ลบ sample หรือลูกออกให้หมดก่อน ระบบไม่มี cascade delete ให้
- ลบสำเร็จได้ 204 ไม่มี response body

### ผูก/ย้าย Sample เข้า Location (put-away)

```
PATCH /samples/:id/location
Body: { "location_id": "LOC-00042" }
→ 200 { data: SampleResponse }
```

- แยกจากการสร้าง sample โดยเจตนา — workflow จริงคือ **รับตัวอย่างเข้าระบบก่อน (POST /samples) แล้วค่อยเอาไปเก็บที่ location จริงทีหลัง** (put-away) ซึ่งอาจมีช่วงเวลาคั่นกลางที่ยังไม่ได้ assign ตำแหน่ง
- ใช้ endpoint นี้ทั้งตอน**ผูกครั้งแรก**และตอน**ย้าย location** (ทับ location เดิมได้เลย ไม่ต้อง unassign ก่อน)
- Error ที่ต้องดักในหน้า UI:
  - `400 validation_failed` — location ที่เลือกไม่ใช่ leaf (ยังมีลูกอยู่) เช่น user เผลอเลือก cabinet/shelf ที่มี slot ย่อยอยู่จริง แทนที่จะเลือก slot
  - `404 not_found` — sample หรือ location id ไม่มีจริง
  - `409 conflict` — location ที่เลือกมี sample อื่นครองอยู่แล้ว (active)

## กติกา "active" กับ "leaf"

- **leaf** = node ที่ไม่มีลูก (ไม่ใช่ field ที่ backend ส่งมาให้เช็คตรงๆ — ต้องดูจากการที่ `GET /locations?parent_id=X` คืน list ว่างเปล่า ถึงจะรู้ว่า X เป็น leaf) ถ้า UI ต้องเช็คบ่อยควรพิจารณา cache ผลไว้ฝั่ง frontend หลัง fetch children รอบแรก
- **active sample** = sample ที่ status เป็น `pending`, `testing`, หรือ `completed` — สถานะเหล่านี้ถือว่า "ยังอยู่ที่เก็บจริง" สถานะ `transferred` (ย้ายออกไปแผนกอื่น) ถือว่า**ปล่อย location คืนแล้ว** ดังนั้น:
  - leaf ที่มี sample สถานะ `transferred` อยู่ **ยังผูก sample ใหม่เข้าไปทับได้** (ไม่ต้อง unassign ก่อน เพราะ backend มองว่าว่างแล้ว)
  - leaf ที่มี sample สถานะ `pending`/`testing`/`completed` อยู่ ผูก sample อื่นเข้าไปซ้ำไม่ได้จนกว่า sample เดิมจะเปลี่ยนเป็น `transferred`

## สร้าง Sample พร้อม location เลยได้ไหม

`POST /samples` ยังรับ field `location_id` (optional, nullable) อยู่ในตัว request แต่ **ปัจจุบัน endpoint นี้ไม่ validate leaf/active-uniqueness ให้** (validation เต็มรูปแบบมีเฉพาะที่ `PATCH /samples/:id/location`) — **แนะนำให้ frontend สร้าง sample โดยไม่ส่ง `location_id` ก่อน แล้วค่อยเรียก `PATCH /samples/:id/location` แยกต่างหากเสมอ** เพื่อให้ผ่าน validation ครบถ้วน จุดนี้เป็น known gap ฝั่ง backend ที่อาจถูกปิดในอนาคต (ตอนนี้ยังไม่ได้ทำ)

## ตัวอย่าง workflow เต็ม (สร้างตู้ใหม่ + ผูก sample)

```
1. POST /locations                         { name: "Fridge-B" }
   → LOC-00050 (cabinet)

2. POST /locations/LOC-00050/children       { prefix: "Shelf", count: 2 }
   → [LOC-00051 "Shelf-1", LOC-00052 "Shelf-2"]

3. POST /locations/LOC-00051/children       { prefix: "Slot", count: 10 }
   → [LOC-00053 "Slot-1", ..., LOC-00062 "Slot-10"]

4. POST /samples                            { name: "...", type: "blood", custodian_user_id: 3 }
   → SMP-2569-00099 (location_id: null)

5. PATCH /samples/SMP-2569-00099/location   { location_id: "LOC-00053" }
   → SMP-2569-00099 (location_id: "LOC-00053")

6. GET /locations/LOC-00053/full-path
   → { full_path: "Fridge-B / Shelf-1 / Slot-1" }
```

## Error codes ที่เกี่ยวข้อง (ทุก endpoint ในเอกสารนี้)

| HTTP | `error.code`        | เมื่อไหร่ |
|------|----------------------|-----------|
| 400  | `validation_failed`  | field ไม่ครบ/ผิด format, generate children กับ sub_slot, assign sample เข้า non-leaf |
| 401  | `unauthorized`       | ไม่ได้ login / token หมดอายุ |
| 404  | `not_found`          | location/sample id ไม่มีจริง |
| 409  | `conflict`           | ชื่อตู้ซ้ำ, sibling ชื่อซ้ำตอน generate children, ลบ location ที่ยังมีลูก/sample, assign เข้า leaf ที่ active sample ครองอยู่แล้ว |
