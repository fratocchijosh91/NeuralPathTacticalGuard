#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.license-dev.env"
PID_FILE="${ROOT_DIR}/.license-server.pid"
LOG_FILE="${ROOT_DIR}/.license-server.log"

SERVER_URL="http://127.0.0.1:8080"
if [[ -f "${ENV_FILE}" ]]; then
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  SERVER_URL="${NP_LICENSE_SERVER_URL:-$SERVER_URL}"
fi

print_header() {
  echo "=== License Server Status ==="
  echo "URL: ${SERVER_URL}"
}

pid_from_file() {
  if [[ -f "${PID_FILE}" ]]; then
    cat "${PID_FILE}" 2>/dev/null || true
  fi
}

is_running_pid() {
  local pid="$1"
  if [[ -z "${pid}" ]]; then
    return 1
  fi
  kill -0 "${pid}" >/dev/null 2>&1
}

process_on_port() {
  local port
  port="$(python3 - <<'PY' "${SERVER_URL}"
import re, sys
url = sys.argv[1]
m = re.search(r":(\d+)", url)
print(m.group(1) if m else "8080")
PY
)"
  lsof -nP -iTCP:${port} -sTCP:LISTEN -t 2>/dev/null | head -n 1 || true
}

print_health() {
  if curl -fsS "${SERVER_URL}/healthz" >/dev/null 2>&1; then
    echo "Healthcheck: OK"
  else
    echo "Healthcheck: DOWN"
  fi
}

print_logs_tail() {
  echo ""
  echo "--- Ultime righe log ---"
  if [[ -f "${LOG_FILE}" ]]; then
    tail -n 20 "${LOG_FILE}"
  else
    echo "(log non trovato: ${LOG_FILE})"
  fi
}

print_header

PID_FILE_VALUE="$(pid_from_file)"
PORT_PID="$(process_on_port)"

if is_running_pid "${PID_FILE_VALUE}"; then
  echo "PID file: ${PID_FILE_VALUE} (running)"
else
  if [[ -n "${PID_FILE_VALUE}" ]]; then
    echo "PID file: ${PID_FILE_VALUE} (stale/non running)"
  else
    echo "PID file: assente"
  fi
fi

if [[ -n "${PORT_PID}" ]]; then
  echo "Processo in ascolto sulla porta: ${PORT_PID}"
else
  echo "Processo in ascolto sulla porta: nessuno"
fi

print_health
print_logs_tail
