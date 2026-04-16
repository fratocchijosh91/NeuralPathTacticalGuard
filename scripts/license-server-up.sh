#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.license-dev.env"
PID_FILE="${ROOT_DIR}/.license-server.pid"
LOG_FILE="${ROOT_DIR}/.license-server.log"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "File ${ENV_FILE} non trovato." >&2
  echo "Esegui prima: ./scripts/dev-license-env.sh" >&2
  exit 1
fi

if command -v go >/dev/null 2>&1; then
  GO_BIN="$(command -v go)"
elif [[ -x "${ROOT_DIR}/.tools/go/bin/go" ]]; then
  GO_BIN="${ROOT_DIR}/.tools/go/bin/go"
else
  echo "Go non trovato. Installa Go o prepara ${ROOT_DIR}/.tools/go/bin/go" >&2
  exit 1
fi

# shellcheck disable=SC1090
set -a && source "${ENV_FILE}" && set +a

SERVER_URL="${NP_LICENSE_SERVER_URL:-http://127.0.0.1:8080}"

if [[ -f "${PID_FILE}" ]]; then
  OLD_PID="$(cat "${PID_FILE}" 2>/dev/null || true)"
  if [[ -n "${OLD_PID}" ]] && kill -0 "${OLD_PID}" >/dev/null 2>&1; then
    echo "License server già attivo (PID ${OLD_PID})."
    echo "Usa ./scripts/license-server-down.sh per fermarlo."
    exit 0
  fi
  rm -f "${PID_FILE}"
fi

if curl -fsS "${SERVER_URL}/healthz" >/dev/null 2>&1; then
  echo "Sembra già esserci un server attivo su ${SERVER_URL}."
  echo "Fermalo prima oppure usa una porta diversa in .license-dev.env."
  exit 1
fi

cd "${ROOT_DIR}"
nohup "${GO_BIN}" run ./cmd/license-server >"${LOG_FILE}" 2>&1 &
PID=$!
echo "${PID}" > "${PID_FILE}"

for _ in $(seq 1 30); do
  if curl -fsS "${SERVER_URL}/healthz" >/dev/null 2>&1; then
    echo "License server avviato."
    echo "  PID: ${PID}"
    echo "  URL: ${SERVER_URL}"
    echo "  LOG: ${LOG_FILE}"
    exit 0
  fi
  sleep 0.2
done

echo "Server avviato ma healthcheck non raggiungibile in tempo." >&2
echo "Controlla log: ${LOG_FILE}" >&2
exit 1
