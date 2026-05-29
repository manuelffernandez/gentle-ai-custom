#!/usr/bin/env bash

set -euo pipefail

fail() {
  printf 'ERROR: %s\n' "$1" >&2
  exit 1
}

info() {
  printf '%s\n' "$1"
}

PYTHON_CMD="${PYTHON:-python3}"
command -v "${PYTHON_CMD}" >/dev/null 2>&1 || fail "python3 is required to apply the Gentle AI overlay policy"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OVERLAY_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
POLICY_FILE="${OVERLAY_ROOT}/policy/gentle-ai-policy.json"

[[ -f "${POLICY_FILE}" ]] || fail "Policy file not found: ${POLICY_FILE}"

eval "$(${PYTHON_CMD} - "${POLICY_FILE}" "${REPO_ROOT}" <<'PY'
import json
import os
import shlex
import sys

policy_path = sys.argv[1]
repo_root = sys.argv[2]

with open(policy_path, 'r', encoding='utf-8') as fh:
    data = json.load(fh)

def q(value: str) -> str:
    return shlex.quote(value)

opencode = data['opencode']
print(f'OPENCODE_CONFIG={q(os.path.expanduser(opencode["config_path"]))}')
print(f'GENERATED_DIR={q(os.path.expanduser(opencode["generated_orchestrators_dir"]))}')
print(f'SNAPSHOT_DIR={q(os.path.join(repo_root, opencode["orchestrator_snapshot_dir"]))}')
print('TARGET_DIRS=(' + ' '.join(q(os.path.expanduser(path)) for path in data['skills']['targets']) + ')')
print('PRUNE_SKILLS=(' + ' '.join(q(skill) for skill in data['skills']['prune']) + ')')
print('KEEP_SKILLS=(' + ' '.join(q(skill) for skill in data['skills']['keep']) + ')')
PY
)"

info "Applying Gentle AI overlay policy..."

# --- Phase 1: skills pruning ---

PRUNED_COUNT=0
MISSING_KEEP_SUMMARY=()

for target_dir in "${TARGET_DIRS[@]}"; do
  if [[ ! -d "${target_dir}" ]]; then
    info "- skip missing skills dir: ${target_dir}"
    continue
  fi

  info "- pruning unwanted skills in ${target_dir}"
  for skill in "${PRUNE_SKILLS[@]}"; do
    skill_path="${target_dir}/${skill}"
    if [[ -e "${skill_path}" ]]; then
      rm -rf "${skill_path}"
      info "  removed ${skill}"
      PRUNED_COUNT=$((PRUNED_COUNT + 1))
    else
      info "  already absent ${skill}"
    fi
  done

  for skill in "${KEEP_SKILLS[@]}"; do
    if [[ ! -e "${target_dir}/${skill}" ]]; then
      MISSING_KEEP_SUMMARY+=("${target_dir} -> ${skill}")
    fi
  done
done

# --- Phase 2: OpenCode config ---

if [[ ! -f "${OPENCODE_CONFIG}" ]]; then
  info "- skip missing OpenCode config: ${OPENCODE_CONFIG}"
  info ""
  info "Summary:"
  info "  skills pruned this run: ${PRUNED_COUNT}"
  if (( ${#MISSING_KEEP_SUMMARY[@]} > 0 )); then
    info "  WARNING — keep skills missing (expected but absent):"
    for entry in "${MISSING_KEEP_SUMMARY[@]}"; do
      info "    - ${entry}"
    done
  fi
  info "Done."
  exit 0
fi

redirect_output="$(${PYTHON_CMD} - "${OPENCODE_CONFIG}" "${GENERATED_DIR}" "${SNAPSHOT_DIR}" "${POLICY_FILE}" <<'PY'
import json
import os
import re
import sys
import tempfile

config_path, generated_dir, snapshot_dir, policy_path = sys.argv[1:5]

# --- Load OpenCode config with explicit validation ---
try:
    with open(config_path, 'r', encoding='utf-8') as fh:
        data = json.load(fh)
except json.JSONDecodeError as e:
    raise SystemExit(
        f'OpenCode config at {config_path} is not valid JSON: {e}. '
        f'Restore it from a backup under ~/.gentle-ai/backups/ or re-run `gentle-ai sync` to regenerate it.'
    )
except OSError as e:
    raise SystemExit(f'Cannot read OpenCode config at {config_path}: {e}')

with open(policy_path, 'r', encoding='utf-8') as fh:
    policy = json.load(fh)

opencode = policy['opencode']
sanitizer = policy['sanitizer']
agents = data.get('agent')
if not isinstance(agents, dict):
    raise SystemExit('OpenCode config does not contain an agent map')

required_markers = sanitizer.get('required_markers', [])
forbidden_markers = sanitizer.get('forbidden_markers', [])
orchestrator_keys = set(opencode.get('orchestrator_agent_keys', []))
orchestrator_prefixes = tuple(opencode.get('orchestrator_agent_prefixes', []))

config_changed = False
generated_count = 0
recovered_count = 0
skipped_count = 0
snapshot_new = 0
snapshot_changed = 0
snapshot_unchanged = 0
topology_warnings = []
written_orchestrator_keys = set()

def is_orchestrator(agent_key: str) -> bool:
    return agent_key in orchestrator_keys or agent_key.startswith(orchestrator_prefixes)

def write_utf8(path: str, content: str) -> None:
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w', encoding='utf-8', newline='\n') as fh:
        fh.write(content)
        if not content.endswith('\n'):
            fh.write('\n')

def remove_block(text: str, pattern: str, label: str) -> str:
    """Remove a multi-line block using MULTILINE + DOTALL (dot matches newlines)."""
    new_text, count = re.subn(pattern, '', text, flags=re.MULTILINE | re.DOTALL)
    if count == 0:
        raise ValueError(f'missing expected block: {label}')
    return new_text

def remove_line(text: str, pattern: str, label: str) -> str:
    """Remove a single-line pattern using MULTILINE only (dot does NOT match newlines)."""
    new_text, count = re.subn(pattern, '', text, flags=re.MULTILINE)
    if count == 0:
        raise ValueError(f'missing expected line: {label}')
    return new_text

def replace_once(text: str, old: str, new: str, label: str) -> str:
    if old not in text:
        raise ValueError(f'missing expected text: {label}')
    return text.replace(old, new, 1)

def sanitize_prompt(text: str) -> str:
    for marker in required_markers:
        if marker not in text:
            raise ValueError(f'missing required marker before sanitizing: {marker}')

    text = replace_once(
        text,
        '3. **Chained PR strategy**: `auto-forecast`, `ask-always`, `single-pr-default`, or `force-chained`.\n4. **Review budget**: maximum changed lines before stopping for reviewer-burden approval.\n',
        '',
        'preflight PR/review choices'
    )
    text = replace_once(
        text,
        'Reply with "use recommended" or with codes like: A1, B1, C1, D1.',
        'Reply with "use recommended" or with codes like: A1, B1.',
        'english preflight codes'
    )
    text = replace_once(
        text,
        'Respondé con "usar recomendado" o con códigos como: A1, B1, C1, D1.',
        'Respondé con "usar recomendado" o con códigos como: A1, B1.',
        'spanish preflight codes'
    )
    text = remove_block(
        text,
        r'^C\. PRs\n.*?^   D3 Other: ask for the number afterwards\.\n',
        'english PR/review prompt block'
    )
    text = remove_block(
        text,
        r'^C\. PRs\n.*?^   D3 Otro: preguntar el número después\.\n',
        'spanish PR/review prompt block'
    )
    text = remove_line(text, r'^- PRs:.*\n', 'PR answer mapping')
    text = remove_line(text, r'^- Review:.*\n', 'review answer mapping')
    text = replace_once(
        text,
        'If the user explicitly provided all four choices in the current conversation, summarize them as the session preflight block and continue.',
        'If the user explicitly provided both choices in the current conversation, summarize them as the session preflight block and continue.',
        'all four choices wording'
    )
    text = remove_block(text, r'^### Delivery Strategy\n.*?(?=^### Chain Strategy\n|^### Dependency Graph\n)', 'Delivery Strategy section')
    text = remove_block(text, r'^### Chain Strategy\n.*?(?=^### Dependency Graph\n)', 'Chain Strategy section')
    text = remove_block(text, r'^### Review Workload Guard \(MANDATORY\)\n.*?(?=^<!-- gentle-ai:sdd-model-assignments -->\n)', 'Review Workload Guard section')
    text = replace_once(
        text,
        '3. The user explicitly asked to apply/continue implementation, or the prior SDD planning phase completed and the orchestrator has passed the review workload guard.',
        '3. The user explicitly asked to apply/continue implementation, or the prior SDD planning phase completed.',
        'apply routing review-workload clause'
    )

    for marker in required_markers:
        if marker not in text:
            raise ValueError(f'missing required marker after sanitizing: {marker}')
    for marker in forbidden_markers:
        if marker in text:
            raise ValueError(f'forbidden marker still present after sanitizing: {marker}')
    return text

# --- Apply agent overrides ---
# Snapshot original agent keys BEFORE the override loop creates any stubs,
# so topology drift checks can tell which override targets had to be invented.
original_agent_keys = set(agents.keys())
created_overrides = []

for override in policy.get('agent_overrides', []):
    key = override['key']
    model = override['model']
    variant = override.get('variant')
    current = agents.get(key)
    if not isinstance(current, dict):
        current = {}
        agents[key] = current
        created_overrides.append(key)
        print(f'  agent override {key} reset to object before applying model', file=sys.stderr)
    if current.get('model') != model:
        current['model'] = model
        config_changed = True
    if variant and current.get('variant') != variant:
        current['variant'] = variant
        config_changed = True
    print(f'  agent override {key} -> {model}' + (f' ({variant})' if variant else ''), file=sys.stderr)

# --- Topology drift checks (non-fatal warnings) ---
orchestrators_in_config = {k for k in original_agent_keys if is_orchestrator(k)}

# Orchestrators present via prefix match but NOT in the explicit keys list.
# These get sanitized silently — surface them so the maintainer notices new
# orchestrators that may need explicit policy entries.
for key in sorted(orchestrators_in_config - orchestrator_keys):
    msg = f'unknown orchestrator matched by prefix only: {key}'
    topology_warnings.append(msg)
    print(f'  topology: {msg}', file=sys.stderr)

# Orchestrators expected by policy but absent from upstream — e.g. agent was
# renamed/removed upstream and the maintainer hasn't updated policy yet.
for key in sorted(orchestrator_keys - original_agent_keys):
    msg = f'expected orchestrator missing from opencode.json: {key}'
    topology_warnings.append(msg)
    print(f'  topology: {msg}', file=sys.stderr)

# agent_overrides that targeted keys that didn't exist upstream — already
# auto-created above; surface as topology drift so the maintainer can review.
for key in sorted(created_overrides):
    msg = f'agent_override target was missing from upstream (created): {key}'
    topology_warnings.append(msg)
    print(f'  topology: {msg}', file=sys.stderr)

# --- Generate orchestrator overlays ---
os.makedirs(generated_dir, exist_ok=True)
os.makedirs(snapshot_dir, exist_ok=True)

for key in sorted(agents.keys()):
    if not is_orchestrator(key):
        continue
    agent = agents.get(key)
    if not isinstance(agent, dict):
        print(f'  skip {key}: agent entry is not an object', file=sys.stderr)
        skipped_count += 1
        continue
    prompt = agent.get('prompt')
    if not isinstance(prompt, str) or not prompt.strip():
        print(f'  skip {key}: prompt missing or not a string', file=sys.stderr)
        skipped_count += 1
        continue

    generated_path = os.path.join(generated_dir, f'{key}.overlay.md')
    desired_prompt = '{file:' + generated_path + '}'
    snapshot_path = os.path.join(snapshot_dir, f'{key}.last.md')

    # Fully-applied state — already pointing at our generated overlay file,
    # and that file exists on disk. Nothing to do.
    if prompt == desired_prompt and os.path.isfile(generated_path):
        print(f'  keep {key}: already points to generated overlay prompt', file=sys.stderr)
        written_orchestrator_keys.add(key)
        continue

    recovered_from_snapshot = False
    inline_prompt = None

    if prompt.startswith('{file:') and prompt.endswith('}'):
        # Prompt is a file reference but either:
        #   - the target file is missing (e.g. user wiped ~/.config/opencode/prompts/), OR
        #   - the reference points at a path different from our desired one.
        # Recover from the snapshot if available; fail loud if no snapshot exists.
        if not os.path.isfile(snapshot_path):
            raise SystemExit(
                f'broken state for orchestrator {key!r}: opencode.json prompt is {prompt!r} '
                f'but the target file is missing and no snapshot exists at {snapshot_path}. '
                f'Run `gentle-ai sync` to reset the orchestrator prompt to inline content, '
                f'then re-run this script.'
            )
        with open(snapshot_path, 'r', encoding='utf-8') as fh:
            inline_prompt = fh.read().rstrip('\n')
        recovered_from_snapshot = True
        print(f'  recovering {key} from snapshot (target file missing or path drift)', file=sys.stderr)
    else:
        # Normal path: prompt is inline content captured from upstream.
        inline_prompt = prompt

    # Snapshot drift tracking — only when capturing fresh inline content.
    # Recovery reads FROM the snapshot, so by definition it doesn't drift.
    if recovered_from_snapshot:
        snapshot_status = 'recovered'
    else:
        normalized = inline_prompt if inline_prompt.endswith('\n') else inline_prompt + '\n'
        if os.path.isfile(snapshot_path):
            with open(snapshot_path, 'r', encoding='utf-8') as fh:
                old_snapshot = fh.read()
            if old_snapshot != normalized:
                snapshot_status = 'changed'
                snapshot_changed += 1
            else:
                snapshot_status = 'unchanged'
                snapshot_unchanged += 1
        else:
            snapshot_status = 'new'
            snapshot_new += 1
        write_utf8(snapshot_path, inline_prompt)

    sanitized = sanitize_prompt(inline_prompt)
    write_utf8(generated_path, sanitized)
    if agent.get('prompt') != desired_prompt:
        agent['prompt'] = desired_prompt
        config_changed = True

    written_orchestrator_keys.add(key)
    if recovered_from_snapshot:
        recovered_count += 1
        print(f'  recovered {key} -> {generated_path} (from snapshot)', file=sys.stderr)
    else:
        generated_count += 1
        print(f'  generated {key} -> {generated_path} (snapshot: {snapshot_status})', file=sys.stderr)

# --- Atomic write of opencode.json + post-write verification ---
if config_changed:
    fd, temp_path = tempfile.mkstemp(prefix='opencode.', suffix='.json', dir=os.path.dirname(config_path))
    try:
        with os.fdopen(fd, 'w', encoding='utf-8') as fh:
            json.dump(data, fh, indent=2, ensure_ascii=False)
            fh.write('\n')
        os.replace(temp_path, config_path)
    except Exception:
        try:
            os.unlink(temp_path)
        except OSError:
            pass
        raise

    # Re-read and verify the override values and orchestrator refs actually
    # persisted. This catches serialization bugs and races between processes.
    with open(config_path, 'r', encoding='utf-8') as fh:
        verify_data = json.load(fh)
    verify_agents = verify_data.get('agent') or {}

    for override in policy.get('agent_overrides', []):
        key = override['key']
        expected_model = override['model']
        expected_variant = override.get('variant')
        actual = verify_agents.get(key) or {}
        if actual.get('model') != expected_model:
            raise SystemExit(
                f'post-write verification failed: agent {key!r} model is '
                f'{actual.get("model")!r} after write, expected {expected_model!r}'
            )
        if expected_variant and actual.get('variant') != expected_variant:
            raise SystemExit(
                f'post-write verification failed: agent {key!r} variant is '
                f'{actual.get("variant")!r} after write, expected {expected_variant!r}'
            )

    for key in sorted(written_orchestrator_keys):
        expected_ref = '{file:' + os.path.join(generated_dir, f'{key}.overlay.md') + '}'
        actual = (verify_agents.get(key) or {}).get('prompt')
        if actual != expected_ref:
            raise SystemExit(
                f'post-write verification failed: orchestrator {key!r} prompt is '
                f'{actual!r} after write, expected {expected_ref!r}'
            )
        overlay_path = os.path.join(generated_dir, f'{key}.overlay.md')
        if not os.path.isfile(overlay_path):
            raise SystemExit(
                f'post-write verification failed: overlay file missing for {key!r} at {overlay_path}'
            )

print(f'CONFIG_STATUS={"updated" if config_changed else "unchanged"}')
print(f'GENERATED_COUNT={generated_count}')
print(f'RECOVERED_COUNT={recovered_count}')
print(f'SKIPPED_COUNT={skipped_count}')
print(f'SNAPSHOT_NEW={snapshot_new}')
print(f'SNAPSHOT_CHANGED={snapshot_changed}')
print(f'SNAPSHOT_UNCHANGED={snapshot_unchanged}')
print(f'TOPOLOGY_WARNINGS={len(topology_warnings)}')
PY
)"

eval "${redirect_output}"

info ""
info "Summary:"
info "  OpenCode config status: ${CONFIG_STATUS}"
info "  skills pruned this run: ${PRUNED_COUNT}"
info "  orchestrators generated (fresh): ${GENERATED_COUNT}"
info "  orchestrators recovered from snapshot: ${RECOVERED_COUNT}"
info "  orchestrators skipped: ${SKIPPED_COUNT}"
info "  snapshots — new: ${SNAPSHOT_NEW}, changed: ${SNAPSHOT_CHANGED}, unchanged: ${SNAPSHOT_UNCHANGED}"
info "  topology warnings: ${TOPOLOGY_WARNINGS}"

if (( ${#MISSING_KEEP_SUMMARY[@]} > 0 )); then
  info ""
  info "WARNING — keep skills missing (expected but absent):"
  for entry in "${MISSING_KEEP_SUMMARY[@]}"; do
    info "  - ${entry}"
  done
fi

if (( SNAPSHOT_CHANGED > 0 )); then
  info ""
  info "NOTE: upstream orchestrator prompts drifted. Review with:"
  info "  git diff overlay/gentle-ai/snapshots/"
fi

if (( TOPOLOGY_WARNINGS > 0 )); then
  info ""
  info "NOTE: topology drift detected. Review the topology: warnings above and update policy/intent if needed."
fi

info ""
info "Done. Restart OpenCode if opencode.json changed."
