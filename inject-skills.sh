#!/usr/bin/env bash
set -euo pipefail

SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHARED_DIR="${SOURCE_DIR}/shared"
SHARED_SKILL="${SHARED_DIR}/skills/commit-planner/SKILL.md"
PLAN_BODY="${SHARED_DIR}/commands/commit-plan-body.md"
APPLY_BODY="${SHARED_DIR}/commands/commit-apply-body.md"
SUPPORTED_TARGETS="opencode claude codex"

usage() {
  local script_name
  script_name="$(basename "$0")"

  printf 'Usage: %s all | [%s ...]\n' "${script_name}" "${SUPPORTED_TARGETS}"
  printf 'Examples:\n'
  printf '  %s opencode\n' "${script_name}"
  printf '  %s claude codex\n' "${script_name}"
  printf '  %s all\n' "${script_name}"
}

die() {
  printf '%s\n' "$1" >&2
  exit 1
}

require_file() {
  local file_path="$1"

  [ -f "${file_path}" ] || die "Missing source: ${file_path}"
}

validate_sources() {
  require_file "${SHARED_SKILL}"
  require_file "${PLAN_BODY}"
  require_file "${APPLY_BODY}"
}

is_supported_target() {
  case "$1" in
    opencode|claude|codex)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

normalize_targets() {
  local target

  if [ "$#" -eq 0 ]; then
    usage >&2
    exit 1
  fi

  if [ "$#" -eq 1 ] && { [ "$1" = "-h" ] || [ "$1" = "--help" ]; }; then
    usage
    exit 0
  fi

  if [ "$#" -eq 1 ] && [ "$1" = "all" ]; then
    printf '%s\n' opencode claude codex
    return 0
  fi

  for target in "$@"; do
    case "${target}" in
      -h|--help)
        usage
        exit 0
        ;;
      all)
        die "Use 'all' by itself, or pass explicit targets only."
        ;;
    esac

    is_supported_target "${target}" || die "Unknown target: ${target}"
    printf '%s\n' "${target}"
  done
}

install_skill() {
  local target_dir="$1"

  mkdir -p "${target_dir}/skills/commit-planner"
  cp "${SHARED_SKILL}" "${target_dir}/skills/commit-planner/SKILL.md"
}

render_opencode_command() {
  local target_file="$1"
  local mode="$2"
  local command_type="$3"
  local description="$4"
  local body_file="$5"

  mkdir -p "$(dirname "${target_file}")"

  {
    printf '%s\n' '---'
    printf 'description: %s\n' "${description}"
    printf '%s\n' 'agent: gentleman' '---' ''
    printf '%s\n' 'Read the skill file at `~/.config/opencode/skills/commit-planner/SKILL.md` FIRST, then follow it exactly.' ''
    printf '%s\n' 'CONTEXT:' '- Working directory: !`echo -n "$(pwd)"`' '- Current project: !`echo -n "$(basename "$(pwd)")"`'
    printf '%s\n' "- Mode: ${mode}" "- Command type: ${command_type}" ''
    cat "${body_file}"
  } > "${target_file}"
}

render_claude_command() {
  local target_file="$1"
  local mode="$2"
  local command_type="$3"
  local description="$4"
  local body_file="$5"

  mkdir -p "$(dirname "${target_file}")"

  {
    printf '%s\n' '---'
    printf 'description: %s\n' "${description}"
    printf '%s\n' 'argument-hint: [optional-context]' 'allowed-tools:' '  - Read' '  - Glob' '  - Bash(git:*)' '  - Bash(pwd:*)' '  - Bash(basename:*)'
    if [ "${mode}" = 'apply' ]; then
      printf '%s\n' 'disable-model-invocation: true'
    fi
    printf '%s\n' '---' ''
    printf '%s\n' 'Read the skill file at `~/.claude/skills/commit-planner/SKILL.md` FIRST, then follow it exactly.' ''
    printf '%s\n' 'CONTEXT:' '- Working directory: !`pwd`' '- Current project: !`basename "$PWD"`'
    printf '%s\n' "- Mode: ${mode}" "- Command type: ${command_type}" ''
    cat "${body_file}"
  } > "${target_file}"
}

render_codex_prompt() {
  local target_file="$1"
  local mode="$2"
  local command_type="$3"
  local description="$4"
  local body_file="$5"

  mkdir -p "$(dirname "${target_file}")"

  {
    printf '%s\n' '---'
    printf 'description: %s\n' "${description}"
    printf '%s\n' 'argument-hint: [optional-context]' 'allowed-tools:' '  - Read' '  - Glob' '  - Bash(git:*)' '  - Bash(pwd:*)' '  - Bash(basename:*)' '---' ''
    printf '%s\n' 'Read the skill file at `~/.codex/skills/commit-planner/SKILL.md` FIRST, then follow it exactly.' ''
    printf '%s\n' 'CONTEXT:' '- Working directory: !`pwd`' '- Current project: !`basename "$PWD"`'
    printf '%s\n' "- Mode: ${mode}" "- Command type: ${command_type}" ''
    cat "${body_file}"
  } > "${target_file}"
}

apply_opencode() {
  local target_dir="${HOME}/.config/opencode"

  install_skill "${target_dir}"
  render_opencode_command "${target_dir}/commands/commit-plan.md" 'plan' 'read-only' 'Propose a post-SDD commit plan without changing git state' "${PLAN_BODY}"
  render_opencode_command "${target_dir}/commands/commit-apply.md" 'apply' 'state-changing' 'Execute an approved post-SDD commit plan, or generate one first if missing' "${APPLY_BODY}"

  printf 'Applied OpenCode overlays -> %s\n' "${target_dir}"
}

apply_claude() {
  local target_dir="${HOME}/.claude"

  install_skill "${target_dir}"
  render_claude_command "${target_dir}/commands/commit-plan.md" 'plan' 'read-only' 'Propose a post-SDD commit plan without changing git state' "${PLAN_BODY}"
  render_claude_command "${target_dir}/commands/commit-apply.md" 'apply' 'state-changing' 'Execute an approved post-SDD commit plan, or generate one first if missing' "${APPLY_BODY}"

  printf 'Applied Claude overlays -> %s\n' "${target_dir}"
}

apply_codex() {
  local target_dir="${HOME}/.codex"

  install_skill "${target_dir}"
  render_codex_prompt "${target_dir}/prompts/commit-plan.md" 'plan' 'read-only' 'Propose a post-SDD commit plan without changing git state' "${PLAN_BODY}"
  render_codex_prompt "${target_dir}/prompts/commit-apply.md" 'apply' 'state-changing' 'Execute an approved post-SDD commit plan, or generate one first if missing' "${APPLY_BODY}"

  printf 'Applied Codex overlays -> %s\n' "${target_dir}"
}

validate_sources
mapfile -t TARGETS < <(normalize_targets "$@")

for target in "${TARGETS[@]}"; do
  case "${target}" in
    opencode)
      apply_opencode
      ;;
    claude)
      apply_claude
      ;;
    codex)
      apply_codex
      ;;
  esac
done

printf 'Reminder: re-run this script after syncs, upgrades, or managed config refreshes.\n'
