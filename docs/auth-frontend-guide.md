# Auth Integration Guide — สำหรับ Frontend

มาจาก grilling session (2026-08-25) ดู `CONTEXT.md` หมวด "Authentication & Sessions" สำหรับ glossary (Session, Token Family, Rolling Refresh, Absolute Session Lifetime, Reuse Detection, Refresh Cookie) และ `docs/adr/0004-refresh-token-in-httponly-cookie.md` สำหรับเหตุผลเชิงสถาปัตยกรรม

**สรุปสิ่งที่เปลี่ยนจากเดิม**: Refresh token ย้ายจาก JSON response body ไปเป็น **httpOnly cookie** ที่ backend set ให้เอง frontend **ไม่ต้องอ่านหรือเก็บ refresh token เองอีกต่อไป** — แค่ต้องแนบ cookie ไปกับ request (`credentials: 'include'`) และส่ง CSRF header ตามที่ระบุด้านล่าง ส่วน access token ยังทำงานเหมือนเดิมทุกประการ (อยู่ใน response body, แปะ `Authorization: Bearer` header เอง)

---

## 1. Endpoint ทั้งหมด

| Method | Path | Auth ที่ต้องมี | ผลลัพธ์ |
|---|---|---|---|
| POST | `/api/v1/auth/login` | ไม่ต้อง | set refresh cookie + คืน access token |
| POST | `/api/v1/auth/refresh` | refresh cookie (auto) | หมุน refresh cookie ใหม่ + คืน access token ใหม่ |
| POST | `/api/v1/auth/logout` | refresh cookie (auto) | เพิกถอน session ปัจจุบัน + ล้าง cookie |
| POST | `/api/v1/auth/logout-all` | `Authorization: Bearer <access_token>` | เพิกถอนทุก session ของ user + ล้าง cookie |

### `POST /auth/login`

Request:
```json
{ "email": "user@example.com", "password": "..." }
```

Response body:
```json
{ "data": { "access_token": "eyJ..." } }
```

**ไม่มี `refresh_token` ใน body แล้ว** — backend set คุกกี้ `refresh_token` (httpOnly, `Secure`, `SameSite=Strict`, `Path=/api/v1/auth`) ให้อัตโนมัติผ่าน `Set-Cookie` response header เบราว์เซอร์จัดการ cookie ให้เองทั้งหมด, JavaScript **อ่านค่าคุกกี้นี้ไม่ได้เลย** (httpOnly) — ไม่ต้องพยายามอ่าน ไม่ต้องเก็บเอง

### `POST /auth/refresh`

ไม่ต้องส่ง body อะไรเลยถ้าเบราว์เซอร์แนบ cookie มาด้วย (ดูข้อ 2 เรื่อง `credentials`) — backend จะหมุน refresh token ให้เองผ่าน `Set-Cookie` ใหม่ ทุกครั้งที่เรียกสำเร็จ

Response body:
```json
{ "data": { "access_token": "eyJ..." } }
```

### `POST /auth/logout`

ไม่ต้องส่ง body — backend อ่าน refresh token จาก cookie เอง เพิกถอน session นั้น แล้วล้างคุกกี้ให้

### `POST /auth/logout-all`

ต้องแปะ `Authorization: Bearer <access_token>` เหมือน endpoint อื่นๆ ที่ต้อง auth (ไม่ใช่ cookie-based) — ใช้ตอนมีปุ่ม "ออกจากระบบทุกอุปกรณ์"

---

## 2. สิ่งที่ frontend ต้องตั้งค่าใน HTTP client

### 2.1 ต้องส่ง cookie ไปกับทุก request ไปยัง `/auth/*`

**Fetch**:
```ts
fetch(`${API_BASE}/auth/refresh`, {
  method: "POST",
  credentials: "include", // จำเป็น! ไม่งั้นเบราว์เซอร์ไม่แนบ/ไม่รับคุกกี้เลย
});
```

**Axios**:
```ts
axios.create({
  baseURL: API_BASE,
  withCredentials: true, // เทียบเท่า credentials: 'include'
});
```

ถ้าลืมตั้งค่านี้ อาการที่จะเจอคือ: login สำเร็จ (ได้ access token) แต่พอเรียก `/auth/refresh` ครั้งถัดไปจะได้ 401 ตลอด เพราะเบราว์เซอร์ไม่ได้แนบคุกกี้ไปให้

### 2.2 CSRF header — บังคับสำหรับ `/auth/refresh` และ `/auth/logout`

ต้องแปะ header นี้ในทุก request ไปสอง endpoint นี้ (ไม่ต้องใส่ตอน login เพราะ login ไม่ได้พึ่ง ambient cookie):

```
X-SMLIMS-CSRF: 1
```

ไม่ต้องเป็นค่าลับหรือสุ่มอะไร — จุดประสงค์คือกัน CSRF form submission ข้าม origin (ซึ่ง set custom header ไม่ได้) ไม่ใช่กันการปลอมแปลงด้วยความลับ ถ้าไม่แปะ header นี้ backend จะตอบ `403 Forbidden` (เฉพาะตอนมีคุกกี้ refresh token แนบมาด้วยเท่านั้น — ถ้าไม่มีคุกกี้เลยจะไม่ถูกบังคับ)

**ตัวอย่าง fetch wrapper ที่ครบทั้งสองข้อ**:
```ts
async function refreshAccessToken(): Promise<string> {
  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: "POST",
    credentials: "include",
    headers: { "X-SMLIMS-CSRF": "1" },
  });
  if (!res.ok) throw new Error("session expired");
  const { data } = await res.json();
  return data.access_token;
}
```

### 2.3 อย่าเก็บ access token ใน `localStorage`

Access token ควรเก็บใน memory เท่านั้น (เช่น state ของ auth store/context) — ไม่ใช่ `localStorage`/`sessionStorage` เพราะ JS ทุกตัวบนหน้านั้นอ่านได้ (ความเสี่ยง XSS) พอ refresh หน้าเว็บ access token ในหน่วยความจำจะหายไป — flow ปกติคือเรียก `/auth/refresh` ตอนแอปโหลดครั้งแรกเพื่อขอ access token ใหม่จากคุกกี้ที่ยังอยู่ (ถ้ายังไม่หมดอายุ)

---

## 3. Flow ที่แนะนำ

### 3.1 App bootstrap (โหลดหน้าเว็บครั้งแรก/refresh หน้า)

1. เรียก `POST /auth/refresh` (พร้อม `credentials: 'include'` + CSRF header) ทันทีที่แอปโหลด
2. ถ้าสำเร็จ (200) → ได้ access token ใหม่ → user ยัง login อยู่ ไม่ต้องเด้งไปหน้า login
3. ถ้า 401 → ไม่มี session ที่ใช้ได้แล้ว (ไม่เคย login, cookie หมดอายุ, หรือถูก revoke) → เด้งไปหน้า login

### 3.2 Access token หมดอายุระหว่างใช้งาน (401 จาก API อื่น)

Access token อายุแค่ 15 นาที ดังนั้นเกือบทุก session จะเจอ 401 กลางทางแน่นอน — ควรทำ interceptor กลาง:

```ts
// ตัวอย่างแนวคิด (axios interceptor)
api.interceptors.response.use(
  (res) => res,
  async (error) => {
    if (error.response?.status === 401 && !error.config._retried) {
      error.config._retried = true;
      try {
        const newAccessToken = await refreshAccessToken();
        setAccessToken(newAccessToken);
        error.config.headers.Authorization = `Bearer ${newAccessToken}`;
        return api.request(error.config); // retry request เดิม
      } catch {
        redirectToLogin();
      }
    }
    return Promise.reject(error);
  }
);
```

ระวัง request หลายตัวพร้อมกันเจอ 401 พร้อมกัน (เช่น โหลดหน้า dashboard ที่ยิงหลาย API) — ควรกันไม่ให้ยิง `/auth/refresh` ซ้ำซ้อนหลายครั้งพร้อมกัน (เช่น cache promise เดียวไว้ใช้ร่วมกันระหว่าง request ที่ 401 พร้อมกัน)

### 3.3 Logout

```ts
await fetch(`${API_BASE}/auth/logout`, {
  method: "POST",
  credentials: "include",
  headers: { "X-SMLIMS-CSRF": "1" },
});
// ล้าง access token ออกจาก memory/state ฝั่ง frontend เอง แล้วเด้งไปหน้า login
```

### 3.4 "ออกจากระบบทุกอุปกรณ์" (ถ้ามีปุ่มนี้ใน UI)

```ts
await fetch(`${API_BASE}/auth/logout-all`, {
  method: "POST",
  credentials: "include",
  headers: { Authorization: `Bearer ${accessToken}` },
});
```

---

## 4. พฤติกรรมที่ frontend ควรรู้ (แต่ไม่ต้องจัดการเอง)

- **Rolling Refresh**: ทุกครั้งที่เรียก `/auth/refresh` สำเร็จ session จะต่ออายุอีก 7 วันอัตโนมัติ (ผ่าน cookie ใหม่) — ตราบใดที่ user active ภายใน 7 วัน session จะไม่มีวันหมดอายุจาก inactivity
- **Absolute Session Lifetime**: แต่ต่อให้ active ตลอด session จะถูกบังคับหมดอายุหลัง 30 วันนับจาก login ครั้งแรก (ต้อง login ใหม่) — จะเจอเป็น 401 ธรรมดาจาก `/auth/refresh` เหมือนกรณีอื่น ไม่มี error code พิเศษแยก
- **Reuse Detection**: ถ้า refresh cookie เก่าที่ถูกใช้ไปแล้วถูกเรียกซ้ำ (เช่นโดน replay attack) ระบบจะ revoke **ทุก session ของ user คนนั้นทันที** — ทุกอุปกรณ์ที่ login ไว้จะโดนเด้งออกพร้อมกัน เป็นพฤติกรรมที่ตั้งใจ (ป้องกัน token หลุด)

---

## 5. Local development

ถ้า frontend dev server รันคนละ port กับ backend (เช่น Vite ที่ `localhost:5173` เรียก API ที่ `localhost:8080`) เบราว์เซอร์จะมองเป็นคนละ origin ทันที ต้องขอให้ทีม backend ตั้ง env `CORS_ALLOW_ORIGINS=http://localhost:5173` (คั่นด้วย comma ถ้ามีหลาย origin) ไม่งั้น cookie จะไม่ถูกส่ง/รับข้ามพอร์ตเลย

ถ้า backend รันบน `http://` (ไม่ใช่ `https://`) ต้องขอให้ตั้ง `COOKIE_SECURE=false` ด้วย ไม่งั้นเบราว์เซอร์จะไม่ยอมรับคุกกี้ที่มี `Secure` flag บน non-https origin (ยกเว้น `localhost` ซึ่งเบราว์เซอร์สมัยใหม่บางตัวผ่อนปรนให้)

Production ที่ frontend/backend same-origin จริง ไม่ต้องตั้ง `CORS_ALLOW_ORIGINS` เลย (ปล่อยว่างไว้) เพราะ same-origin request ไม่ผ่าน CORS อยู่แล้ว
