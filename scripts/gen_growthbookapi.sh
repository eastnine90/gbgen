#!/usr/bin/env bash
set -euo pipefail

REPO_URL="https://github.com/growthbook/growthbook.git"
OPENAPI_DIR="packages/back-end/src/api/openapi"
OPENAPI_ENTRYPOINT="${OPENAPI_DIR}/openapi.yaml"

OUT_DIR="internal/growthbookapi"
PKG_NAME="growthbookapi"

# Default generator image. Override if needed:
#   OPENAPI_GENERATOR_IMAGE=openapitools/openapi-generator-cli:v7.7.0 ./scripts/gen_growthbookapi.sh
OPENAPI_GENERATOR_IMAGE="${OPENAPI_GENERATOR_IMAGE:-openapitools/openapi-generator-cli:latest}"

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
# Keep the temp workspace inside the repo root (Docker Desktop file sharing friendly),
# but avoid creating a persistent parent directory like ".tmp/".
TMP_DIR="$(mktemp -d -p "${PROJECT_ROOT}" .gen_growthbookapi.XXXXXX)"
cleanup() {
  # Allow keeping temp files for debugging:
  #   KEEP_TMP=1 ./scripts/gen_growthbookapi.sh
  if [[ "${KEEP_TMP:-}" == "1" ]]; then
    echo "==> KEEP_TMP=1 set; leaving temp workspace: ${TMP_DIR}" >&2
    return 0
  fi

  if [[ -n "${TMP_DIR:-}" && -d "${TMP_DIR}" ]]; then
    rm -rf "${TMP_DIR}"
  fi
}
trap cleanup EXIT INT TERM

require_cmd() {
  local cmd="$1"
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "ERROR: required command not found: ${cmd}" >&2
    exit 1
  fi
}

require_cmd git
require_cmd docker
require_cmd go

echo "==> Preparing temp workspace: ${TMP_DIR}"

echo "==> Sparse-checkout GrowthBook OpenAPI directory from ${REPO_URL}"
git -c advice.detachedHead=false clone \
  --filter=blob:none \
  --no-checkout \
  --depth 1 \
  "${REPO_URL}" \
  "${TMP_DIR}/growthbook"

pushd "${TMP_DIR}/growthbook" >/dev/null
git sparse-checkout init --cone
git sparse-checkout set "${OPENAPI_DIR}"
git checkout --quiet

if [[ ! -f "${OPENAPI_ENTRYPOINT}" ]]; then
  echo "ERROR: OpenAPI entrypoint not found at ${OPENAPI_ENTRYPOINT}" >&2
  exit 1
fi

popd >/dev/null

echo '==> Bundling GrowthBook OpenAPI into a single file (resolve split specs + path/operation $ref)'
BUNDLED_SPEC="${TMP_DIR}/openapi.bundled.yaml"
docker run --rm \
  -u "$(id -u):$(id -g)" \
  -v "${TMP_DIR}:/work" \
  -w /work \
  node:20-alpine sh -lc \
  "npx --yes @apidevtools/swagger-cli@latest bundle --type yaml --outfile /work/openapi.bundled.yaml /work/growthbook/${OPENAPI_ENTRYPOINT}"

echo "==> Prefiltering bundled OpenAPI to features-only (paths + referenced components)"
MIN_SPEC="${TMP_DIR}/openapi.features.min.yaml"
(cd "${PROJECT_ROOT}/tools/openapi_prefilter" && go run . \
  -in "${BUNDLED_SPEC}" \
  -out "${MIN_SPEC}" \
  -tag "features")

echo "==> Generating Go client via openapi-generator (${OPENAPI_GENERATOR_IMAGE})"
GEN_OUT="${TMP_DIR}/generated"
rm -rf "${GEN_OUT}"
mkdir -p "${GEN_OUT}"

# openapi-generator will create a full client skeleton. We only vendor the Go sources into internal/.
docker run --rm \
  -u "$(id -u):$(id -g)" \
  -v "${TMP_DIR}:/tmp" \
  "${OPENAPI_GENERATOR_IMAGE}" generate \
  --skip-validate-spec \
  -i "/tmp/openapi.features.min.yaml" \
  -g go \
  -o "/tmp/generated" \
  --additional-properties "packageName=${PKG_NAME},withGoMod=false,useOneOfDiscriminatorLookup=true,enumClassPrefix=true" \
  --global-property "apis=Features,models,supportingFiles,apiTests=false,modelTests=false,apiDocs=false,modelDocs=false"

echo "==> Copying generated .go sources into ${OUT_DIR}"
rm -rf "${OUT_DIR}"
mkdir -p "${OUT_DIR}"
find "${GEN_OUT}" -type f -name '*.go' \
  ! -path '*/test/*' \
  ! -path '*/docs/*' \
  -print0 | while IFS= read -r -d '' f; do
  rel="${f#${GEN_OUT}/}"
  dest_dir="${OUT_DIR}/$(dirname -- "${rel}")"
  mkdir -p "${dest_dir}"
  cp "${f}" "${dest_dir}/"
done

echo "==> Done"

