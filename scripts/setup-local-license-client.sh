#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.license-dev.env"
CONFIG_FILE="${ROOT_DIR}/config.json"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "File ${ENV_FILE} non trovato." >&2
  echo "Esegui prima: ./scripts/dev-license-env.sh" >&2
  exit 1
fi

# shellcheck disable=SC1090
source "${ENV_FILE}"

if [[ -z "${NP_LICENSE_SERVER_URL:-}" || -z "${NP_LICENSE_PUBLIC_KEY_B64:-}" ]]; then
  echo "Variabili mancanti in ${ENV_FILE}." >&2
  echo "Servono NP_LICENSE_SERVER_URL e NP_LICENSE_PUBLIC_KEY_B64." >&2
  exit 1
fi

python3 - <<'PY' "${CONFIG_FILE}" "${NP_LICENSE_SERVER_URL}" "${NP_LICENSE_PUBLIC_KEY_B64}"
import json
import os
import sys
from datetime import datetime, timezone

config_path = sys.argv[1]
server_url = sys.argv[2]
public_key = sys.argv[3]

defaults = {
    "app_title": "NeuralPath Tactical Guard",
    "iphone_ip": "172.20.10.1",
    "android_ip": "10.145.250.191",
    "lag_threshold_ms": 100,
    "refresh_interval_ms": 1500,
    "reports_dir": "reports",
    "logs_dir": "logs",
    "mode": "real",
    "trial_days": 7,
    "license_file": "license.key",
    "license_server_url": "",
    "license_public_key": "",
    "first_run_at": datetime.now(timezone.utc).isoformat(),
    "license_activated_at": "0001-01-01T00:00:00Z",
}

cfg = {}
if os.path.exists(config_path):
    with open(config_path, "r", encoding="utf-8") as f:
        raw = f.read().strip()
        if raw:
            cfg = json.loads(raw)

for k, v in defaults.items():
    cfg.setdefault(k, v)

cfg["license_server_url"] = server_url
cfg["license_public_key"] = public_key

with open(config_path, "w", encoding="utf-8") as f:
    json.dump(cfg, f, indent=2)
    f.write("\n")
PY

echo "Configurazione client aggiornata in ${CONFIG_FILE}"
echo "  license_server_url = ${NP_LICENSE_SERVER_URL}"
echo "  license_public_key = ${NP_LICENSE_PUBLIC_KEY_B64}"
