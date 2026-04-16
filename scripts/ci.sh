#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

if command -v go >/dev/null 2>&1; then
  GO_BIN="$(command -v go)"
elif [[ -x "${ROOT_DIR}/.tools/go/bin/go" ]]; then
  GO_BIN="${ROOT_DIR}/.tools/go/bin/go"
else
  echo "Go non trovato." >&2
  exit 1
fi

echo "=== CI Pipeline ==="
echo "Go: $("${GO_BIN}" version)"
echo ""

PASS=0
FAIL=0

run_step() {
  local name="$1"
  shift
  echo "--- ${name} ---"
  if "$@"; then
    echo "PASS: ${name}"
    ((PASS++))
  else
    echo "FAIL: ${name}"
    ((FAIL++))
  fi
  echo ""
}

run_step "go vet" "${GO_BIN}" vet ./...
run_step "go test" "${GO_BIN}" test ./...
run_step "go test -race" "${GO_BIN}" test -race ./...
run_step "gofmt check" bash -c '
  UNFORMATTED=$("'"${GO_BIN}"'" fmt -l . 2>/dev/null | grep -v "^\.tools/" || true)
  if [[ -n "${UNFORMATTED}" ]]; then
    echo "File non formattati:"
    echo "${UNFORMATTED}"
    exit 1
  fi
'
run_step "go build (app)" "${GO_BIN}" build -o /dev/null .
run_step "go build (license-server)" "${GO_BIN}" build -o /dev/null ./cmd/license-server
run_step "go build (license-keygen)" "${GO_BIN}" build -o /dev/null ./cmd/license-keygen
run_step "go build (license-token-check)" "${GO_BIN}" build -o /dev/null ./cmd/license-token-check

echo "==========================="
echo "Risultato: ${PASS} passati, ${FAIL} falliti"
echo "==========================="

if [[ "${FAIL}" -gt 0 ]]; then
  exit 1
fi
