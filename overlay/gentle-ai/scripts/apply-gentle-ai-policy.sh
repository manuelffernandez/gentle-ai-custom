#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
GO_CMD="${GO:-go}"

command -v "${GO_CMD}" >/dev/null 2>&1 || {
  printf 'ERROR: go is required to apply the Gentle AI overlay policy\n' >&2
  exit 1
}

cd "${REPO_ROOT}"
exec env GENTLE_AI_CUSTOM_ENTRYPOINT="$(basename "$0")" "${GO_CMD}" run ./cmd/gentle-ai-overlay --repo-root "${REPO_ROOT}" apply-policy "$@"
