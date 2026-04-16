#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.license-dev.env"

SERVER_URL=""
WEBHOOK_SECRET=""
EMAIL="cliente@example.com"
LICENSE_KEY=""
REFERENCE="order_smoke_001"

usage() {
  cat <<'EOF'
Uso: ./scripts/test-stripe-webhook.sh [opzioni]

Opzioni:
  --server-url <url>         Override URL server
  --webhook-secret <secret>  Override secret webhook Stripe
  --email <email>            Email cliente per evento fake
  --license-key <key>        Chiave licenza da inserire nel metadata
  --reference <id>           Client reference id (default order_smoke_001)
  -h, --help                 Mostra aiuto

Esempio:
  ./scripts/test-stripe-webhook.sh --reference order_123 --email user@example.com
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --server-url)
      SERVER_URL="${2:-}"
      shift 2
      ;;
    --webhook-secret)
      WEBHOOK_SECRET="${2:-}"
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
    --reference)
      REFERENCE="${2:-}"
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
WEBHOOK_SECRET="${WEBHOOK_SECRET:-${NP_STRIPE_WEBHOOK_SECRET:-}}"

if [[ -z "${SERVER_URL}" ]]; then
  echo "Server URL non impostato. Usa --server-url o NP_LICENSE_SERVER_URL." >&2
  exit 1
fi
if [[ -z "${WEBHOOK_SECRET}" ]]; then
  echo "Webhook secret non impostata. Usa --webhook-secret o NP_STRIPE_WEBHOOK_SECRET." >&2
  exit 1
fi

PAYLOAD="$(
python3 - <<'PY' "${REFERENCE}" "${EMAIL}" "${LICENSE_KEY}"
import json
import sys
import time

reference = sys.argv[1].strip()
email = sys.argv[2].strip()
license_key = sys.argv[3].strip()
evt_id = f"evt_{int(time.time())}"
sess_id = f"cs_test_{int(time.time())}"

metadata = {}
if license_key:
    metadata["license_key"] = license_key

payload = {
    "id": evt_id,
    "type": "checkout.session.completed",
    "data": {
        "object": {
            "id": sess_id,
            "payment_status": "paid",
            "customer_email": email,
            "client_reference_id": reference,
            "metadata": metadata,
        }
    },
}
print(json.dumps(payload, separators=(",", ":")))
PY
)"

TS="$(date +%s)"
SIGNATURE="$(
python3 - <<'PY' "${WEBHOOK_SECRET}" "${TS}" "${PAYLOAD}"
import hashlib
import hmac
import sys

secret = sys.argv[1].encode("utf-8")
timestamp = sys.argv[2].encode("utf-8")
payload = sys.argv[3].encode("utf-8")
signed = timestamp + b"." + payload
digest = hmac.new(secret, signed, hashlib.sha256).hexdigest()
print(digest)
PY
)"

HEADER="t=${TS},v1=${SIGNATURE}"

echo "Invio webhook fake a ${SERVER_URL}/v1/webhooks/stripe ..."
RESP="$(
curl -fsS \
  -X POST \
  -H "Content-Type: application/json" \
  -H "Stripe-Signature: ${HEADER}" \
  -d "${PAYLOAD}" \
  "${SERVER_URL}/v1/webhooks/stripe"
)"

python3 - <<'PY' "${RESP}"
import json
import sys

resp = json.loads(sys.argv[1])
print("Stato:", resp.get("status", "n/a"))
if "license_key" in resp:
    print("License key:", resp["license_key"])
if resp.get("message"):
    print("Message:", resp["message"])
PY
