-- Chatbot POC (Select AI) - Oracle ADB schema.
-- Mirrors internal/domain/{sample,testresult,inventory,purchaseorder} into a
-- synthetic dataset on the POC ADB. This is NOT the system of record
-- (Postgres/GORM/golang-migrate remain authoritative) and is not synced
-- from Postgres - see docs/chatbot-poc-plan.md.
--
-- Run as the dedicated CHATBOT_APP user (see 000_create_app_user.sql), not ADMIN:
--   sqlcl -S "CHATBOT_APP/<password>@limsdb_high" @scripts/oracle/001_schema.sql
--
-- Column/table comments are written to be unambiguous for Select AI's
-- NL->SQL translation (DBMS_CLOUD_AI), not just for human readers.

CREATE TABLE samples (
    id          VARCHAR2(30)  PRIMARY KEY,
    name        VARCHAR2(200) NOT NULL,
    sample_type VARCHAR2(20)  NOT NULL
                 CHECK (sample_type IN ('blood','urine','water','tissue','food','serum')),
    custodian   VARCHAR2(100) NOT NULL,
    location    VARCHAR2(100) NOT NULL,
    status      VARCHAR2(20)  NOT NULL
                 CHECK (status IN ('pending','testing','completed','transferred')),
    received_at TIMESTAMP     NOT NULL
);

COMMENT ON TABLE samples IS 'ตัวอย่างที่ห้องปฏิบัติการรับเข้ามาเพื่อตรวจวิเคราะห์ (Chain-of-Custody เริ่มต้นที่นี่)';
COMMENT ON COLUMN samples.id IS 'รหัสตัวอย่าง เช่น SMP-2569-00001';
COMMENT ON COLUMN samples.name IS 'ชื่อ/คำอธิบายตัวอย่าง';
COMMENT ON COLUMN samples.sample_type IS 'ประเภทตัวอย่าง: blood=เลือด, urine=ปัสสาวะ, water=น้ำ, tissue=เนื้อเยื่อ, food=อาหาร, serum=ซีรัม';
COMMENT ON COLUMN samples.custodian IS 'ผู้รับผิดชอบดูแลตัวอย่าง ณ ปัจจุบัน';
COMMENT ON COLUMN samples.location IS 'ตำแหน่ง/แผนกที่จัดเก็บหรือดำเนินการตัวอย่างอยู่ในขณะนี้';
COMMENT ON COLUMN samples.status IS 'สถานะของตัวอย่าง: pending=รอดำเนินการ, testing=กำลังทดสอบ, completed=ทดสอบเสร็จสิ้น, transferred=ส่งต่อแผนกอื่น';
COMMENT ON COLUMN samples.received_at IS 'วันเวลาที่รับตัวอย่างเข้าระบบ';

CREATE TABLE test_results (
    id         VARCHAR2(30)  PRIMARY KEY,
    sample_id  VARCHAR2(30)  NOT NULL REFERENCES samples(id),
    test_name  VARCHAR2(200) NOT NULL,
    analyst    VARCHAR2(100) NOT NULL,
    result     VARCHAR2(200),
    flag       VARCHAR2(10)  CHECK (flag IN ('hi','lo','ok')),
    ref_range  VARCHAR2(100),
    status     VARCHAR2(30)  NOT NULL
                CHECK (status IN ('analyzing','pending_verification','approved'))
);

COMMENT ON TABLE test_results IS 'ผลการทดสอบของแต่ละตัวอย่าง (หนึ่งตัวอย่างมีได้หลายผลการทดสอบ)';
COMMENT ON COLUMN test_results.id IS 'รหัสผลการทดสอบ';
COMMENT ON COLUMN test_results.sample_id IS 'อ้างอิงถึงตัวอย่างใน samples.id ที่ผลการทดสอบนี้เป็นของ';
COMMENT ON COLUMN test_results.test_name IS 'ชื่อรายการทดสอบที่ทำ เช่น CBC, Glucose';
COMMENT ON COLUMN test_results.analyst IS 'ผู้ทำการวิเคราะห์/ทดสอบ';
COMMENT ON COLUMN test_results.result IS 'ค่าผลการทดสอบที่วัดได้';
COMMENT ON COLUMN test_results.flag IS 'ธงบ่งชี้ผลเทียบค่าอ้างอิง: hi=สูงกว่าปกติ, lo=ต่ำกว่าปกติ, ok=อยู่ในเกณฑ์ปกติ';
COMMENT ON COLUMN test_results.ref_range IS 'ช่วงค่าอ้างอิงปกติสำหรับรายการทดสอบนี้';
COMMENT ON COLUMN test_results.status IS 'สถานะผลการทดสอบ: analyzing=กำลังวิเคราะห์, pending_verification=รอตรวจสอบยืนยัน, approved=อนุมัติผลแล้ว';

CREATE TABLE inventory_items (
    id             VARCHAR2(30)  PRIMARY KEY,
    name           VARCHAR2(200) NOT NULL,
    category       VARCHAR2(100),
    quantity       NUMBER        NOT NULL,
    unit           VARCHAR2(20)  NOT NULL,
    min_qty        NUMBER        NOT NULL,
    max_qty        NUMBER        NOT NULL,
    default_vendor VARCHAR2(200)
);

COMMENT ON TABLE inventory_items IS 'รายการวัสดุ/สารเคมี/อุปกรณ์คงคลังของห้องปฏิบัติการ';
COMMENT ON COLUMN inventory_items.id IS 'รหัสรายการคงคลัง';
COMMENT ON COLUMN inventory_items.name IS 'ชื่อรายการ';
COMMENT ON COLUMN inventory_items.category IS 'หมวดหมู่ของรายการ';
COMMENT ON COLUMN inventory_items.quantity IS 'จำนวนคงเหลือปัจจุบัน';
COMMENT ON COLUMN inventory_items.unit IS 'หน่วยนับ เช่น ขวด, กล่อง, ชิ้น';
COMMENT ON COLUMN inventory_items.min_qty IS 'จำนวนขั้นต่ำที่ควรมีสำรองไว้ (ต่ำกว่านี้ถือว่าสต็อกต่ำ ต้องสั่งซื้อเพิ่ม)';
COMMENT ON COLUMN inventory_items.max_qty IS 'จำนวนสูงสุดที่ควรสำรองไว้';
COMMENT ON COLUMN inventory_items.default_vendor IS 'ผู้ขาย/ผู้จำหน่ายหลักที่ใช้สั่งซื้อรายการนี้';

CREATE TABLE purchase_orders (
    id         VARCHAR2(30)  PRIMARY KEY,
    item_id    VARCHAR2(30)  NOT NULL REFERENCES inventory_items(id),
    quantity   NUMBER        NOT NULL,
    vendor     VARCHAR2(200) NOT NULL,
    order_date DATE          NOT NULL,
    status     VARCHAR2(30)  NOT NULL
                CHECK (status IN ('pending_approval','sent_to_vendor','received','cancelled'))
);

COMMENT ON TABLE purchase_orders IS 'ใบสั่งซื้อสำหรับเติมสต็อกรายการคงคลัง';
COMMENT ON COLUMN purchase_orders.id IS 'รหัสใบสั่งซื้อ';
COMMENT ON COLUMN purchase_orders.item_id IS 'อ้างอิงถึงรายการคงคลังใน inventory_items.id ที่สั่งซื้อ';
COMMENT ON COLUMN purchase_orders.quantity IS 'จำนวนที่สั่งซื้อ';
COMMENT ON COLUMN purchase_orders.vendor IS 'ผู้ขาย/ผู้จำหน่ายที่สั่งซื้อจริงในใบสั่งซื้อนี้';
COMMENT ON COLUMN purchase_orders.order_date IS 'วันที่ออกใบสั่งซื้อ';
COMMENT ON COLUMN purchase_orders.status IS 'สถานะใบสั่งซื้อ: pending_approval=รออนุมัติ, sent_to_vendor=ส่งให้ผู้ขายแล้ว, received=รับของแล้ว, cancelled=ยกเลิก';

EXIT
