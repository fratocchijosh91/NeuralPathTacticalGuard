#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.license-dev.env"
PID_FILE="${ROOT_DIR}/.license-server.pid"

SERVER_URL="http://127.0.0.1:8080"
if [[ -f "${ENV_FILE}" ]]; then
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  SERVER_URL="${NP_LICENSE_SERVER_URL:-$SERVER_URL}"
fi

kill_pid() {
  local pid="$1"
  if [[ -z "${pid}" ]]; then
    return 0
  fi
  if ! kill -0 "${pid}" >/dev/null 2>&1; then
    return 0
  fi

  kill "${pid}" >/dev/null 2>&1 || true
  for _ in $(seq 1 20); do
    if ! kill -0 "${pid}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.1
  done
  kill -9 "${pid}" >/dev/null 2>&1 || true
}

if [[ -f "${PID_FILE}" ]]; then
  PID="$(cat "${PID_FILE}" 2>/dev/null || true)"
  kill_pid "${PID}"
  rm -f "${PID_FILE}"
fi

PORT="$(python3 - <<'PY' "${SERVER_URL}"
import re, sys
url = sys.argv[1]
m = re.search(r":(\d+)", url)
print(m.group(1) if m else "8080")
PY
)"

EXTRA_PID="$(lsof -nP -iTCP:${PORT} -sTCP:LISTEN -t 2>/dev/null | head -n 1 || true)"
if [[ -n "${EXTRA_PID}" ]]; then
  kill_pid "${EXTRA_PID}"
fi

if curl -fsS "${SERVER_URL}/healthz" >/dev/null 2>&1; then
  echo "Server ancora attivo su ${SERVER_URL}." >&2
  exit 1
fi

echo "License server fermato (se era attivo)."
