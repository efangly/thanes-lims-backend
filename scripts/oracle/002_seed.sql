-- Chatbot POC (Select AI) - synthetic seed data.
-- Data is entirely fabricated for demo purposes; NOT synced from Postgres.
-- See docs/chatbot-poc-plan.md Phase 2 for the demo scenarios this dataset
-- is designed to cover.
--
-- Run as CHATBOT_APP (same user that owns the Phase 1 schema):
--   sqlcl -S "CHATBOT_APP/<password>@limsdb_high" @scripts/oracle/002_seed.sql
--
-- Covers ~1 month of activity (SYSDATE - 30 .. SYSDATE) so demo questions
-- like "สรุปรายสัปดาห์" / "เดือนนี้" / "ค้างเกิน 7 วัน" all have enough
-- spread to answer meaningfully. Dates are relative to SYSDATE at seed time
-- so the dataset stays usable whenever it's (re)run.

-- ---------------------------------------------------------------------
-- Clear any previously seeded rows first (safe to re-run this script).
-- Child tables before parents to respect FKs.
-- ---------------------------------------------------------------------
DELETE FROM test_results;
DELETE FROM purchase_orders;
DELETE FROM samples;
DELETE FROM inventory_items;

-- ---------------------------------------------------------------------
-- samples (40 rows, spread across ~30 days)
-- ---------------------------------------------------------------------
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00001', 'ตัวอย่างเลือดตรวจ CBC ผู้ป่วยนอก', 'blood',  'สมชาย ใจดี',      'แผนกรับตัวอย่าง',          'completed',   SYSDATE - 30);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00002', 'ตัวอย่างน้ำดื่มโรงงาน A',            'water',  'สมชาย ใจดี',      'ห้องปฏิบัติการเคมี',        'completed',   SYSDATE - 29);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00003', 'ตัวอย่างปัสสาวะตรวจคัดกรอง',         'urine',  'วิภา สายใจ',      'ห้องปฏิบัติการชีวเคมี',     'completed',   SYSDATE - 28);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00004', 'ตัวอย่างเนื้อเยื่อชิ้นเนื้อตรวจ',     'tissue', 'วิภา สายใจ',      'ห้องปฏิบัติการพยาธิ',       'completed',   SYSDATE - 27);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00005', 'ตัวอย่างซีรัมตรวจภูมิคุ้มกัน',        'serum',  'ประยุทธ์ แสงทอง',  'ห้องปฏิบัติการภูมิคุ้มกัน', 'completed',   SYSDATE - 26);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00006', 'ตัวอย่างอาหารทะเลตรวจปนเปื้อน',      'food',   'ประยุทธ์ แสงทอง',  'ห้องปฏิบัติการจุลชีววิทยา', 'completed',   SYSDATE - 25);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00007', 'ตัวอย่างเลือดตรวจไขมัน',             'blood',  'อรทัย พงษ์ศรี',   'ห้องปฏิบัติการชีวเคมี',     'completed',   SYSDATE - 25);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00008', 'ตัวอย่างน้ำเสียโรงงาน B',            'water',  'อรทัย พงษ์ศรี',   'ห้องปฏิบัติการเคมี',        'completed',   SYSDATE - 24);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00009', 'ตัวอย่างปัสสาวะตรวจการทำงานไต',      'urine',  'สมชาย ใจดี',      'ห้องปฏิบัติการชีวเคมี',     'transferred', SYSDATE - 23);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00010', 'ตัวอย่างเนื้อเยื่อตรวจโลหะหนัก',      'tissue', 'วิภา สายใจ',      'ห้องปฏิบัติการเคมี',        'completed',   SYSDATE - 22);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00011', 'ตัวอย่างซีรัมตรวจฮอร์โมน',            'serum',  'ประยุทธ์ แสงทอง',  'ห้องปฏิบัติการภูมิคุ้มกัน', 'completed',   SYSDATE - 22);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00012', 'ตัวอย่างเลือดตรวจน้ำตาลในเลือด',      'blood',  'อรทัย พงษ์ศรี',   'แผนกรับตัวอย่าง',          'completed',   SYSDATE - 21);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00013', 'ตัวอย่างน้ำดื่มโรงงาน C',            'water',  'สมชาย ใจดี',      'ห้องปฏิบัติการเคมี',        'completed',   SYSDATE - 20);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00014', 'ตัวอย่างเลือดตรวจตับ',               'blood',  'อรทัย พงษ์ศรี',   'ห้องปฏิบัติการชีวเคมี',     'completed',   SYSDATE - 20);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00015', 'ตัวอย่างปัสสาวะตรวจน้ำตาล',          'urine',  'วิภา สายใจ',      'ห้องปฏิบัติการชีวเคมี',     'completed',   SYSDATE - 19);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00016', 'ตัวอย่างอาหารทะเลตรวจปรอท',          'food',   'ประยุทธ์ แสงทอง',  'ห้องปฏิบัติการจุลชีววิทยา', 'completed',   SYSDATE - 18);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00017', 'ตัวอย่างซีรัมตรวจไวรัสตับอักเสบ',     'serum',  'ประยุทธ์ แสงทอง',  'ห้องปฏิบัติการภูมิคุ้มกัน', 'transferred', SYSDATE - 18);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00018', 'ตัวอย่างเนื้อเยื่อตรวจสารหนู',        'tissue', 'วิภา สายใจ',      'ห้องปฏิบัติการเคมี',        'completed',   SYSDATE - 17);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00019', 'ตัวอย่างเลือดตรวจ CBC พนักงานตรวจสุขภาพ', 'blood', 'สมชาย ใจดี',   'แผนกรับตัวอย่าง',          'completed',   SYSDATE - 16);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00020', 'ตัวอย่างน้ำดื่มโรงงาน A (รอบ 2)',    'water',  'สมชาย ใจดี',      'ห้องปฏิบัติการเคมี',        'completed',   SYSDATE - 16);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00021', 'ตัวอย่างเลือดตรวจ CBC ผู้ป่วยใน',    'blood',  'สมชาย ใจดี',      'แผนกรับตัวอย่าง',          'pending',     SYSDATE - 15);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00022', 'ตัวอย่างน้ำดื่มโรงงาน A',            'water',  'สมชาย ใจดี',      'ห้องปฏิบัติการเคมี',        'pending',     SYSDATE - 14);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00023', 'ตัวอย่างปัสสาวะตรวจการติดเชื้อ',      'urine',  'วิภา สายใจ',      'ห้องปฏิบัติการชีวเคมี',     'completed',   SYSDATE - 13);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00024', 'ตัวอย่างเนื้อเยื่อชิ้นเนื้อตรวจซ้ำ',   'tissue', 'วิภา สายใจ',      'ห้องปฏิบัติการพยาธิ',       'completed',   SYSDATE - 12);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00025', 'ตัวอย่างซีรัมตรวจภูมิคุ้มกัน (ติดตามผล)', 'serum', 'ประยุทธ์ แสงทอง', 'ห้องปฏิบัติการภูมิคุ้มกัน', 'testing',   SYSDATE - 11);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00026', 'ตัวอย่างอาหารทะเลตรวจปนเปื้อน (รอบ 2)', 'food', 'ประยุทธ์ แสงทอง', 'ห้องปฏิบัติการจุลชีววิทยา', 'completed', SYSDATE - 10);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00027', 'ตัวอย่างน้ำดื่มโรงงาน A',            'water',  'สมชาย ใจดี',      'ห้องปฏิบัติการเคมี',        'pending',     SYSDATE - 10);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00028', 'ตัวอย่างเลือดตรวจไขมัน (ติดตามผล)',   'blood',  'อรทัย พงษ์ศรี',   'ห้องปฏิบัติการชีวเคมี',     'completed',   SYSDATE - 9);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00029', 'ตัวอย่างเนื้อเยื่อตรวจโลหะหนัก (รอบ 2)', 'tissue', 'วิภา สายใจ',   'ห้องปฏิบัติการเคมี',        'completed',   SYSDATE - 9);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00030', 'ตัวอย่างซีรัมตรวจฮอร์โมนไทรอยด์',     'serum',  'ประยุทธ์ แสงทอง',  'ห้องปฏิบัติการภูมิคุ้มกัน', 'testing',     SYSDATE - 8);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00031', 'ตัวอย่างเลือดตรวจน้ำตาลในเลือด',      'blood',  'อรทัย พงษ์ศรี',   'แผนกรับตัวอย่าง',          'pending',     SYSDATE - 7);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00032', 'ตัวอย่างน้ำเสียโรงงาน B (รอบ 2)',    'water',  'อรทัย พงษ์ศรี',   'ห้องปฏิบัติการเคมี',        'testing',     SYSDATE - 7);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00033', 'ตัวอย่างปัสสาวะตรวจการทำงานไต (ติดตามผล)', 'urine', 'สมชาย ใจดี',  'ห้องปฏิบัติการชีวเคมี',     'transferred', SYSDATE - 6);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00034', 'ตัวอย่างซีรัมตรวจภูมิคุ้มกัน',        'serum',  'ประยุทธ์ แสงทอง',  'ห้องปฏิบัติการภูมิคุ้มกัน', 'completed',   SYSDATE - 6);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00035', 'ตัวอย่างอาหารทะเลตรวจปนเปื้อน',      'food',   'ประยุทธ์ แสงทอง',  'ห้องปฏิบัติการจุลชีววิทยา', 'testing',     SYSDATE - 5);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00036', 'ตัวอย่างเลือดตรวจไขมัน',             'blood',  'อรทัย พงษ์ศรี',   'ห้องปฏิบัติการชีวเคมี',     'completed',   SYSDATE - 5);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00037', 'ตัวอย่างปัสสาวะตรวจคัดกรอง',         'urine',  'วิภา สายใจ',      'ห้องปฏิบัติการชีวเคมี',     'testing',     SYSDATE - 4);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00038', 'ตัวอย่างเนื้อเยื่อชิ้นเนื้อตรวจ',     'tissue', 'วิภา สายใจ',      'ห้องปฏิบัติการพยาธิ',       'testing',     SYSDATE - 2);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00039', 'ตัวอย่างซีรัมตรวจฮอร์โมน',            'serum',  'ประยุทธ์ แสงทอง',  'ห้องปฏิบัติการภูมิคุ้มกัน', 'testing',     SYSDATE - 3);
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at) VALUES
('SMP-2569-00040', 'ตัวอย่างน้ำเสียโรงงาน B',            'water',  'อรทัย พงษ์ศรี',   'ห้องปฏิบัติการเคมี',        'pending',     SYSDATE - 1);

-- ---------------------------------------------------------------------
-- test_results (linked to samples above; most completed/transferred
-- samples have an approved result, testing samples are mid-flow)
-- ---------------------------------------------------------------------
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00001', 'SMP-2569-00001', 'CBC - WBC',        'สมชาย ใจดี',      '7.2 x10^3/uL', 'ok', '4.0-10.0 x10^3/uL', 'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00002', 'SMP-2569-00002', 'Coliform Count',   'สมชาย ใจดี',      '40 CFU/g',   'ok', '<100 CFU/g',      'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00003', 'SMP-2569-00003', 'Urinalysis - Protein', 'วิภา สายใจ',  'negative',   'ok', 'negative',        'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00004', 'SMP-2569-00004', 'Histopathology',   'วิภา สายใจ',      'benign',     'ok', NULL,              'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00005', 'SMP-2569-00005', 'IgG',              'ประยุทธ์ แสงทอง', '1850 mg/dL', 'hi', '700-1600 mg/dL', 'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00006', 'SMP-2569-00005', 'IgM',              'ประยุทธ์ แสงทอง', '95 mg/dL',   'ok', '40-230 mg/dL',    'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00007', 'SMP-2569-00006', 'Coliform Count',   'ประยุทธ์ แสงทอง', '620 CFU/g',  'hi', '<100 CFU/g',      'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00008', 'SMP-2569-00006', 'Salmonella spp.',  'ประยุทธ์ แสงทอง', 'not detected', 'ok', 'not detected', 'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00009', 'SMP-2569-00007', 'LDL Cholesterol',  'อรทัย พงษ์ศรี',   '190 mg/dL',  'hi', '<130 mg/dL',      'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00010', 'SMP-2569-00007', 'HDL Cholesterol',  'อรทัย พงษ์ศรี',   '32 mg/dL',   'lo', '>40 mg/dL',       'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00011', 'SMP-2569-00008', 'BOD',              'อรทัย พงษ์ศรี',   '35 mg/L',    'hi', '<20 mg/L',        'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00012', 'SMP-2569-00009', 'Creatinine',       'สมชาย ใจดี',      '0.9 mg/dL',  'ok', '0.6-1.2 mg/dL',   'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00013', 'SMP-2569-00010', 'Lead (Pb)',        'วิภา สายใจ',      '4.2 ppm',    'hi', '<1.0 ppm',        'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00014', 'SMP-2569-00010', 'Cadmium (Cd)',     'วิภา สายใจ',      '0.3 ppm',    'ok', '<0.5 ppm',        'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00015', 'SMP-2569-00011', 'TSH',              'ประยุทธ์ แสงทอง', '4.1 uIU/mL', 'ok', '0.4-4.5 uIU/mL',  'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00016', 'SMP-2569-00012', 'Glucose (Fasting)', 'อรทัย พงษ์ศรี',  '105 mg/dL',  'hi', '70-100 mg/dL',    'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00017', 'SMP-2569-00013', 'Coliform Count',   'สมชาย ใจดี',      '15 CFU/g',   'ok', '<100 CFU/g',      'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00018', 'SMP-2569-00014', 'ALT',              'อรทัย พงษ์ศรี',   '65 U/L',     'hi', '7-56 U/L',        'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00019', 'SMP-2569-00015', 'Urinalysis - Glucose', 'วิภา สายใจ',  'negative',   'ok', 'negative',        'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00020', 'SMP-2569-00016', 'Mercury (Hg)',     'ประยุทธ์ แสงทอง', '0.4 ppm',    'ok', '<0.5 ppm',        'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00021', 'SMP-2569-00017', 'HBsAg',            'ประยุทธ์ แสงทอง', 'reactive',   'hi', 'non-reactive',    'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00022', 'SMP-2569-00018', 'Arsenic (As)',     'วิภา สายใจ',      '0.8 ppm',    'hi', '<0.3 ppm',        'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00023', 'SMP-2569-00019', 'CBC - WBC',        'สมชาย ใจดี',      '6.5 x10^3/uL', 'ok', '4.0-10.0 x10^3/uL', 'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00024', 'SMP-2569-00020', 'Coliform Count',   'สมชาย ใจดี',      '55 CFU/g',   'ok', '<100 CFU/g',      'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00025', 'SMP-2569-00023', 'Urinalysis - WBC',  'วิภา สายใจ',     '25 /HPF',    'hi', '0-5 /HPF',        'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00026', 'SMP-2569-00024', 'Histopathology',   'วิภา สายใจ',      'malignant - grade 2', 'hi', NULL,    'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00027', 'SMP-2569-00025', 'IgG',              'ประยุทธ์ แสงทอง', NULL,         NULL, '700-1600 mg/dL', 'analyzing');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00028', 'SMP-2569-00026', 'Salmonella spp.',  'ประยุทธ์ แสงทอง', 'not detected', 'ok', 'not detected', 'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00029', 'SMP-2569-00028', 'LDL Cholesterol',  'อรทัย พงษ์ศรี',   '150 mg/dL',  'hi', '<130 mg/dL',      'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00030', 'SMP-2569-00029', 'Lead (Pb)',        'วิภา สายใจ',      '0.9 ppm',    'lo', '<1.0 ppm',        'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00031', 'SMP-2569-00030', 'TSH',              'ประยุทธ์ แสงทอง', NULL,         NULL, '0.4-4.5 uIU/mL',  'analyzing');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00032', 'SMP-2569-00032', 'BOD',              'อรทัย พงษ์ศรี',   NULL,         NULL, '<20 mg/L',        'analyzing');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00033', 'SMP-2569-00033', 'Creatinine',       'สมชาย ใจดี',      '1.4 mg/dL',  'hi', '0.6-1.2 mg/dL',   'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00034', 'SMP-2569-00034', 'IgM',              'ประยุทธ์ แสงทอง', '110 mg/dL',  'ok', '40-230 mg/dL',    'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00035', 'SMP-2569-00035', 'Coliform Count',   'ประยุทธ์ แสงทอง', NULL,         NULL, '<100 CFU/g',      'analyzing');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00036', 'SMP-2569-00036', 'LDL Cholesterol',  'อรทัย พงษ์ศรี',   '128 mg/dL',  'ok', '<130 mg/dL',      'approved');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00037', 'SMP-2569-00037', 'Urinalysis - Protein', 'วิภา สายใจ',  NULL,         NULL, 'negative',        'analyzing');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00038', 'SMP-2569-00038', 'Histopathology',   'วิภา สายใจ',      NULL,         NULL, NULL,              'analyzing');
INSERT INTO test_results (id, sample_id, test_name, analyst, result, flag, ref_range, status) VALUES
('TR-2569-00039', 'SMP-2569-00039', 'TSH',              'ประยุทธ์ แสงทอง', '4.1 uIU/mL', 'ok', '0.4-4.5 uIU/mL',  'pending_verification');

-- ---------------------------------------------------------------------
-- inventory_items
-- ---------------------------------------------------------------------
INSERT INTO inventory_items (id, name, category, quantity, unit, min_qty, max_qty, default_vendor) VALUES
('INV-0001', 'น้ำยา Glucose Reagent',       'สารเคมี',      5,   'ขวด',  20,  100, 'บริษัท ไทยเคมีแล็บ จำกัด');
INSERT INTO inventory_items (id, name, category, quantity, unit, min_qty, max_qty, default_vendor) VALUES
('INV-0002', 'ถุงมือไนไตรไซส์ M',           'วัสดุสิ้นเปลือง', 3,  'กล่อง', 15,  60,  'บริษัท เมดิคอลซัพพลาย จำกัด');
INSERT INTO inventory_items (id, name, category, quantity, unit, min_qty, max_qty, default_vendor) VALUES
('INV-0003', 'หลอดเก็บซีรัม (Serum Vial)',  'วัสดุสิ้นเปลือง', 8,  'กล่อง', 30,  150, 'บริษัท ไบโอแล็บ ซัพพลาย จำกัด');
INSERT INTO inventory_items (id, name, category, quantity, unit, min_qty, max_qty, default_vendor) VALUES
('INV-0004', 'ปลายปิเปต 200uL',             'วัสดุสิ้นเปลือง', 220, 'กล่อง', 50,  500, 'บริษัท ไบโอแล็บ ซัพพลาย จำกัด');
INSERT INTO inventory_items (id, name, category, quantity, unit, min_qty, max_qty, default_vendor) VALUES
('INV-0005', 'น้ำยา Coliform Test Kit',     'สารเคมี',      55,  'ชุด',  10,  80,  'บริษัท ไทยเคมีแล็บ จำกัด');
INSERT INTO inventory_items (id, name, category, quantity, unit, min_qty, max_qty, default_vendor) VALUES
('INV-0006', 'ชุดทดสอบโลหะหนัก (Heavy Metal Kit)', 'สารเคมี', 6, 'ชุด', 12,  50,  'บริษัท ไทยเคมีแล็บ จำกัด');
INSERT INTO inventory_items (id, name, category, quantity, unit, min_qty, max_qty, default_vendor) VALUES
('INV-0007', 'สำลีแอลกอฮอล์',                'วัสดุสิ้นเปลือง', 300, 'กล่อง', 50,  400, 'บริษัท เมดิคอลซัพพลาย จำกัด');
INSERT INTO inventory_items (id, name, category, quantity, unit, min_qty, max_qty, default_vendor) VALUES
('INV-0008', 'เข็มเจาะเลือด (Lancet)',       'วัสดุสิ้นเปลือง', 40,  'กล่อง', 20,  200, 'บริษัท เมดิคอลซัพพลาย จำกัด');
INSERT INTO inventory_items (id, name, category, quantity, unit, min_qty, max_qty, default_vendor) VALUES
('INV-0009', 'น้ำยา TSH Immunoassay Kit',   'สารเคมี',      14,  'ชุด',  15,  60,  'บริษัท ไบโอแล็บ ซัพพลาย จำกัด');
INSERT INTO inventory_items (id, name, category, quantity, unit, min_qty, max_qty, default_vendor) VALUES
('INV-0010', 'หลอดเก็บเลือด EDTA',          'วัสดุสิ้นเปลือง', 90,  'กล่อง', 30,  200, 'บริษัท ไบโอแล็บ ซัพพลาย จำกัด');
INSERT INTO inventory_items (id, name, category, quantity, unit, min_qty, max_qty, default_vendor) VALUES
('INV-0011', 'น้ำยา HBsAg Test Kit',        'สารเคมี',      9,   'ชุด',  15,  50,  'บริษัท ไบโอแล็บ ซัพพลาย จำกัด');
INSERT INTO inventory_items (id, name, category, quantity, unit, min_qty, max_qty, default_vendor) VALUES
('INV-0012', 'ชุดทดสอบ Salmonella',         'สารเคมี',      25,  'ชุด',  10,  60,  'บริษัท ไทยเคมีแล็บ จำกัด');

-- ---------------------------------------------------------------------
-- purchase_orders (spread across ~30 days)
-- ---------------------------------------------------------------------
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0001', 'INV-0002', 20, 'บริษัท เมดิคอลซัพพลาย จำกัด',    TRUNC(SYSDATE) - 30, 'received');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0002', 'INV-0004', 100, 'บริษัท ไบโอแล็บ ซัพพลาย จำกัด', TRUNC(SYSDATE) - 28, 'received');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0003', 'INV-0007', 80,  'บริษัท เมดิคอลซัพพลาย จำกัด',    TRUNC(SYSDATE) - 25, 'received');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0004', 'INV-0005', 30, 'บริษัท ไทยเคมีแล็บ จำกัด',       TRUNC(SYSDATE) - 21, 'cancelled');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0005', 'INV-0009', 20, 'บริษัท ไบโอแล็บ ซัพพลาย จำกัด',  TRUNC(SYSDATE) - 18, 'received');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0006', 'INV-0011', 15, 'บริษัท ไบโอแล็บ ซัพพลาย จำกัด',  TRUNC(SYSDATE) - 16, 'received');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0007', 'INV-0012', 30, 'บริษัท ไทยเคมีแล็บ จำกัด',       TRUNC(SYSDATE) - 14, 'received');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0008', 'INV-0008', 60, 'บริษัท เมดิคอลซัพพลาย จำกัด',    TRUNC(SYSDATE) - 12, 'received');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0009', 'INV-0006', 30, 'บริษัท ไทยเคมีแล็บ จำกัด',       TRUNC(SYSDATE) - 10, 'sent_to_vendor');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0010', 'INV-0010', 50, 'บริษัท ไบโอแล็บ ซัพพลาย จำกัด',  TRUNC(SYSDATE) - 9,  'received');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0011', 'INV-0003', 60, 'บริษัท ไบโอแล็บ ซัพพลาย จำกัด',  TRUNC(SYSDATE) - 6,  'sent_to_vendor');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0012', 'INV-0001', 50, 'บริษัท ไทยเคมีแล็บ จำกัด',       TRUNC(SYSDATE) - 4,  'pending_approval');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0013', 'INV-0002', 40, 'บริษัท เมดิคอลซัพพลาย จำกัด',    TRUNC(SYSDATE) - 3,  'sent_to_vendor');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0014', 'INV-0009', 20, 'บริษัท ไบโอแล็บ ซัพพลาย จำกัด',  TRUNC(SYSDATE) - 2,  'pending_approval');
INSERT INTO purchase_orders (id, item_id, quantity, vendor, order_date, status) VALUES
('PO-2569-0015', 'INV-0006', 25, 'บริษัท ไทยเคมีแล็บ จำกัด',       TRUNC(SYSDATE) - 1,  'pending_approval');

COMMIT;

-- ---------------------------------------------------------------------
-- Record counts (sanity check when run interactively)
-- ---------------------------------------------------------------------
SELECT 'samples' AS table_name, COUNT(*) AS row_count FROM samples
UNION ALL
SELECT 'test_results', COUNT(*) FROM test_results
UNION ALL
SELECT 'inventory_items', COUNT(*) FROM inventory_items
UNION ALL
SELECT 'purchase_orders', COUNT(*) FROM purchase_orders;

EXIT
