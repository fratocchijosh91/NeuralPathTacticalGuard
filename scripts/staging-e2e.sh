#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.license-dev.env"

SERVER_URL=""
ADMIN_API_KEY=""
WEBHOOK_SECRET=""
REFERENCE="order_e2e_$(date +%s)"
EMAIL="e2e@example.com"
LICENSE_KEY=""

PASS=0
FAIL=0

usage() {
  cat <<'EOF'
Uso: ./scripts/staging-e2e.sh [opzioni]

Opzioni:
  --server-url <url>         Override URL server
  --admin-api-key <key>      Override API key admin
  --webhook-secret <secret>  Override secret webhook Stripe
  --reference <id>           Reference test (default order_e2e_<timestamp>)
  --email <email>            Email test (default e2e@example.com)
  --license-key <key>        Chiave esplicita da usare nei test
  -h, --help                 Mostra aiuto
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --server-url)
      SERVER_URL="${2:-}"
      shift 2
      ;;
    --admin-api-key)
      ADMIN_API_KEY="${2:-}"
      shift 2
      ;;
    --webhook-secret)
      WEBHOOK_SECRET="${2:-}"
      shift 2
      ;;
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
ADMIN_API_KEY="${ADMIN_API_KEY:-${NP_ADMIN_API_KEY:-}}"
WEBHOOK_SECRET="${WEBHOOK_SECRET:-${NP_STRIPE_WEBHOOK_SECRET:-}}"

if [[ -z "${SERVER_URL}" ]]; then
  echo "Server URL non impostato. Usa --server-url o NP_LICENSE_SERVER_URL." >&2
  exit 1
fi
if [[ -z "${ADMIN_API_KEY}" ]]; then
  echo "Admin API key non impostata. Usa --admin-api-key o NP_ADMIN_API_KEY." >&2
  exit 1
fi
if [[ -z "${WEBHOOK_SECRET}" ]]; then
  echo "Webhook secret non impostata. Usa --webhook-secret o NP_STRIPE_WEBHOOK_SECRET." >&2
  exit 1
fi

step() {
  local name="$1"
  shift
  echo ">> ${name}"
  if "$@"; then
    echo "PASS: ${name}"
    PASS=$((PASS + 1))
  else
    echo "FAIL: ${name}"
    FAIL=$((FAIL + 1))
  fi
  echo ""
}

extract_license_key_from_output() {
  python3 - <<'PY' "$1"
import re
import sys
text = sys.argv[1]
m = re.search(r"License key:\s*(\S+)", text)
print(m.group(1) if m else "")
PY
}

do_healthcheck() {
  curl -fsS "${SERVER_URL}/healthz" >/dev/null
}

do_admin_create() {
  local output
  if [[ -n "${LICENSE_KEY}" ]]; then
    output="$("${ROOT_DIR}/scripts/admin-create-license.sh" \
      --server-url "${SERVER_URL}" \
      --api-key "${ADMIN_API_KEY}" \
      --reference "${REFERENCE}" \
      --email "${EMAIL}" \
      --license-key "${LICENSE_KEY}")"
  else
    output="$("${ROOT_DIR}/scripts/admin-create-license.sh" \
      --server-url "${SERVER_URL}" \
      --api-key "${ADMIN_API_KEY}" \
      --reference "${REFERENCE}" \
      --email "${EMAIL}")"
  fi

  echo "${output}"
  local created
  created="$(extract_license_key_from_output "${output}")"
  if [[ -z "${created}" ]]; then
    return 1
  fi
  LICENSE_KEY="${created}"
  return 0
}

do_activate() {
  local machine_id payload resp token
  machine_id="$(
python3 - <<'PY'
import hashlib
import socket
import platform
fingerprint = f"{socket.gethostname().strip().lower()}|{platform.system().lower()}|{platform.machine().lower()}"
print(hashlib.sha256(fingerprint.encode("utf-8")).digest()[:8].hex().upper())
PY
)"

  payload="$(
python3 - <<'PY' "${LICENSE_KEY}" "${machine_id}"
import json
import sys
print(json.dumps({
  "license_key": sys.argv[1],
  "machine_id": sys.argv[2],
  "product": "neuralpath-tactical-guard",
  "version": "staging-e2e",
}))
PY
)"

  resp="$(
curl -fsS -X POST \
  -H "Content-Type: application/json" \
  -d "${payload}" \
  "${SERVER_URL}/v1/licenses/activate"
)"
  echo "${resp}"

  token="$(
python3 - <<'PY' "${resp}"
import json
import sys
obj = json.loads(sys.argv[1])
print(obj.get("token", ""))
PY
)"
  [[ -n "${token}" ]]
}

do_webhook() {
  "${ROOT_DIR}/scripts/test-stripe-webhook.sh" \
    --server-url "${SERVER_URL}" \
    --webhook-secret "${WEBHOOK_SECRET}" \
    --reference "${REFERENCE}" \
    --email "${EMAIL}" \
    --license-key "${LICENSE_KEY}" >/dev/null
}

echo "=== Staging E2E ==="
echo "Server: ${SERVER_URL}"
echo "Reference: ${REFERENCE}"
echo "Email: ${EMAIL}"
echo ""

step "Healthcheck server" do_healthcheck
step "Admin create license" do_admin_create
step "Activate endpoint" do_activate
step "Stripe webhook test" do_webhook

echo "==========================="
echo "Risultato: ${PASS} passati, ${FAIL} falliti"
echo "License key usata: ${LICENSE_KEY}"
echo "==========================="

if [[ "${FAIL}" -gt 0 ]]; then
  exit 1
fi
