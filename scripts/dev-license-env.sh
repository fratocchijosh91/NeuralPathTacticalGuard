#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.license-dev.env"
RUN_SERVER=false

for arg in "$@"; do
  case "$arg" in
    --run)
      RUN_SERVER=true
      ;;
    --help|-h)
      echo "Uso: $0 [--run]"
      echo "  --run   avvia subito il server licenze locale"
      exit 0
      ;;
    *)
      echo "Argomento non riconosciuto: $arg" >&2
      exit 1
      ;;
  esac
done

if command -v go >/dev/null 2>&1; then
  GO_BIN="$(command -v go)"
elif [[ -x "${ROOT_DIR}/.tools/go/bin/go" ]]; then
  GO_BIN="${ROOT_DIR}/.tools/go/bin/go"
else
  echo "Go non trovato. Installa Go o prepara ${ROOT_DIR}/.tools/go/bin/go" >&2
  exit 1
fi

echo "Uso Go: ${GO_BIN}"

KEYS_OUTPUT="$("${GO_BIN}" run ./cmd/license-keygen)"
PRIVATE_KEY="$(printf '%s\n' "${KEYS_OUTPUT}" | awk -F= '/^NP_LICENSE_PRIVATE_KEY_B64=/{print $2}')"
PUBLIC_KEY="$(printf '%s\n' "${KEYS_OUTPUT}" | awk -F= '/^NP_LICENSE_PUBLIC_KEY_B64=/{print $2}')"

if [[ -z "${PRIVATE_KEY}" || -z "${PUBLIC_KEY}" ]]; then
  echo "Impossibile generare keypair licenze." >&2
  exit 1
fi

cat > "${ENV_FILE}" <<EOF
# File generato automaticamente da scripts/dev-license-env.sh
# NON usare in produzione.
NP_LICENSE_ADDR=:8080
NP_LICENSE_PRODUCT=neuralpath-tactical-guard
NP_LICENSE_TIER=PRO
NP_LICENSE_PREFIX=NP
NP_LICENSE_TOKEN_TTL_HOURS=720
NP_LICENSE_RATE_LIMIT_PER_MIN=10
NP_LICENSE_ALLOW_ANY_KEY=true
NP_LICENSE_KEYS=
NP_LICENSE_KEYS_PATH=data/allowed-keys.json
NP_STRIPE_WEBHOOK_SECRET=whsec_dev_only_change_me
NP_ADMIN_API_KEY=admin_dev_change_me
NP_LICENSE_PRIVATE_KEY_B64=${PRIVATE_KEY}
NP_LICENSE_PUBLIC_KEY_B64=${PUBLIC_KEY}
NP_LICENSE_SERVER_URL=http://127.0.0.1:8080
EOF

echo ""
echo "Creato ${ENV_FILE}"
echo "Passi successivi:"
echo "  1) source \"${ENV_FILE}\""
echo "  2) go run ./cmd/license-server"
echo ""
echo "Valori da mettere nel client (config.json o env):"
echo "  license_server_url = http://127.0.0.1:8080"
echo "  license_public_key = ${PUBLIC_KEY}"

if [[ "${RUN_SERVER}" == "true" ]]; then
  echo ""
  echo "Avvio server licenze locale..."
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  "${GO_BIN}" run ./cmd/license-server
fi
