#!/usr/bin/env bash
set -euo pipefail

SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GO_CMD="${GO:-go}"

command -v "${GO_CMD}" >/dev/null 2>&1 || {
  printf 'ERROR: go is required to run the Gentle AI custom overlay CLI\n' >&2
  exit 1
}

cd "${SOURCE_DIR}"
exec "${GO_CMD}" run ./cmd/gentle-ai-overlay --repo-root "${SOURCE_DIR}" apply-custom "$@"
