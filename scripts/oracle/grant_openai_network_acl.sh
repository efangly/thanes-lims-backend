#!/usr/bin/env bash
# Chatbot POC (Select AI) - Phase 3 step 1b (OpenAI provider): grant the
# CHATBOT_APP schema network access to api.openai.com so Select AI can call
# OpenAI's API from inside the ADB.
#
# One-time, ADMIN-only step - CHATBOT_APP wasn't granted
# DBMS_NETWORK_ACL_ADMIN in Phase 0 (000_create_app_user.sql), and network
# ACLs on ADB require the privileged user to manage. Needs ORACLE_ADMIN_DSN
# in .env (ADMIN/<password>@limsdb_high) - only needed for this run, safe
# to blank back out afterwards.
#
# Usage: scripts/oracle/grant_openai_network_acl.sh
set -euo pipefail
cd "$(dirname "$0")/../.."

set -a
source .env
set +a

: "${ORACLE_ADMIN_DSN:?set ORACLE_ADMIN_DSN in .env (ADMIN/<password>@limsdb_high) - see .env comment}"
: "${ORACLE_TNS_ADMIN:?set ORACLE_TNS_ADMIN in .env}"

export TNS_ADMIN="$ORACLE_TNS_ADMIN"

sqlcl -S "$ORACLE_ADMIN_DSN" <<SQL
BEGIN
  DBMS_NETWORK_ACL_ADMIN.APPEND_HOST_ACE(
    host => 'api.openai.com',
    ace  => xs\$ace_type(privilege_list => xs\$name_list('http'),
                principal_name => 'CHATBOT_APP',
                principal_type => xs_acl.ptype_db)
  );
END;
/
EXIT
SQL

echo "Network ACL granted for api.openai.com -> CHATBOT_APP. You can blank ORACLE_ADMIN_DSN back out in .env now."
