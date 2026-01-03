#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Run gbgen E2E tests using a local GrowthBook stack (Docker Compose).

This starts:
  - MongoDB
  - GrowthBook (UI on :3000, API on :3100 by default)

It configures a deterministic API key for GrowthBook via env:
  E2E_GB_SECRET_API_KEY (default: secret_gbgen_e2e)

Env overrides:
  E2E_GB_PORT_UI=3000
  E2E_GB_PORT_API=3100
  E2E_GB_SECRET_API_KEY=secret_gbgen_e2e
  E2E_GB_SECRET_API_KEY_ROLE=admin

Example:
  ./scripts/test_e2e_local.sh

  E2E_GB_PORT_API=13100 ./scripts/test_e2e_local.sh
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

require_cmd() {
  local cmd="$1"
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "ERROR: required command not found: ${cmd}" >&2
    exit 1
  fi
}

require_cmd docker
require_cmd go
require_cmd curl

compose() {
  if docker compose version >/dev/null 2>&1; then
    docker compose "$@"
    return
  fi
  if command -v docker-compose >/dev/null 2>&1; then
    docker-compose "$@"
    return
  fi
  echo "ERROR: docker compose (plugin) or docker-compose is required" >&2
  exit 1
}

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.e2e.yml"

E2E_GB_PORT_UI="${E2E_GB_PORT_UI:-3000}"
E2E_GB_PORT_API="${E2E_GB_PORT_API:-3100}"
E2E_GB_SECRET_API_KEY="${E2E_GB_SECRET_API_KEY:-secret_gbgen_e2e}"
E2E_GB_SECRET_API_KEY_ROLE="${E2E_GB_SECRET_API_KEY_ROLE:-admin}"

export E2E_GB_PORT_UI E2E_GB_PORT_API E2E_GB_SECRET_API_KEY E2E_GB_SECRET_API_KEY_ROLE

cleanup() {
  echo "==> Stopping local GrowthBook stack" >&2
  compose -f "${COMPOSE_FILE}" down -v >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

cd "${PROJECT_ROOT}"

echo "==> Starting local GrowthBook stack"
compose -f "${COMPOSE_FILE}" up -d --remove-orphans

echo "==> Waiting for GrowthBook API to become healthy (http://localhost:${E2E_GB_PORT_API}/healthcheck)"
deadline=$((SECONDS + 120))
until curl -fsS "http://localhost:${E2E_GB_PORT_API}/healthcheck" >/dev/null 2>&1; do
  if (( SECONDS > deadline )); then
    echo "ERROR: GrowthBook API did not become healthy in time" >&2
    compose -f "${COMPOSE_FILE}" ps >&2 || true
    exit 1
  fi
  sleep 2
done

echo "==> Running gbgen integration tests against local GrowthBook"
GBGEN_API_BASE_URL="http://localhost:${E2E_GB_PORT_API}" \
GBGEN_API_KEY="${E2E_GB_SECRET_API_KEY}" \
  bash scripts/test_integration.sh

echo "==> E2E tests passed"


