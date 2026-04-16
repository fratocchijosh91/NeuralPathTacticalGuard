#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.license-dev.env"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "File ${ENV_FILE} non trovato." >&2
  echo "Esegui prima: ./scripts/dev-license-env.sh" >&2
  exit 1
fi

# shellcheck disable=SC1090
source "${ENV_FILE}"

if command -v go >/dev/null 2>&1; then
  GO_BIN="$(command -v go)"
elif [[ -x "${ROOT_DIR}/.tools/go/bin/go" ]]; then
  GO_BIN="${ROOT_DIR}/.tools/go/bin/go"
else
  echo "Go non trovato. Installa Go o prepara ${ROOT_DIR}/.tools/go/bin/go" >&2
  exit 1
fi

: "${NP_LICENSE_SERVER_URL:?NP_LICENSE_SERVER_URL mancante in ${ENV_FILE}}"
: "${NP_LICENSE_PUBLIC_KEY_B64:?NP_LICENSE_PUBLIC_KEY_B64 mancante in ${ENV_FILE}}"
: "${NP_LICENSE_PRIVATE_KEY_B64:?NP_LICENSE_PRIVATE_KEY_B64 mancante in ${ENV_FILE}}"

MACHINE_ID="$(
python3 - <<'PY' "$(${GO_BIN} env GOOS)" "$(${GO_BIN} env GOARCH)"
import hashlib
import socket
import sys

goos = sys.argv[1]
goarch = sys.argv[2]
host = socket.gethostname().strip().lower()
fingerprint = f"{host}|{goos}|{goarch}"
digest = hashlib.sha256(fingerprint.encode("utf-8")).digest()[:8]
print(digest.hex().upper())
PY
)"

if [[ "${NP_LICENSE_ALLOW_ANY_KEY:-false}" == "true" ]]; then
  TEST_KEY="NP-PRO-SMOKE-TEST"
else
  FIRST_KEY="$(printf '%s' "${NP_LICENSE_KEYS:-}" | awk -F, '{gsub(/^ +| +$/, "", $1); print $1}')"
  if [[ -z "${FIRST_KEY}" ]]; then
    echo "NP_LICENSE_ALLOW_ANY_KEY=false ma NP_LICENSE_KEYS è vuoto." >&2
    exit 1
  fi
  TEST_KEY="${FIRST_KEY}"
fi

TMP_DIR="$(mktemp -d)"
SERVER_LOG="${TMP_DIR}/license-server.log"
SERVER_PID=""

cleanup() {
  if [[ -n "${SERVER_PID}" ]] && kill -0 "${SERVER_PID}" >/dev/null 2>&1; then
    kill "${SERVER_PID}" >/dev/null 2>&1 || true
    wait "${SERVER_PID}" >/dev/null 2>&1 || true
  fi
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

echo "Avvio server licenze locale..."
"${GO_BIN}" run ./cmd/license-server >"${SERVER_LOG}" 2>&1 &
SERVER_PID="$!"

echo "Attendo healthcheck..."
for _ in $(seq 1 30); do
  if curl -fsS "${NP_LICENSE_SERVER_URL}/healthz" >/dev/null 2>&1; then
    break
  fi
  sleep 0.5
done

if ! curl -fsS "${NP_LICENSE_SERVER_URL}/healthz" >/dev/null 2>&1; then
  echo "Server non raggiungibile. Log:" >&2
  cat "${SERVER_LOG}" >&2
  exit 1
fi

ACTIVATION_PAYLOAD="$(
python3 - <<'PY' "${TEST_KEY}" "${MACHINE_ID}"
import json
import sys
print(json.dumps({
    "license_key": sys.argv[1],
    "machine_id": sys.argv[2],
    "product": "neuralpath-tactical-guard",
    "version": "smoke-test",
}))
PY
)"

echo "Richiedo token di attivazione..."
ACTIVATION_RESPONSE="$(curl -fsS -X POST \
  -H "Content-Type: application/json" \
  -d "${ACTIVATION_PAYLOAD}" \
  "${NP_LICENSE_SERVER_URL}/v1/licenses/activate")"

TOKEN="$(
python3 - <<'PY' "${ACTIVATION_RESPONSE}"
import json
import sys
obj = json.loads(sys.argv[1])
print(obj.get("token", ""))
PY
)"

if [[ -z "${TOKEN}" ]]; then
  echo "Risposta attivazione senza token: ${ACTIVATION_RESPONSE}" >&2
  exit 1
fi

echo "Verifico token firmato..."
"${GO_BIN}" run ./cmd/license-token-check \
  -token "${TOKEN}" \
  -public-key-b64 "${NP_LICENSE_PUBLIC_KEY_B64}" \
  -expected-product "neuralpath-tactical-guard" \
  -expected-prefix "NP" \
  -expected-tier "PRO" \
  -expected-machine-id "${MACHINE_ID}"

echo ""
echo "Smoke test licensing completato con successo."
