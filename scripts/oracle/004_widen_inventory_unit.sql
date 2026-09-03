-- Widen inventory_items.unit for real Postgres data.
-- The original VARCHAR2(20) (byte semantics) overflows on multi-byte Thai
-- unit labels (e.g. "ไมโครลิตร/หลอด" = 24 bytes) once the mirror is fed the
-- system of record instead of the synthetic ASCII seed. See ORA-12899.
--
-- Run as CHATBOT_APP:
--   sqlcl -S "CHATBOT_APP/<password>@limsdb_high" @scripts/oracle/004_widen_inventory_unit.sql

ALTER TABLE inventory_items MODIFY (unit VARCHAR2(50 CHAR));

EXIT
