#!/usr/bin/env bash
# Chatbot POC (Select AI) - Phase 3 step 1 (OpenAI provider): create the
# OpenAI credential object inside the POC Oracle ADB via
# DBMS_CLOUD.CREATE_CREDENTIAL.
#
# Switched from OCI Generative AI because the ADB's region
# (ap-samutprakan-1) has no Generative AI service - see
# docs/chatbot-poc-plan.md Phase 3 notes.
#
# Reads OPENAI_API_KEY and the DSN from .env. The API key is never written
# into a file tracked by git - the PL/SQL block is built in memory and
# piped to sqlcl over stdin.
#
# Usage: scripts/oracle/create_openai_credential.sh
set -euo pipefail
cd "$(dirname "$0")/../.."

set -a
source .env
set +a

: "${OPENAI_API_KEY:?set OPENAI_API_KEY in .env}"
: "${ORACLE_DSN:?set ORACLE_DSN in .env}"
: "${ORACLE_TNS_ADMIN:?set ORACLE_TNS_ADMIN in .env}"

export TNS_ADMIN="$ORACLE_TNS_ADMIN"

sqlcl -S "$ORACLE_DSN" <<SQL
BEGIN
  DBMS_CLOUD.CREATE_CREDENTIAL(
    credential_name => 'OPENAI_CRED',
    username        => 'OPENAI',
    password        => '${OPENAI_API_KEY}'
  );
END;
/
SELECT credential_name, username, enabled FROM user_credentials WHERE credential_name = 'OPENAI_CRED';
EXIT
SQL
