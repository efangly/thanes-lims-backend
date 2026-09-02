# Chatbot Integration Guide — สำหรับ Frontend

POC ถาม-ตอบข้อมูลห้องแล็บด้วยภาษาธรรมชาติ (ดู `docs/chatbot-poc-plan.md`, `CONTEXT.md` หมวด "AI Chatbot (POC)")

**สรุป**: มี endpoint เดียว `POST /api/v1/chat` รับคำถามภาษาไทย/อังกฤษ **แบบ single-turn**
(ไม่มีบทสนทนาต่อเนื่อง — แต่ละครั้งเป็นอิสระ) backend เรียก Claude API สร้าง SQL รันกับ Oracle
(อ่านอย่างเดียว) แล้วสรุปคำตอบเป็นภาษาไทยกลับมา ครอบคลุมข้อมูล **Sample / TestResult /
Inventory / PurchaseOrder** เท่านั้น ใช้ JWT + RBAC เดียวกับ API อื่น

---

## 1. Endpoint

| Method | Path | Auth |
|---|---|---|
| POST | `/api/v1/chat` | `Authorization: Bearer <access_token>` + permission `chatbot:view` |

`chatbot:view` ถูก grant ให้ทุก role (Admin, Lab Manager, Scientist, QA, General) แล้ว
(migration 000037) — ปกติ user ที่ login ได้จะเรียกได้เลย

### Request

```json
{ "question": "มี sample อะไรบ้างที่ยังค้าง pending เกิน 7 วัน" }
```

| field | ชนิด | เงื่อนไข |
|---|---|---|
| `question` | string | required, 1–500 ตัวอักษร |

### Response (200)

```json
{
  "success": true,
  "data": {
    "answer": "มี 3 sample ที่ค้าง pending เกิน 7 วัน:\n\n| รหัส | ... |\n...",
    "sql_queries": [
      "SELECT id, name, status, received_at FROM samples WHERE status = 'pending' AND received_at < SYSDATE - 7"
    ],
    "rows": 3,
    "elapsed_ms": 6120,
    "cache_read_tokens": 0,
    "cache_write_tokens": 0
  }
}
```

| field | ความหมาย |
|---|---|
| `answer` | คำตอบภาษาไทย — **เป็น Markdown** (มีตาราง, **ตัวหนา**, bullet) → ควร render ด้วย markdown renderer ไม่ใช่ plain text |
| `sql_queries` | SQL ที่ backend รันจริง (เรียงตามลำดับ) — เอาไว้โชว์ใน section "ดูรายละเอียด" เพื่อความโปร่งใส ให้ผู้ใช้เห็นว่าตอบจากข้อมูลจริง |
| `rows` | จำนวนแถวรวมที่ดึงมาจาก DB |
| `elapsed_ms` | เวลาที่ backend ใช้ทั้งหมด (LLM + SQL) — ปกติ ~5,000–10,000 ms |
| `cache_read_tokens` / `cache_write_tokens` | telemetry prompt cache — ไม่ต้องแสดงให้ผู้ใช้ทั่วไป |

---

## 2. ตัวอย่างเรียกใช้ (fetch)

```ts
async function askChatbot(question: string, accessToken: string) {
  const res = await fetch("/api/v1/chat", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify({ question }),
    signal: AbortSignal.timeout(90_000), // backend อาจใช้เวลาถึง ~60-90s ในเคสที่ LLM วน query หลายรอบ
  });

  const body = await res.json();
  if (!res.ok || !body.success) {
    throw new Error(body?.error?.message ?? `chat failed (${res.status})`);
  }
  return body.data; // { answer, sql_queries, rows, elapsed_ms, ... }
}
```

> ใช้ access token แบบเดียวกับ endpoint อื่น (ดู `docs/auth-frontend-guide.md`) — ถ้าเจอ 401
> ให้ลอง `POST /auth/refresh` แล้วยิงซ้ำ ตาม flow เดิม

---

## 3. Error

รูปแบบ error เหมือนทั้งระบบ: `{ "success": false, "error": { "code": "...", "message": "..." } }`

| HTTP | `error.code` | เมื่อไร | frontend ควรทำ |
|---|---|---|---|
| 400 | `validation_failed` | `question` ว่าง / เกิน 500 ตัว | โชว์ข้อความ validate ใต้ช่องกรอก |
| 401 | `http_error` | ไม่มี token / token หมดอายุ | refresh แล้วลองใหม่ → ถ้ายังไม่ได้ ให้ logout |
| 403 | `http_error` | role ไม่มี `chatbot:view` | ซ่อนเมนู chatbot สำหรับ role นั้น |
| 404 | `http_error` | feature ปิดอยู่ (ADB ต่อไม่ได้ตอน backend start) | โชว์ "ระบบผู้ช่วยไม่พร้อมใช้งานชั่วคราว" + ปิดปุ่มส่ง |
| 500 | `internal_error` | LLM error / query ผิดซ้ำจนเกิน limit / DB error | โชว์ "ตอบไม่สำเร็จ ลองถามใหม่อีกครั้ง" |

> **404 = feature ทั้งก้อนไม่พร้อม** (route ไม่ถูก mount) ไม่ใช่ "ไม่พบข้อมูล"
> แนะนำให้ frontend probe ครั้งแรกด้วยคำถามสั้น ๆ หรือเช็คจาก config flag ฝั่ง frontend เอง
> แล้วซ่อน UI chatbot ทั้งหมดถ้าได้ 404

---

## 4. หมายเหตุ UX

- **Latency สูง (5–10 วิ ปกติ, บางเคสถึง ~30 วิ)** — ต้องมี loading state ชัดเจน, disable ปุ่มส่งระหว่างรอ,
  อย่าให้กดซ้ำ (แต่ละ request เสียเงินเรียก LLM)
- **Single-turn** — ไม่มี memory ข้ามคำถาม ถ้าจะทำ UI แชท ต้องส่ง context ที่ต้องการไปในคำถามเองทุกครั้ง
  (เช่น "จากรายการเมื่อกี้..." จะไม่เข้าใจ)
- **`answer` เป็น Markdown** — ใช้ lib เช่น `react-markdown` / `marked` และรองรับ GFM table
- **โชว์ `sql_queries`** ใน accordion "ดู SQL ที่ใช้" ช่วยให้ stakeholder เชื่อว่าตอบจากข้อมูลจริง (จุดขายของ demo)
- **ขอบเขตคำถาม** — ตอบได้เฉพาะ Sample, TestResult, Inventory, PurchaseOrder
  ถ้าถามนอกขอบเขต (เช่น เรื่อง Equipment, ผู้ใช้, การตั้งค่า) โมเดลจะบอกว่าตอบไม่ได้ — เป็นพฤติกรรมที่ตั้งใจ
- **อ่านอย่างเดียว** — chatbot แก้/ลบข้อมูลไม่ได้ (มี guard 2 ชั้น) ถ้าผู้ใช้สั่งให้ลบ/แก้ จะถูกปฏิเสธ

---

## 5. คำถามตัวอย่างสำหรับ demo (จาก seed data)

| คำถาม | ได้อะไร |
|---|---|
| มี sample อะไรบ้างที่ยังค้างสถานะ pending เกิน 7 วัน | รายการ SMP-xxxx + จำนวนวันที่ค้าง |
| test result อะไรบ้างที่ flag เป็น hi หรือ lo | ~15 รายการผลผิดปกติ |
| สารเคมี/วัสดุคงคลังอะไรบ้างที่ต่ำกว่าจุดสั่งซื้อขั้นต่ำ | INV-xxxx ที่ quantity < min_qty |
| มีใบสั่งซื้อที่ยังรออนุมัติหรือส่งให้ vendor แล้วกี่ใบ | นับ + รายการ PO ตามสถานะ |
| ตัวอย่างของ "วิภา สายใจ" มีอะไรบ้าง และสถานะเป็นอย่างไร | รายการ sample ตาม custodian |
