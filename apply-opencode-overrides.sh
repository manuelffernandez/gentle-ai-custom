#!/usr/bin/env bash
set -euo pipefail

SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_DIR="${HOME}/.config/opencode"

mkdir -p "${TARGET_DIR}/skills/commit-planner"
mkdir -p "${TARGET_DIR}/commands"

cp "${SOURCE_DIR}/opencode/skills/commit-planner/SKILL.md" "${TARGET_DIR}/skills/commit-planner/SKILL.md"
cp "${SOURCE_DIR}/opencode/commands/commit-plan.md" "${TARGET_DIR}/commands/commit-plan.md"
cp "${SOURCE_DIR}/opencode/commands/commit-apply.md" "${TARGET_DIR}/commands/commit-apply.md"

printf 'Custom OpenCode overrides applied to %s\n' "${TARGET_DIR}"
printf 'Reminder: if you run gentle-ai sync again, re-run this script.\n'
