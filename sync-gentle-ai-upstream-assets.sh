#!/usr/bin/env bash
set -euo pipefail

SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GO_CMD="${GO:-go}"

command -v "${GO_CMD}" >/dev/null 2>&1 || {
  printf 'ERROR: go is required to sync approved Gentle AI upstream assets\n' >&2
  exit 1
}

cd "${SOURCE_DIR}"
exec env GENTLE_AI_CUSTOM_ENTRYPOINT="$(basename "$0")" "${GO_CMD}" run ./cmd/gentle-ai-overlay --repo-root "${SOURCE_DIR}" sync-upstream-assets "$@"
