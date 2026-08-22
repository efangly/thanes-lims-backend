#!/usr/bin/env bash
# Chatbot POC (Select AI) - Phase 3 step 1: create the OCI Generative AI
# credential object inside the POC Oracle ADB via DBMS_CLOUD.CREATE_CREDENTIAL.
#
# Reads the OCI API signing key details from .env (OCI_USER_OCID,
# OCI_TENANCY_OCID, OCI_FINGERPRINT, OCI_PRIVATE_KEY_PATH) and the DSN
# already used for the chatbot schema (ORACLE_DSN, ORACLE_TNS_ADMIN).
# The PL/SQL block is built in memory and piped to sqlcl over stdin - the
# private key is never written into a file tracked by git.
#
# Usage: scripts/oracle/create_oci_credential.sh
set -euo pipefail
cd "$(dirname "$0")/../.."

set -a
source .env
set +a

: "${OCI_USER_OCID:?set OCI_USER_OCID in .env}"
: "${OCI_TENANCY_OCID:?set OCI_TENANCY_OCID in .env}"
: "${OCI_FINGERPRINT:?set OCI_FINGERPRINT in .env}"
: "${OCI_PRIVATE_KEY_PATH:?set OCI_PRIVATE_KEY_PATH in .env}"
: "${ORACLE_DSN:?set ORACLE_DSN in .env}"
: "${ORACLE_TNS_ADMIN:?set ORACLE_TNS_ADMIN in .env}"

if [[ ! -f "$OCI_PRIVATE_KEY_PATH" ]]; then
  echo "error: OCI_PRIVATE_KEY_PATH ($OCI_PRIVATE_KEY_PATH) not found" >&2
  exit 1
fi

# CREATE_CREDENTIAL expects the key body only - strip the PEM header/footer
# and newlines so it's a single base64 string.
PRIVATE_KEY=$(grep -v -- '-----' "$OCI_PRIVATE_KEY_PATH" | tr -d '\n')

export TNS_ADMIN="$ORACLE_TNS_ADMIN"

sqlcl -S "$ORACLE_DSN" <<SQL
BEGIN
  DBMS_CLOUD.CREATE_CREDENTIAL(
    credential_name => 'OCI_GENAI_CRED',
    user_ocid       => '${OCI_USER_OCID}',
    tenancy_ocid    => '${OCI_TENANCY_OCID}',
    private_key     => '${PRIVATE_KEY}',
    fingerprint     => '${OCI_FINGERPRINT}'
  );
END;
/
SELECT credential_name, username, enabled FROM user_credentials WHERE credential_name = 'OCI_GENAI_CRED';
EXIT
SQL
