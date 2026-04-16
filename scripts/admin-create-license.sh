#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.license-dev.env"

REFERENCE=""
EMAIL=""
LICENSE_KEY=""
SERVER_URL=""
API_KEY=""

usage() {
  cat <<'EOF'
Uso: ./scripts/admin-create-license.sh [opzioni]

Opzioni:
  --reference <valore>    Riferimento ordine/cliente
  --email <valore>        Email cliente
  --license-key <valore>  Chiave esplicita (opzionale)
  --server-url <url>      Override URL server (default da env)
  --api-key <valore>      Override API key admin (default da env)
  -h, --help              Mostra aiuto

Esempi:
  ./scripts/admin-create-license.sh --reference order_123 --email user@example.com
  ./scripts/admin-create-license.sh --license-key NP-PRO-ABCDEF123456
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --reference)
      REFERENCE="${2:-}"
      shift 2
      ;;
    --email)
      EMAIL="${2:-}"
      shift 2
      ;;
    --license-key)
      LICENSE_KEY="${2:-}"
      shift 2
      ;;
    --server-url)
      SERVER_URL="${2:-}"
      shift 2
      ;;
    --api-key)
      API_KEY="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Argomento non riconosciuto: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -f "${ENV_FILE}" ]]; then
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
fi

SERVER_URL="${SERVER_URL:-${NP_LICENSE_SERVER_URL:-}}"
API_KEY="${API_KEY:-${NP_ADMIN_API_KEY:-}}"

if [[ -z "${SERVER_URL}" ]]; then
  echo "Server URL non impostato. Usa --server-url o NP_LICENSE_SERVER_URL." >&2
  exit 1
fi
if [[ -z "${API_KEY}" ]]; then
  echo "API key admin non impostata. Usa --api-key o NP_ADMIN_API_KEY." >&2
  exit 1
fi

PAYLOAD="$(
python3 - <<'PY' "${REFERENCE}" "${EMAIL}" "${LICENSE_KEY}"
import json
import sys

reference = sys.argv[1].strip()
email = sys.argv[2].strip()
license_key = sys.argv[3].strip()

payload = {}
if reference:
    payload["reference"] = reference
if email:
    payload["email"] = email
if license_key:
    payload["license_key"] = license_key

print(json.dumps(payload))
PY
)"

RESP="$(
curl -fsS \
  -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d "${PAYLOAD}" \
  "${SERVER_URL}/v1/admin/licenses/create"
)"

python3 - <<'PY' "${RESP}"
import json
import sys

resp = json.loads(sys.argv[1])
print("Stato:", resp.get("status", "n/a"))
print("License key:", resp.get("license_key", ""))
if resp.get("message"):
    print("Message:", resp["message"])
PY
