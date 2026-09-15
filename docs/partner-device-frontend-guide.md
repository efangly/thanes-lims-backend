# Partner Device API — Frontend Integration Guide

เอกสารนี้สรุป API ของฟีเจอร์ "Partner Device" (อุปกรณ์วัดอุณหภูมิ/ความชื้นของ SMtrack ที่ผูกเข้ากับ Location ในระบบเรา) สำหรับทีม frontend เอาไปสร้างหน้าจัดการและแสดงผล — **นี่คือ API ฝั่ง lab/backend ที่ frontend เรียก ไม่ใช่ API ที่ backend เรียกไป SMtrack** (อันนั้นดูที่ `docs/partner-api-guide.md`)

อ้างอิงเพิ่มเติม: `CONTEXT.md#environment` (glossary: Partner Device, Gauge, Env Alert, Ward), `docs/adr/0011-partner-device-no-persist-proxy.md` (ทำไมไม่เก็บ raw reading), `docs/adr/0012-partner-api-grpc-transport.md` (ทำไมเปลี่ยนเป็น gRPC ฝั่ง backend↔SMtrack — ไม่กระทบ endpoint ที่ frontend เรียก), Swagger ที่ `GET /api/v1/swagger/*`

## แนวคิดหลัก

- **Partner Device** = การ**ผูก** (`Serial` ↔ `Location`) ระหว่างอุปกรณ์จริงของ SMtrack กับ Location ที่มี Gauge (threshold) อยู่แล้วในระบบเรา — ไม่ใช่ตัวอุปกรณ์เอง
- Backend มี background job **poll ทุก ๆ 30 วินาที** (ตั้งค่าได้ฝั่ง backend) ดึงค่าจาก SMtrack, ประเมินกับ threshold ของ Gauge, cache ผลไว้ที่ Redis (TTL สั้น ๆ) แล้ว broadcast ผ่าน SSE — **frontend ไม่เคยเรียก SMtrack ตรง ๆ**
- ค่าที่อ่านได้ (`temp_display`, `humidity_display`, ...) เป็น**ค่าปัจจุบัน/ล่าสุดเท่านั้น** — **ไม่มีการเก็บ history ระยะยาวในระบบเรา** (ดู ADR 0011/0012) ถ้าต้องการกราฟย้อนหลัง ตอนนี้ backend ยังไม่มี endpoint ให้ (SMtrack เองก็ให้แค่ย้อนหลัง 1 ชั่วโมงต่อการเรียกหนึ่งครั้งหลังเปลี่ยนมาใช้ gRPC)
- **Ward เป็นแนวคิดของฝั่ง SMtrack เท่านั้น** (เช่น ICU, OPD) ไม่ใช่ `Location` ของเรา ห้ามเอามาปนกัน — ward ใช้แค่ตอนค้นหาอุปกรณ์ (ดู endpoint `/discover` ด้านล่าง) ไม่ถูกเก็บไว้ที่ไหนในระบบเรา

## Feature flag ที่ frontend ต้องรู้

ทั้งฟีเจอร์นี้ปิดเปิดได้ทั้งชุดฝั่ง backend (`PARTNER_API_ENABLED`) แยกเป็น 2 กลุ่ม endpoint:

| เปิดเสมอ (ไม่ผูกกับ flag) | เปิดเฉพาะเมื่อ `PARTNER_API_ENABLED=true` |
|---|---|
| `POST /partner-devices`, `GET /partner-devices`, `GET /partner-devices/:serial`, `PATCH /partner-devices/:serial` (CRUD การผูก mapping — เก็บใน Postgres ล้วน ๆ) | `GET /partner-devices/discover` (ต้องเรียก SMtrack จริง) |
| | `GET /partner-devices/:serial/snapshot`, `GET /partner-devices/stream` (มีข้อมูลก็ต่อเมื่อ poller ทำงานอยู่) |

**เมื่อ flag ปิด**: `GET /discover` จะได้ `404 not_found` เพราะ route ไม่ถูกลงทะเบียนเลย (ไม่ใช่ 503) ส่วน `snapshot`/`stream` ยังเรียกได้ปกติแต่จะไม่มีข้อมูลใหม่ ๆ เข้ามา (cache จะว่างเปล่าหรือหยุดอัปเดต) — ถ้า UI ต้องซ่อนปุ่ม/เมนูที่เกี่ยวกับ live data เมื่อ flag ปิด ให้ backend ทีมช่วยยืนยันสถานะ flag แยกต่างหาก (ปัจจุบันยังไม่มี endpoint บอกสถานะ flag ตรง ๆ ให้ frontend เช็ค)

## Permission ที่ต้องมี (RBAC module: `partnerdevice`)

| Action | endpoint ที่ใช้ |
|---|---|
| `partnerdevice:view` | `GET /partner-devices`, `GET /partner-devices/:serial`, `GET /partner-devices/:serial/snapshot`, `GET /partner-devices/stream`, `GET /partner-devices/discover` |
| `partnerdevice:create` | `POST /partner-devices` |
| `partnerdevice:edit` | `PATCH /partner-devices/:serial` |

ทุก endpoint อยู่ใต้ `/api/v1` และต้องแนบ `Authorization: Bearer <token>` ตอบกลับเป็น envelope มาตรฐาน `{ success, data, error: { code, message } }`

## Data model

```jsonc
// PartnerDeviceResponse — การผูก mapping (CRUD)
{
  "serial": "eTPV2-2P-L0168-1068-055",
  "location": "Fridge-A",   // ต้องเป็น Location ที่มี Gauge อยู่แล้ว
  "active": true            // false = หยุด poll ชั่วคราว โดยไม่ลบ mapping
}
```

```jsonc
// SnapshotResponse — ค่าล่าสุดที่ poller เก็บไว้ (จาก cache เท่านั้น)
{
  "serial": "eTPV2-2P-L0168-1068-055",
  "location": "Fridge-A",
  "name": "eTPV2-2P-055",
  "status": true,
  "firmware": "1.0.0",
  "online": true,
  "temp_display": 23.55,
  "humidity_display": 61.61,
  "send_time": "2026-09-15T08:50:00Z",  // เวลาที่ SMtrack ส่งค่านี้มา (ไม่ใช่เวลาที่ poll)
  "level": "crit",                      // "ok" | "warn" | "crit" ตาม threshold ของ Gauge
  "fetched_at": "2026-09-15T16:19:16+07:00", // เวลาที่ backend poll สำเร็จล่าสุด
  "stale": false                        // true = poll ล้มเหลวชั่วคราว, ค่านี้เป็นของเก่าที่ fallback มาโชว์
}
```

```jsonc
// DiscoverDevicesResponse — ผลค้นหาอุปกรณ์จาก SMtrack ตาม ward (ใช้ก่อนสร้าง mapping)
{
  "devices": [
    {
      "serial": "eTPV2-2P-L0168-1068-055",
      "name": "eTPV2-2P-055",
      "status": true,
      "firmware": "1.0.0",
      "online": true,
      "temp_display": 23.55,      // omit ทั้ง 3 field นี้ถ้าอุปกรณ์ยังไม่เคยส่งค่าเข้ามาเลย
      "humidity_display": 61.61,
      "send_time": "2026-09-15T08:50:00Z"
    },
    { "serial": "eTEP-test", "name": "Test SM", "status": true, "firmware": "1.0.0", "online": true }
  ],
  "total": 4,
  "page": 1,
  "limit": 5
}
```

**สำคัญ**: `temp_display`/`humidity_display`/`send_time` ใน `DiscoverDevicesResponse` เป็น **nullable/omit ได้** — ต้องเช็ค `!= null` ก่อนโชว์ ห้ามสมมติว่าเป็น `0` เพราะ "ไม่มีค่า" กับ "ค่าจริงคือ 0" ต่างกัน (ต่างจาก `SnapshotResponse` ที่ field พวกนี้ไม่ nullable เพราะ snapshot มีก็ต่อเมื่อ poll สำเร็จแล้ว)

## Endpoints

### ค้นหาอุปกรณ์ SMtrack ตาม ward (ช่วยหา Serial ก่อนสร้าง mapping)

```
GET /partner-devices/discover?ward={ward}&page=1&limit=20
→ 200 { data: DiscoverDevicesResponse }
```

- `ward` **required** — เป็นค่าที่ทีม SMtrack/admin ให้มา ไม่ใช่ `Location` ของเรา (ดู "แนวคิดหลัก" ด้านบน) — ไม่ส่งมาได้ `400 validation_failed`
- `page` default 1, `limit` default 20 (backend clamp ให้ไม่เกิน 100 อัตโนมัติ ไม่ต้อง validate ฝั่ง frontend)
- ward ที่ API key ของระบบไม่มีสิทธิ์เห็น (หรือไม่มีจริง) → `404 not_found` — **จงใจแยกไม่ออก**จาก "ward ไม่มีจริง" (เป็น design ของฝั่ง SMtrack เอง เพื่อไม่ให้เดา ward ที่มีอยู่ได้)
- ใช้เป็นขั้นตอนแรกก่อนสร้าง mapping จริง (ดู workflow ด้านล่าง) — ผลลัพธ์จาก endpoint นี้**ไม่ถูกบันทึกที่ไหนเลย** เป็นแค่การอ่านสด ๆ จาก SMtrack

### สร้าง mapping (ผูก Serial เข้ากับ Location)

```
POST /partner-devices
Body: { "serial": "eTPV2-2P-L0168-1068-055", "location": "Fridge-A", "active": true }
→ 201 { data: PartnerDeviceResponse }
```

- `location` **ต้องเป็น Location ที่มี Gauge อยู่แล้ว** — ไม่มี Gauge ที่ location นั้น → `404 not_found` (ระบบไม่สร้าง Gauge ให้อัตโนมัติ ต้องไปสร้าง Gauge ก่อนที่หน้า environment)
- `serial` ซ้ำกับที่ผูกไว้แล้ว → `409 conflict`
- **backend ไม่ validate ว่า `serial` มีอยู่จริงบน SMtrack ตอนสร้าง** — สร้าง mapping ด้วย serial ผิด ๆ ได้ (จะไปพังตอน poll แทน คือ snapshot จะไม่เคยมีข้อมูล/error) แนะนำให้ frontend ใช้ `/discover` เพื่อเลือก serial ที่มีอยู่จริงเสมอ แทนการให้ user พิมพ์เอง

### รายการ mapping ทั้งหมด / ดูตัวเดียว

```
GET /partner-devices                → 200 { data: [PartnerDeviceResponse, ...] }
GET /partner-devices/:serial        → 200 { data: PartnerDeviceResponse }
```

### แก้ไข mapping (เปลี่ยน Location หรือปิด/เปิด polling)

```
PATCH /partner-devices/:serial
Body: { "location": "Fridge-A", "active": false }
→ 200 { data: PartnerDeviceResponse }
```

- **ไม่มี endpoint ลบ** — ปิดการ poll ชั่วคราวด้วย `active: false` เท่านั้น (mapping ยังอยู่ในระบบเสมอ) ถ้า user ต้องการ "เอาออกจริง ๆ" ให้สื่อสารกับทีม backend เพิ่มเติม
- **ไม่ใช่ partial update** — endpoint นี้แทนที่ทั้ง `location`/`active` ทุกครั้ง ไม่ใช่ merge field ที่ส่งมาเข้ากับของเดิม `location` เป็น required field (ไม่ส่ง → `400`) แต่ **`active` ไม่มี validation บังคับ ถ้าไม่ส่ง field นี้มาเลยจะถูกตีความเป็น `false` โดยอัตโนมัติ** (ค่า default ของ boolean) — เท่ากับปิดการ poll โดยไม่ตั้งใจ **ต้อง GET ค่าปัจจุบันมาก่อนแล้วส่ง `active` เดิมกลับไปเสมอ ถ้าจะแก้แค่ `location`** อย่าคิดว่า field ที่ไม่ส่งจะถูก "คงค่าเดิม" ให้

### ดูค่าล่าสุด (จาก cache)

```
GET /partner-devices/:serial/snapshot
→ 200 { data: SnapshotResponse }
```

- อ่านจาก cache ที่ background poller เก็บไว้เท่านั้น **ไม่เรียก SMtrack ตรง ๆ ตอนนี้** — ถ้าเพิ่งสร้าง mapping ใหม่ ต้องรอรอบ poll ถัดไปก่อน (ปกติภายใน 30 วินาที) ถึงจะมีข้อมูล
- ยังไม่เคย poll สำเร็จเลยสักครั้ง (เช่น เพิ่งสร้าง mapping, หรือ `active: false` มาตลอด) → `404 not_found`
- `stale: true` = poll ล้มเหลวแบบชั่วคราว (rate limit / เครือข่ายมีปัญหาฝั่ง SMtrack) แต่ backend ยังมีค่าเก่าคืนให้แทนการ error ทันที — **แนะนำให้ frontend โชว์ badge "ข้อมูลอาจไม่ล่าสุด" เมื่อ `stale === true`** แทนการซ่อนค่าไปเลย

### Live update ผ่าน SSE

```
GET /partner-devices/stream    (text/event-stream)
```

ส่ง event ใหม่ทุกครั้งที่ poller ได้ข้อมูลใหม่ (ของ**ทุก device ที่ active** รวมกันในสตรีมเดียว ไม่แยก stream ต่อ serial) แต่ละ event เป็น JSON ตรงกับ `SnapshotResponse` — ใช้ `s.serial` แยกว่าเป็นของอุปกรณ์ไหนฝั่ง client เอง

**ข้อควรระวัง (สำคัญ)**: endpoint นี้**ใช้ native `EventSource` ไม่ได้** เพราะต้องแนบ `Authorization: Bearer <token>` ซึ่ง `EventSource` ตั้ง custom header ไม่ได้ — ต้องอ่านผ่าน `fetch()` + `ReadableStream` แทน (เหมือน endpoint SSE อื่น ๆ ในระบบนี้) ตัวอย่าง:

```js
const res = await fetch('/api/v1/partner-devices/stream', {
  headers: { Authorization: `Bearer ${token}` },
});
const reader = res.body.getReader();
const decoder = new TextDecoder();
let buf = '';
while (true) {
  const { value, done } = await reader.read();
  if (done) break;
  buf += decoder.decode(value, { stream: true });
  let idx;
  while ((idx = buf.indexOf('\n\n')) !== -1) {
    const chunk = buf.slice(0, idx);
    buf = buf.slice(idx + 2);
    if (chunk.startsWith('data: ')) {
      const snapshot = JSON.parse(chunk.slice(6));
      // อัปเดต UI ตาม snapshot.serial
    }
  }
}
```

## Workflow เต็ม (ผูกอุปกรณ์ใหม่ + แสดงผล)

```
1. GET /partner-devices/discover?ward=bbd2930b-fb7c-4038-aa1c-73216aabd4bb
   → เลือก serial "eTPV2-2P-L0168-1068-055" จากรายการที่คืนมา

2. (ถ้ายังไม่มี Gauge) ไปสร้าง Gauge ที่ Location ปลายทางก่อน ผ่านหน้า environment/gauges

3. POST /partner-devices    { serial: "eTPV2-2P-L0168-1068-055", location: "Fridge-A", active: true }
   → mapping ถูกสร้าง แต่ยังไม่มี snapshot

4. รอ ~30 วินาที (รอบ poll ถัดไป) แล้ว
   GET /partner-devices/eTPV2-2P-L0168-1068-055/snapshot
   → ได้ค่าล่าสุด + level

5. เปิด GET /partner-devices/stream ค้างไว้ในหน้าจอ
   → รับ update อัตโนมัติทุกรอบ poll โดยไม่ต้อง poll ฝั่ง frontend เอง (polling ฝั่ง client ก็ทำได้ ถ้าไม่อยากใช้ SSE — เรียก snapshot ซ้ำทุก ๆ 30+ วินาทีก็พอ)
```

## Error codes ที่เกี่ยวข้อง (ทุก endpoint ในเอกสารนี้)

| HTTP | `error.code` | เมื่อไหร่ |
|------|---|---|
| 400 | `validation_failed` | `serial`/`location` ว่าง, `ward` ว่างตอน discover |
| 401 | `unauthorized` | ไม่ได้ login / token หมดอายุ |
| 404 | `not_found` | serial mapping ไม่มีจริง, location ไม่มี Gauge, snapshot ยังไม่เคย poll สำเร็จ, ward ไม่มี/ไม่มีสิทธิ์เห็น, **หรือ route `/discover` ไม่ถูกลงทะเบียนเพราะ flag ปิด** |
| 409 | `conflict` | สร้าง mapping ด้วย serial ที่ผูกไว้แล้ว |
