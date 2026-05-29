#!/usr/bin/env bash
set -euo pipefail

SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHARED_DIR="${SOURCE_DIR}/shared"
COMMIT_SKILL="${SHARED_DIR}/skills/commit-planner/SKILL.md"
PR_SKILL="${SHARED_DIR}/skills/pr-finalizer/SKILL.md"
PLAN_BODY="${SHARED_DIR}/commands/commit-plan-body.md"
APPLY_BODY="${SHARED_DIR}/commands/commit-apply-body.md"
FAST_BODY="${SHARED_DIR}/commands/commit-fast-body.md"
PR_CREATE_BODY="${SHARED_DIR}/commands/pr-create-body.md"
PR_REGENERATE_BODY="${SHARED_DIR}/commands/pr-regenerate-body.md"
SUPPORTED_TARGETS="opencode claude codex gemini antigravity"

usage() {
  local script_name
  script_name="$(basename "$0")"

  printf 'Usage: %s all | [%s ...]\n' "${script_name}" "${SUPPORTED_TARGETS}"
  printf 'Examples:\n'
  printf '  %s opencode\n' "${script_name}"
  printf '  %s claude codex\n' "${script_name}"
  printf '  %s gemini\n' "${script_name}"
  printf '  %s antigravity\n' "${script_name}"
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
  require_file "${COMMIT_SKILL}"
  require_file "${PR_SKILL}"
  require_file "${PLAN_BODY}"
  require_file "${APPLY_BODY}"
  require_file "${FAST_BODY}"
  require_file "${PR_CREATE_BODY}"
  require_file "${PR_REGENERATE_BODY}"
}

is_supported_target() {
  case "$1" in
    opencode|claude|codex|gemini|antigravity)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

normalize_targets() {
  local target
  local -a result=()
  local seen
  local t

  if [ "$#" -eq 0 ]; then
    usage >&2
    exit 1
  fi

  if [ "$#" -eq 1 ] && { [ "$1" = "-h" ] || [ "$1" = "--help" ]; }; then
    usage >&2
    exit 0
  fi

  if [ "$#" -eq 1 ] && [ "$1" = "all" ]; then
    printf '%s\n' opencode claude codex gemini antigravity
    return 0
  fi

  for target in "$@"; do
    case "${target}" in
      -h|--help)
        usage >&2
        exit 0
        ;;
      all)
        die "Use 'all' by itself, or pass explicit targets only."
        ;;
    esac

    is_supported_target "${target}" || die "Unknown target: ${target}"

    seen=0
    for t in "${result[@]+"${result[@]}"}"; do
      [ "${t}" = "${target}" ] && seen=1 && break
    done
    [ "${seen}" -eq 0 ] && result+=("${target}")
  done

  printf '%s\n' "${result[@]}"
}

install_skill() {
  local target_dir="$1"
  local skill_name="$2"
  local skill_source="$3"

  mkdir -p "${target_dir}/skills/${skill_name}"
  cp "${skill_source}" "${target_dir}/skills/${skill_name}/SKILL.md"
}

render_opencode_command() {
  local target_file="$1"
  local skill_name="$2"
  local mode="$3"
  local command_type="$4"
  local description="$5"
  local body_file="$6"

  mkdir -p "$(dirname "${target_file}")"

  {
    printf '%s\n' '---'
    printf 'description: %s\n' "${description}"
    printf '%s\n' '---' ''
    printf 'Read the skill file at `~/.config/opencode/skills/%s/SKILL.md` FIRST, then follow it exactly.\n\n' "${skill_name}"
    printf '%s\n' 'CONTEXT:' '- Working directory: !`echo -n "$(pwd)"`' '- Current project: !`echo -n "$(basename "$(pwd)")"`'
    printf '%s\n' "- Mode: ${mode}" "- Command type: ${command_type}" ''
    cat "${body_file}"
  } > "${target_file}"
}

render_claude_command() {
  local target_file="$1"
  local skill_name="$2"
  local mode="$3"
  local command_type="$4"
  local description="$5"
  local body_file="$6"

  mkdir -p "$(dirname "${target_file}")"

  {
    printf '%s\n' '---'
    printf 'description: %s\n' "${description}"
    printf '%s\n' 'argument-hint: [optional-context]' 'allowed-tools:' '  - Read' '  - Glob' '  - Bash(git:*)' '  - Bash(gh:*)' '  - Bash(pwd:*)' '  - Bash(basename:*)'
    if [ "${mode}" = 'apply' ] || [ "${mode}" = 'auto' ]; then
      printf '%s\n' 'disable-model-invocation: true'
    fi
    printf '%s\n' '---' ''
    printf 'Read the skill file at `~/.claude/skills/%s/SKILL.md` FIRST, then follow it exactly.\n\n' "${skill_name}"
    printf '%s\n' 'CONTEXT:' '- Working directory: !`pwd`' '- Current project: !`basename "$PWD"`'
    printf '%s\n' "- Mode: ${mode}" "- Command type: ${command_type}" ''
    cat "${body_file}"
  } > "${target_file}"
}

render_codex_prompt() {
  local target_file="$1"
  local skill_name="$2"
  local mode="$3"
  local command_type="$4"
  local description="$5"
  local body_file="$6"

  mkdir -p "$(dirname "${target_file}")"

  {
    printf '%s\n' '---'
    printf 'description: %s\n' "${description}"
    printf '%s\n' 'argument-hint: [optional-context]' 'allowed-tools:' '  - Read' '  - Glob' '  - Bash(git:*)' '  - Bash(gh:*)' '  - Bash(pwd:*)' '  - Bash(basename:*)' '---' ''
    printf 'Read the skill file at `~/.codex/skills/%s/SKILL.md` FIRST, then follow it exactly.\n\n' "${skill_name}"
    printf '%s\n' 'CONTEXT:' '- Working directory: !`pwd`' '- Current project: !`basename "$PWD"`'
    printf '%s\n' "- Mode: ${mode}" "- Command type: ${command_type}" ''
    cat "${body_file}"
  } > "${target_file}"
}

render_gemini_skill() {
  local target_file="$1"
  local skill_name="$2"
  local command_name="$3"
  local mode="$4"
  local command_type="$5"
  local description="$6"
  local body_file="$7"

  mkdir -p "$(dirname "${target_file}")"

  {
    printf '%s\n' '---'
    printf 'name: %s\n' "${command_name}"
    printf 'description: %s\n' "${description}"
    printf '%s\n' '---' ''
    printf 'Read the skill file at `~/.gemini/skills/%s/SKILL.md` FIRST, then follow it exactly.\n\n' "${skill_name}"
    printf '%s\n' 'CONTEXT:' '- Working directory: !`pwd`' '- Current project: !`basename "$PWD"`'
    printf '%s\n' "- Mode: ${mode}" "- Command type: ${command_type}" ''
    cat "${body_file}"
  } > "${target_file}"
}

render_antigravity_skill() {
  local target_file="$1"
  local skill_name="$2"
  local command_name="$3"
  local mode="$4"
  local command_type="$5"
  local description="$6"
  local body_file="$7"

  mkdir -p "$(dirname "${target_file}")"

  {
    printf '%s\n' '---'
    printf 'name: %s\n' "${command_name}"
    printf 'description: %s\n' "${description}"
    printf '%s\n' '---' ''
    printf 'Read the skill file at `~/.gemini/antigravity/skills/%s/SKILL.md` FIRST, then follow it exactly.\n\n' "${skill_name}"
    printf '%s\n' 'CONTEXT:' '- Working directory: !`pwd`' '- Current project: !`basename "$PWD"`'
    printf '%s\n' "- Mode: ${mode}" "- Command type: ${command_type}" ''
    cat "${body_file}"
  } > "${target_file}"
}

apply_opencode() {
  local target_dir="${HOME}/.config/opencode"

  install_skill "${target_dir}" 'commit-planner' "${COMMIT_SKILL}"
  install_skill "${target_dir}" 'pr-finalizer' "${PR_SKILL}"
  render_opencode_command "${target_dir}/commands/commit-plan.md" 'commit-planner' 'plan' 'read-only' 'Propose a post-SDD commit plan without changing git state' "${PLAN_BODY}"
  render_opencode_command "${target_dir}/commands/commit-apply.md" 'commit-planner' 'apply' 'state-changing' 'Execute an approved post-SDD commit plan, or generate one first if missing' "${APPLY_BODY}"
  render_opencode_command "${target_dir}/commands/commit-fast.md" 'commit-planner' 'auto' 'state-changing' 'Generate and execute a commit plan in one shot without approval pause' "${FAST_BODY}"
  render_opencode_command "${target_dir}/commands/pr-create.md" 'pr-finalizer' 'create' 'state-changing' 'Draft a PR from committed changes and optionally create it after approval' "${PR_CREATE_BODY}"
  render_opencode_command "${target_dir}/commands/pr-regenerate.md" 'pr-finalizer' 'regenerate' 'state-changing' 'Regenerate or update an existing PR from the current committed diff after approval' "${PR_REGENERATE_BODY}"

  printf 'Applied OpenCode overlays -> %s\n' "${target_dir}"
}

apply_claude() {
  local target_dir="${HOME}/.claude"

  install_skill "${target_dir}" 'commit-planner' "${COMMIT_SKILL}"
  install_skill "${target_dir}" 'pr-finalizer' "${PR_SKILL}"
  render_claude_command "${target_dir}/commands/commit-plan.md" 'commit-planner' 'plan' 'read-only' 'Propose a post-SDD commit plan without changing git state' "${PLAN_BODY}"
  render_claude_command "${target_dir}/commands/commit-apply.md" 'commit-planner' 'apply' 'state-changing' 'Execute an approved post-SDD commit plan, or generate one first if missing' "${APPLY_BODY}"
  render_claude_command "${target_dir}/commands/commit-fast.md" 'commit-planner' 'auto' 'state-changing' 'Generate and execute a commit plan in one shot without approval pause' "${FAST_BODY}"
  render_claude_command "${target_dir}/commands/pr-create.md" 'pr-finalizer' 'create' 'state-changing' 'Draft a PR from committed changes and optionally create it after approval' "${PR_CREATE_BODY}"
  render_claude_command "${target_dir}/commands/pr-regenerate.md" 'pr-finalizer' 'regenerate' 'state-changing' 'Regenerate or update an existing PR from the current committed diff after approval' "${PR_REGENERATE_BODY}"

  printf 'Applied Claude overlays -> %s\n' "${target_dir}"
}

apply_codex() {
  local target_dir="${HOME}/.codex"

  install_skill "${target_dir}" 'commit-planner' "${COMMIT_SKILL}"
  install_skill "${target_dir}" 'pr-finalizer' "${PR_SKILL}"
  render_codex_prompt "${target_dir}/prompts/commit-plan.md" 'commit-planner' 'plan' 'read-only' 'Propose a post-SDD commit plan without changing git state' "${PLAN_BODY}"
  render_codex_prompt "${target_dir}/prompts/commit-apply.md" 'commit-planner' 'apply' 'state-changing' 'Execute an approved post-SDD commit plan, or generate one first if missing' "${APPLY_BODY}"
  render_codex_prompt "${target_dir}/prompts/commit-fast.md" 'commit-planner' 'auto' 'state-changing' 'Generate and execute a commit plan in one shot without approval pause' "${FAST_BODY}"
  render_codex_prompt "${target_dir}/prompts/pr-create.md" 'pr-finalizer' 'create' 'state-changing' 'Draft a PR from committed changes and optionally create it after approval' "${PR_CREATE_BODY}"
  render_codex_prompt "${target_dir}/prompts/pr-regenerate.md" 'pr-finalizer' 'regenerate' 'state-changing' 'Regenerate or update an existing PR from the current committed diff after approval' "${PR_REGENERATE_BODY}"

  printf 'Applied Codex overlays -> %s\n' "${target_dir}"
}

apply_gemini() {
  local target_dir="${HOME}/.gemini"

  install_skill "${target_dir}" 'commit-planner' "${COMMIT_SKILL}"
  install_skill "${target_dir}" 'pr-finalizer' "${PR_SKILL}"
  render_gemini_skill "${target_dir}/skills/commit-plan/SKILL.md" 'commit-planner' 'commit-plan' 'plan' 'read-only' 'Propose a post-SDD commit plan without changing git state' "${PLAN_BODY}"
  render_gemini_skill "${target_dir}/skills/commit-apply/SKILL.md" 'commit-planner' 'commit-apply' 'apply' 'state-changing' 'Execute an approved post-SDD commit plan, or generate one first if missing' "${APPLY_BODY}"
  render_gemini_skill "${target_dir}/skills/commit-fast/SKILL.md" 'commit-planner' 'commit-fast' 'auto' 'state-changing' 'Generate and execute a commit plan in one shot without approval pause' "${FAST_BODY}"
  render_gemini_skill "${target_dir}/skills/pr-create/SKILL.md" 'pr-finalizer' 'pr-create' 'create' 'state-changing' 'Draft a PR from committed changes and optionally create it after approval' "${PR_CREATE_BODY}"
  render_gemini_skill "${target_dir}/skills/pr-regenerate/SKILL.md" 'pr-finalizer' 'pr-regenerate' 'regenerate' 'state-changing' 'Regenerate or update an existing PR from the current committed diff after approval' "${PR_REGENERATE_BODY}"

  printf 'Applied Gemini overlays -> %s\n' "${target_dir}"
}

apply_antigravity() {
  local target_dir="${HOME}/.gemini/antigravity"

  install_skill "${target_dir}" 'commit-planner' "${COMMIT_SKILL}"
  install_skill "${target_dir}" 'pr-finalizer' "${PR_SKILL}"
  render_antigravity_skill "${target_dir}/skills/commit-plan/SKILL.md" 'commit-planner' 'commit-plan' 'plan' 'read-only' 'Propose a post-SDD commit plan without changing git state' "${PLAN_BODY}"
  render_antigravity_skill "${target_dir}/skills/commit-apply/SKILL.md" 'commit-planner' 'commit-apply' 'apply' 'state-changing' 'Execute an approved post-SDD commit plan, or generate one first if missing' "${APPLY_BODY}"
  render_antigravity_skill "${target_dir}/skills/commit-fast/SKILL.md" 'commit-planner' 'commit-fast' 'auto' 'state-changing' 'Generate and execute a commit plan in one shot without approval pause' "${FAST_BODY}"
  render_antigravity_skill "${target_dir}/skills/pr-create/SKILL.md" 'pr-finalizer' 'pr-create' 'create' 'state-changing' 'Draft a PR from committed changes and optionally create it after approval' "${PR_CREATE_BODY}"
  render_antigravity_skill "${target_dir}/skills/pr-regenerate/SKILL.md" 'pr-finalizer' 'pr-regenerate' 'regenerate' 'state-changing' 'Regenerate or update an existing PR from the current committed diff after approval' "${PR_REGENERATE_BODY}"

  printf 'Applied Antigravity overlays -> %s\n' "${target_dir}"
}

should_apply_gentle_overlay() {
  local target

  for target in "${TARGETS[@]}"; do
    case "${target}" in
      opencode|claude)
        return 0
        ;;
    esac
  done

  return 1
}

_targets_raw=$(normalize_targets "$@") || exit $?
[ -z "${_targets_raw}" ] && exit 0
mapfile -t TARGETS <<< "${_targets_raw}"
validate_sources

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
    gemini)
      apply_gemini
      ;;
    antigravity)
      apply_antigravity
      ;;
  esac
done

if should_apply_gentle_overlay; then
  "${SOURCE_DIR}/overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh"
fi

printf 'Reminder: re-run this script after syncs, upgrades, or managed config refreshes.\n'
