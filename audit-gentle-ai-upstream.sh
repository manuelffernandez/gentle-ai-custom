#!/usr/bin/env bash
set -euo pipefail

SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_CMD="${PYTHON:-python3}"

command -v "${PYTHON_CMD}" >/dev/null 2>&1 || {
  printf 'ERROR: python3 is required to audit the Gentle AI upstream baseline\n' >&2
  exit 1
}

exec "${PYTHON_CMD}" "${SOURCE_DIR}/overlay/gentle-ai/scripts/audit-gentle-ai-upstream.py"
