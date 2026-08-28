# Inventory — Stock Issue (เบิกออก) API

เอกสารสำหรับ frontend นำไป implement หน้า "เบิกของออกจากคลัง" (Phase 9)

รวม endpoint ที่เกี่ยวข้องกับ lot ทั้งหมดที่หน้านี้ต้องใช้ (list lots + issue)

---

## แนวคิด

- สต็อกของแต่ละ Inventory Item เก็บเป็น **Lot** หลายก้อน (`InventoryLot`)
- `quantity` ของ Item = ผลรวม `quantity` ของทุก lot (เป็นค่า derived — แก้ตรง ๆ ไม่ได้)
- การเบิกออก **ผู้ใช้เลือก lot เองเสมอ** ระบบไม่เลือก FEFO ให้อัตโนมัติ
- เบิกได้หลาย lot ในครั้งเดียว (multi-line — 1 บรรทัดต่อ 1 lot)
- ถ้าเบิกเกินยอดคงเหลือของ lot:
  - ค่าเริ่มต้น → **ไม่หักอะไรเลย** ตอบยอดคงเหลือจริงกลับมาให้ ผู้ใช้เลือกได้ว่าจะ
    1. เพิ่ม lot อีกบรรทัดเพื่อดึงส่วนที่ขาดจาก lot อื่น หรือ
    2. ยืนยันเบิกเกินจาก lot เดิม → ส่งใหม่พร้อม `force: true`
  - `force: true` → หักตามที่ขอ ยอม `quantity` ของ lot **ติดลบ** (สัญญาณว่ายอดจริงกับระบบไม่ตรง ต้องไปตรวจสอบ — ไม่ใช่ error)

---

## Auth

ทุก endpoint ต้องมี header:

```
Authorization: Bearer <access_token>
```

ต้องมีสิทธิ์:

| Endpoint | Permission |
|---|---|
| `GET /inventory/{id}/lots` | `inventory:view` |
| `POST /inventory/{id}/issue` | `inventory:edit` |

ไม่มีสิทธิ์ → `403` `{ "success": false, "error": { "code": "forbidden", ... } }`

---

## Response envelope

ทุก response ใช้รูปแบบเดียวกัน:

```jsonc
{
  "success": true,
  "data": { ... },
  "error": null
}
```

กรณี error:

```jsonc
{
  "success": false,
  "error": { "code": "validation_failed", "message": "..." }
}
```

| HTTP | `error.code` | เกิดเมื่อ |
|---|---|---|
| 400 | `validation_failed` | body ไม่ผ่าน validation (ดูกฎด้านล่าง) |
| 401 | `unauthorized` | ไม่มี/token หมดอายุ |
| 403 | `forbidden` | ไม่มี permission |
| 404 | `not_found` | ไม่พบ item หรือ lot |

> **สำคัญ:** กรณี "เบิกเกิน" **ไม่ใช่ error** — HTTP ยังเป็น `200` และ `success: true` แต่ `data.applied = false` (ดูหัวข้อ Issue Stock)

---

## 1. ดึงรายการ Lot ของ Item

```
GET /inventory/{id}/lots
```

ใช้ตอนเปิดหน้าเบิก เพื่อให้ผู้ใช้เลือก lot

### Response `200`

```jsonc
{
  "success": true,
  "data": [
    {
      "id": "LOT-00001",
      "item_id": "INV-00007",
      "lot_no": "A-2026-01",
      "expire_date": "2027-03-31T00:00:00Z",  // null ได้
      "quantity": 40
    },
    {
      "id": "LOT-00002",
      "item_id": "INV-00007",
      "lot_no": "A-2026-02",
      "expire_date": null,
      "quantity": 12
    }
  ]
}
```

- เรียงตาม `expire_date` (ใกล้หมดอายุก่อน, `null` ท้ายสุด) แล้วตามด้วย `lot_no`
- `quantity` อาจติดลบได้ ถ้าเคยมีการ force issue เกินมาก่อน
- lot ที่ `quantity = 0` ก็ยังคืนมาในลิสต์

---

## 2. เบิกของออก (Stock Issue)

```
POST /inventory/{id}/issue
```

### Request body

```jsonc
{
  "lines": [
    { "lot_id": "LOT-00001", "quantity": 10 },
    { "lot_id": "LOT-00002", "quantity": 5 }
  ],
  "force": false
}
```

| field | type | required | หมายเหตุ |
|---|---|---|---|
| `lines` | array | ✅ | อย่างน้อย 1 บรรทัด |
| `lines[].lot_id` | string | ✅ | ต้องเป็น lot ของ item `{id}` นี้ |
| `lines[].quantity` | int | ✅ | ต้อง `> 0` |
| `force` | bool | – | default `false` — `true` = ยอมให้ lot ติดลบ |

### กฎ validation (ตอบ `400 validation_failed`)

- `lines` ว่าง / ไม่มี
- `quantity <= 0` ในบรรทัดใดบรรทัดหนึ่ง
- `lot_id` ว่าง
- `lot_id` ซ้ำกันในหลายบรรทัด → ต้องรวมเป็นบรรทัดเดียว
- `lot_id` ไม่ได้เป็นของ item นี้

### กฎ not found (ตอบ `404 not_found`)

- ไม่พบ item `{id}`
- ไม่พบ lot ตาม `lot_id`

---

### กรณี A — เบิกได้ครบ (หรือส่ง `force: true`)

`quantity` ของทุกบรรทัด ≤ ยอดคงเหลือของ lot นั้น → หักทันที

**Response `200`**

```jsonc
{
  "success": true,
  "data": {
    "applied": true,
    "shortfalls": [],
    "item": {
      "id": "INV-00007",
      "name": "Pipette Tip 200µL",
      "category": "consumable",
      "quantity": 37,          // ยอดรวมใหม่หลังหัก
      "unit": "box",
      "min": 10,
      "max": 100,
      "pct": 37,
      "status": "normal",      // normal | low | out (คำนวณจาก min/max)
      "default_vendor": "...",
      "custodian_user_id": 3,
      "manufacturer": "...",
      "vendor_id": null,
      "location_id": "LOC-000123"
    },
    "lots": [
      { "id": "LOT-00001", "item_id": "INV-00007", "lot_no": "A-2026-01", "expire_date": "2027-03-31T00:00:00Z", "quantity": 30 },
      { "id": "LOT-00002", "item_id": "INV-00007", "lot_no": "A-2026-02", "expire_date": null, "quantity": 7 }
    ]
  }
}
```

- `data.lots` = เฉพาะ lot ที่ถูกแตะในครั้งนี้ (สถานะหลังหัก) — ไม่ใช่ทุก lot ของ item
- `data.item.quantity` = ยอดรวมใหม่

---

### กรณี B — เบิกเกิน และ `force` = false → ไม่หักอะไรเลย

**Response `200`** (ยังเป็น success — ไม่ใช่ error)

```jsonc
{
  "success": true,
  "data": {
    "applied": false,
    "shortfalls": [
      {
        "lot_id": "LOT-00002",
        "lot_no": "A-2026-02",
        "requested": 20,
        "available": 12
      }
    ],
    "item": { /* ItemResponse — ยอดเดิม ยังไม่ถูกแตะ */ },
    "lots": [
      /* lot ทั้งหมดที่อ้างถึงใน lines (สถานะปัจจุบัน ยังไม่หัก) */
    ]
  }
}
```

- `applied: false` → **ไม่มีการเปลี่ยนแปลงใด ๆ ใน DB**
- `shortfalls` = เฉพาะบรรทัดที่ขอเกิน — บอก `requested` (ที่ขอ) กับ `available` (คงเหลือจริง)
- บรรทัดที่ไม่เกินก็ไม่ถูกหักเช่นกัน (all-or-nothing)

**Frontend ควรทำ:** แสดง dialog สรุปว่า lot ไหนไม่พอ ขาดเท่าไหร่ แล้วให้ผู้ใช้เลือก

1. **ปรับจำนวน / เพิ่ม lot** → แก้ `lines` แล้วเรียก `POST /issue` ใหม่ (ยังคง `force: false`)
2. **ยืนยันเบิกเกิน** → เรียก `POST /issue` ซ้ำด้วย `lines` ชุดเดิม + `force: true`
   → lot จะติดลบ, `applied: true`

---

## ตัวอย่าง flow หน้าจอ

```
1. เปิดหน้าเบิก item X
   GET /inventory/X/lots           → แสดงตารางให้เลือก lot + จำนวน

2. ผู้ใช้เลือก 2 lot กรอกจำนวน กด "เบิก"
   POST /inventory/X/issue { lines: [...], force: false }

3a. data.applied == true
    → สำเร็จ แสดงยอดใหม่ (data.item.quantity), refresh ตาราง lot

3b. data.applied == false
    → popup: "LOT A-2026-02 คงเหลือ 12 แต่ขอเบิก 20"
       [ แก้จำนวน ]   [ เพิ่ม lot อื่น ]   [ ยืนยันเบิกเกิน (จะติดลบ) ]

4. ถ้ากด "ยืนยันเบิกเกิน"
   POST /inventory/X/issue { lines: <ชุดเดิม>, force: true }
   → data.applied == true, lot ติดลบ
   → แนะนำให้ frontend แสดง badge เตือนว่า lot นี้ยอดติดลบ (ต้องตรวจนับ)
```

---

## หมายเหตุเพิ่มเติม

- **ไม่มี** endpoint ยกเลิก/คืนของที่เบิกไปแล้ว — ถ้าต้องคืนให้ใช้ `POST /inventory/{id}/receive` (รับเข้า lot) ด้วย `lot_no` เดิม
- การเบิกถูกบันทึกใน audit trail อัตโนมัติ (field-diff ของ item) เฉพาะกรณี `applied: true`
- ไม่มีตาราง transaction/ledger แยก — ประวัติการเคลื่อนไหวดูจาก audit log (`GET /audit`)
- `expire_date` เป็น RFC3339 string หรือ `null`
