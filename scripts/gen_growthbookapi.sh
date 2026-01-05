#!/usr/bin/env bash
set -euo pipefail

DOC_URL="https://github.com/growthbook/growthbook.git"
OPENAPI_DIR="packages/back-end/src/api/openapi"

OUT_FILE="internal/growthbookapi/gen.go"
PKG_NAME="growthbookapi"
TMP_DIR=".tmp"

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
mkdir -p ${TMP_DIR}
cleanup() {
  rm -rf ${TMP_DIR}
}
trap cleanup EXIT INT TERM

require_cmd() {
  local cmd="$1"
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "ERROR: required command not found: ${cmd}" >&2
    exit 1
  fi
}

require_cmd yq
require_cmd curl
require_cmd go

echo "==> Preparing temp workspace: ${TMP_DIR}"

echo "==> Downloading OpenAPI doc from ${DOC_URL}"
curl -o ${TMP_DIR}/openapi.yaml https://api.growthbook.io/api/v1/openapi.yaml

echo "==> Patching OpenAPI doc to make compatible with oapi-codegen"
yq -i '.components.schemas.PaginationFields.properties.nextOffset |= {"type":"integer", "nullable":true}' ${TMP_DIR}/openapi.yaml



echo "==> Generating Go client via oapi-codegen"
GEN_OPENAPI_FILE=${PROJECT_ROOT}/${TMP_DIR}/openapi.yaml go generate ${OUT_FILE}