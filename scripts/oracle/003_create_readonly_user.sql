-- Chatbot POC - dedicated READ-ONLY Oracle ADB user for the Go backend.
-- Run once as ADMIN (CHATBOT_APP is not privileged to CREATE USER).
--
-- Since the 2026-09-02 pivot the chatbot builds SQL via the Claude API and
-- runs it against this ADB. The generated SQL is untrusted, so the DB
-- credential it runs under is granted SELECT only on the four POC tables -
-- a hard backstop behind the Go-side SELECT/WITH guard.
--
-- Usage:
--   sqlcl -S "ADMIN/<admin_password>@limsdb_high" @scripts/oracle/003_create_readonly_user.sql
--
-- RO_PASSWORD must satisfy ADB password complexity: 12-30 chars, at least one
-- uppercase, one lowercase, one digit, no 3+ repeated chars in a row, and must
-- not contain the username. Put the resulting DSN in ORACLE_CHATBOT_DSN:
--   ORACLE_CHATBOT_DSN=CHATBOT_RO/"<password>"@limsdb_high

DEFINE ro_user = CHATBOT_RO
DEFINE ro_password = "Ro_9xKQ2mVt7ZbHnPdLcW!"

CREATE USER &ro_user IDENTIFIED BY "&ro_password";
GRANT CREATE SESSION TO &ro_user;

GRANT SELECT ON CHATBOT_APP.samples         TO &ro_user;
GRANT SELECT ON CHATBOT_APP.test_results    TO &ro_user;
GRANT SELECT ON CHATBOT_APP.inventory_items TO &ro_user;
GRANT SELECT ON CHATBOT_APP.purchase_orders TO &ro_user;

-- Let generated SQL reference the tables unqualified (SELECT ... FROM samples).
CREATE OR REPLACE SYNONYM &ro_user..samples         FOR CHATBOT_APP.samples;
CREATE OR REPLACE SYNONYM &ro_user..test_results    FOR CHATBOT_APP.test_results;
CREATE OR REPLACE SYNONYM &ro_user..inventory_items FOR CHATBOT_APP.inventory_items;
CREATE OR REPLACE SYNONYM &ro_user..purchase_orders FOR CHATBOT_APP.purchase_orders;

EXIT
