#!/usr/bin/env bash
# Chatbot POC (Select AI) - Phase 3 step 2: create the Select AI profile
# inside the POC Oracle ADB via DBMS_CLOUD_AI.CREATE_PROFILE, scoped to the
# credential from create_openai_credential.sh and the 4 Phase 1 tables only.
#
# Provider is OpenAI, not OCI Generative AI - the ADB's region
# (ap-samutprakan-1) has no OCI Generative AI service. This also drops the
# now-unused OCI_GENAI_CRED credential from the earlier attempt. See
# docs/chatbot-poc-plan.md Phase 3 notes for the pivot.
#
# "conversation" is deliberately NOT set to "true": multi-turn is out of
# scope for this POC per the plan doc.
#
# Usage: scripts/oracle/create_ai_profile.sh
set -euo pipefail
cd "$(dirname "$0")/../.."

set -a
source .env
set +a

: "${ORACLE_DSN:?set ORACLE_DSN in .env}"
: "${ORACLE_TNS_ADMIN:?set ORACLE_TNS_ADMIN in .env}"

export TNS_ADMIN="$ORACLE_TNS_ADMIN"

sqlcl -S "$ORACLE_DSN" <<SQL
BEGIN
  DBMS_CLOUD_AI.DROP_PROFILE(profile_name => 'CHATBOT_AI_PROFILE', force => true);
EXCEPTION
  WHEN OTHERS THEN NULL;
END;
/

BEGIN
  DBMS_CLOUD.DROP_CREDENTIAL(credential_name => 'OCI_GENAI_CRED');
EXCEPTION
  WHEN OTHERS THEN NULL;
END;
/

BEGIN
  DBMS_CLOUD_AI.CREATE_PROFILE(
    profile_name => 'CHATBOT_AI_PROFILE',
    attributes   => '{
        "provider": "openai",
        "credential_name": "OPENAI_CRED",
        "object_list": [
            {"owner": "CHATBOT_APP", "name": "SAMPLES"},
            {"owner": "CHATBOT_APP", "name": "TEST_RESULTS"},
            {"owner": "CHATBOT_APP", "name": "INVENTORY_ITEMS"},
            {"owner": "CHATBOT_APP", "name": "PURCHASE_ORDERS"}
        ],
        "model": "gpt-5.4"
    }'
  );
END;
/
SELECT profile_name, status FROM user_cloud_ai_profiles WHERE profile_name = 'CHATBOT_AI_PROFILE';
EXIT
SQL
