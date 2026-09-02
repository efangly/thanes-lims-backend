# เปิดทางเน็ตเวิร์กให้ OKE ต่อ Oracle ADB (chatbot POC)

Pod `thanes-lims-api` บน OKE ต้องต่อไปยัง POC Oracle ADB (`limsdb_high`, service `LIMSDB`)
เพื่อรัน SQL ของ chatbot ตอนนี้ ADB ปฏิเสธการต่อด้วย:

```
ORA-12506: TNS:listener rejected connection based on service ACL filtering
```

แปลว่า **public IP ต้นทางของ pod ยังไม่อยู่ใน Access Control List (ACL) ของ ADB**
เอกสารนี้อธิบายวิธีเปิด — มี 2 แบบตาม endpoint ของ ADB

> ระบบ auth (mTLS wallet) จัดการแล้ว: secret `adb-wallet` mount ที่ `/app/wallet`,
> `ORACLE_TNS_ADMIN=/app/wallet` ตั้งใน ConfigMap แล้ว. เอกสารนี้คือชั้น network เท่านั้น

---

## แบบ A — ADB มี Public endpoint + ACL (กรณีปัจจุบัน)

ADB เปิด public แต่กรองด้วย IP allowlist. ต้องเพิ่ม **egress public IP ของ OKE**

### 1. หา egress public IP ของ pod

Worker node ของ OKE อยู่ใน private subnet → ออกเน็ตผ่าน **NAT Gateway** ของ VCN
IP ที่ ADB เห็นคือ public IP ของ NAT Gateway นั้น

**วิธีที่ 1 — ถามจากใน pod โดยตรง (แม่นสุด):**
```sh
kubectl -n thanes-lims run egress-check --rm -it --restart=Never \
  --image=curlimages/curl -- curl -s https://ifconfig.me
# => x.x.x.x  <- นี่คือ IP ที่ต้องใส่ใน ACL
```

**วิธีที่ 2 — OCI Console:**
Networking → Virtual Cloud Networks → *(VCN ของ OKE cluster)* → NAT Gateways →
คอลัมน์ **Public IP address**

> ถ้ามี NAT Gateway หลายตัว / node หลาย subnet ให้รันวิธีที่ 1 ซ้ำ 2-3 ครั้ง
> (pod อาจถูก schedule คนละ node) หรือใส่ทุก NAT GW IP

### 2. เพิ่ม IP ลง ADB Access Control List

OCI Console → **Oracle Database → Autonomous Database → `LIMSDB`** → แท็บ/ปุ่ม
**Network** → **Access control list** → **Edit**

- **IP notation type**: `IP address` (หรือ `CIDR block` ถ้าอยากเผื่อทั้ง block)
- **Values**: `<NAT_GW_PUBLIC_IP>` (เช่น `203.0.113.45`) — จะเติม `/32` ให้เอง
- กด **Save changes**

ADB จะเข้าสถานะ **Updating** ~1-3 นาที แล้วกลับเป็น **Available**

CLI (ทางเลือก):
```sh
oci db autonomous-database update --autonomous-database-id <ADB_OCID> \
  --whitelisted-ips '["<EXISTING_IP_1>","<EXISTING_IP_2>","<NAT_GW_PUBLIC_IP>"]'
# NOTE: --whitelisted-ips เป็นการ "แทนที่ทั้งลิสต์" ต้องใส่ IP เดิมที่มีอยู่มาด้วยเสมอ
```

### 3. ตรวจสอบ

```sh
kubectl -n thanes-lims rollout restart deploy/thanes-lims-api
kubectl -n thanes-lims logs -l app=thanes-lims-api --tail=20 | grep -i oracle
# ต้องเห็น: "oracle: connected to ADB"
# ไม่ใช่:  "chatbot: oracle connect failed: ... ORA-12506"
```

หรือยิง endpoint จริง (ดู `docs/chatbot-frontend-integration.md`):
```sh
curl -sS https://lims.siamatic.work/api/v1/chat \
  -H "Authorization: Bearer <access_token>" -H "Content-Type: application/json" \
  -d '{"question":"มี sample กี่รายการ"}'
```

---

## แบบ B — ADB Private endpoint (ถ้าย้ายไป private ในอนาคต)

ถ้า ADB ใช้ **private endpoint** ในบาง VCN แทน public ACL — ไม่ใช้ whitelist IP แล้ว
แต่คุมด้วย **Security List / NSG** ของ subnet และต้องมีเส้นทางระหว่าง VCN

### ถ้า ADB กับ OKE อยู่ VCN เดียวกัน

1. **NSG ของ ADB private endpoint** — เพิ่ม Ingress rule:
   - Source: CIDR ของ worker node subnet (เช่น `10.0.1.0/24`) หรือ NSG ของ node pool
   - Protocol: TCP, Destination port: **1522** (ADB mTLS listener)
2. **Security List / NSG ของ worker subnet** — เพิ่ม Egress rule:
   - Destination: CIDR/NSG ของ ADB private endpoint
   - Protocol: TCP, port **1522**

### ถ้าคนละ VCN

เพิ่มจากข้างบน:
3. **Local Peering Gateway (LPG)** ระหว่าง 2 VCN + route rule ทั้งสองฝั่งชี้ไป LPG
4. `sqlnet.ora` / `tnsnames.ora` ใน wallet ต้องเป็นชุด **private endpoint** (host เป็น FQDN
   ของ private endpoint ไม่ใช่ public) — ถ้าเปลี่ยน endpoint ต้อง re-download wallet และ
   อัปเดต secret `adb-wallet`:
   ```sh
   kubectl -n thanes-lims create secret generic adb-wallet \
     --from-file=<unzipped-wallet-dir>/ --dry-run=client -o yaml | kubectl apply -f -
   kubectl -n thanes-lims rollout restart deploy/thanes-lims-api
   ```

---

## เช็คลิสต์ wallet (ทั้ง 2 แบบ)

secret `adb-wallet` ต้องเป็นสำเนา wallet ที่ปรับให้รันในคอนเทนเนอร์:

- [ ] `sqlnet.ora` — `WALLET_LOCATION` DIRECTORY ต้องเป็น `"/app/wallet"` (ไม่ใช่ path เครื่อง dev)
- [ ] `tnsnames.ora` — มี alias `limsdb_high` (ตรงกับ `ORACLE_DSN` ใน secret `thanes-lims-secrets`)
- [ ] ไฟล์ครบ: `cwallet.sso`, `ewallet.p12`, `sqlnet.ora`, `tnsnames.ora` (+ jks/pem ไม่บังคับ)

ตรวจเร็ว ๆ:
```sh
kubectl -n thanes-lims exec deploy/thanes-lims-api -- sh -c 'cat /app/wallet/sqlnet.ora; grep limsdb_high /app/wallet/tnsnames.ora'
```

---

## สรุปสิ่งที่ต้องทำ (แบบ A, กรณีปัจจุบัน)

| # | งาน | ที่ทำ |
|---|-----|-------|
| 1 | หา NAT GW public IP ของ OKE | `kubectl run egress-check ... curl ifconfig.me` |
| 2 | เพิ่ม IP นั้นลง ADB Access Control List | OCI Console → ADB `LIMSDB` → Network → ACL |
| 3 | restart deployment + เช็ค log `oracle: connected to ADB` | `kubectl rollout restart` |

หลังจากนี้ `/api/v1/chat` จะทำงานได้ (ต้องมี image `0.0.11+` ที่มีโค้ด chatbot deploy อยู่ด้วย)
