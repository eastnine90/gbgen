#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Run gbgen integration tests (opt-in; uses real GrowthBook API).

Required (env or flags):
  GBGEN_API_BASE_URL   or --api-base-url
  GBGEN_API_KEY        or --api-key

Optional:
  GBGEN_PROJECT_ID           or --project-id
  GBGEN_IT_EXPECT_FEATURE_ID or --expect-feature-id

Examples:
  GBGEN_API_BASE_URL="https://api.growthbook.io" GBGEN_API_KEY="..." \
    go test -tags=integration ./internal/integration

  ./scripts/test_integration.sh --api-base-url "https://api.growthbook.io" --api-key "..." --project-id "..." \
    --expect-feature-id "checkout-redesign"
EOF
}

api_base_url="${GBGEN_API_BASE_URL:-}"
api_key="${GBGEN_API_KEY:-}"
project_id="${GBGEN_PROJECT_ID:-}"
expect_feature_id="${GBGEN_IT_EXPECT_FEATURE_ID:-}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    --api-base-url)
      api_base_url="${2:-}"; shift 2
      ;;
    --api-key)
      api_key="${2:-}"; shift 2
      ;;
    --project-id)
      project_id="${2:-}"; shift 2
      ;;
    --expect-feature-id)
      expect_feature_id="${2:-}"; shift 2
      ;;
    *)
      echo "Unknown arg: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "${api_base_url}" ]]; then
  echo "ERROR: missing GBGEN_API_BASE_URL (or --api-base-url)" >&2
  exit 2
fi
if [[ -z "${api_key}" ]]; then
  echo "ERROR: missing GBGEN_API_KEY (or --api-key)" >&2
  exit 2
fi

export GBGEN_API_BASE_URL="${api_base_url}"
export GBGEN_API_KEY="${api_key}"
if [[ -n "${project_id}" ]]; then
  export GBGEN_PROJECT_ID="${project_id}"
fi
if [[ -n "${expect_feature_id}" ]]; then
  export GBGEN_IT_EXPECT_FEATURE_ID="${expect_feature_id}"
fi

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

go test -tags=integration -count=1 ./internal/integration


