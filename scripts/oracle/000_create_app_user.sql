-- Chatbot POC (Select AI) - dedicated Oracle ADB schema/user.
-- Run once as ADMIN. Keeps the chatbot's tables and Select AI objects out of
-- the ADMIN schema, and gives the Go backend (godror) a least-privilege
-- credential instead of connecting as ADMIN.
--
-- Usage:
--   sqlcl -S "ADMIN/<admin_password>@limsdb_high" @scripts/oracle/000_create_app_user.sql
--
-- CHATBOT_APP_PASSWORD must satisfy ADB password complexity: 12-30 chars,
-- at least one uppercase, one lowercase, one digit, no 3+ repeated chars in
-- a row, and must not contain the username.

DEFINE app_user = CHATBOT_APP
DEFINE app_password = "Cb_NRNOD45eEFbZimPkQd6H9!"

CREATE USER &app_user IDENTIFIED BY "&app_password";

-- Session + table/view DDL for Phase 1 schema, unlimited quota on the
-- ADB default tablespace (DATA) so table/index creation isn't blocked.
GRANT CREATE SESSION TO &app_user;
GRANT CREATE TABLE TO &app_user;
GRANT CREATE VIEW TO &app_user;
ALTER USER &app_user QUOTA UNLIMITED ON DATA;

-- Select AI (Phase 3): CREATE_CREDENTIAL / CREATE_PROFILE and DBMS_CLOUD_AI
-- narrate calls both require EXECUTE on these packages.
GRANT EXECUTE ON DBMS_CLOUD TO &app_user;
GRANT EXECUTE ON DBMS_CLOUD_AI TO &app_user;

EXIT
