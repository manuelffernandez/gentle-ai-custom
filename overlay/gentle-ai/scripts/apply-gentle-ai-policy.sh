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
    else
      info "  already absent ${skill}"
    fi
  done

  missing_keep=()
  for skill in "${KEEP_SKILLS[@]}"; do
    if [[ ! -e "${target_dir}/${skill}" ]]; then
      missing_keep+=("${skill}")
    fi
  done

  if (( ${#missing_keep[@]} > 0 )); then
    info "  warning: keep skills missing in ${target_dir}: $(IFS=', '; echo "${missing_keep[*]}")"
  fi
done

if [[ ! -f "${OPENCODE_CONFIG}" ]]; then
  info "- skip missing OpenCode config: ${OPENCODE_CONFIG}"
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

with open(config_path, 'r', encoding='utf-8') as fh:
    data = json.load(fh)

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
skipped_count = 0

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

for override in policy.get('agent_overrides', []):
    key = override['key']
    model = override['model']
    variant = override.get('variant')
    current = agents.get(key)
    if not isinstance(current, dict):
        current = {}
        agents[key] = current
        print(f'  agent override {key} reset to object before applying model', file=sys.stderr)
    if current.get('model') != model:
        current['model'] = model
        config_changed = True
    if variant and current.get('variant') != variant:
        current['variant'] = variant
        config_changed = True
    print(f'  agent override {key} -> {model}' + (f' ({variant})' if variant else ''), file=sys.stderr)

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

    if prompt == desired_prompt and os.path.isfile(generated_path):
        print(f'  keep {key}: already points to generated overlay prompt', file=sys.stderr)
        continue

    if prompt.startswith('{file:') and prompt.endswith('}'):
        print(f'  skip {key}: prompt is external file ref and no inline content is available', file=sys.stderr)
        skipped_count += 1
        continue

    write_utf8(snapshot_path, prompt)
    sanitized = sanitize_prompt(prompt)
    write_utf8(generated_path, sanitized)
    agent['prompt'] = desired_prompt
    config_changed = True
    generated_count += 1
    print(f'  generated {key} -> {generated_path}', file=sys.stderr)

if config_changed:
    fd, temp_path = tempfile.mkstemp(prefix='opencode.', suffix='.json', dir=os.path.dirname(config_path))
    with os.fdopen(fd, 'w', encoding='utf-8') as fh:
        json.dump(data, fh, indent=2, ensure_ascii=False)
        fh.write('\n')
    os.replace(temp_path, config_path)

print(f'CONFIG_STATUS={"updated" if config_changed else "unchanged"}')
print(f'GENERATED_COUNT={generated_count}')
print(f'SKIPPED_COUNT={skipped_count}')
PY
)"

eval "${redirect_output}"

info "- OpenCode config status: ${CONFIG_STATUS}"
info "- generated orchestrator prompts: ${GENERATED_COUNT}"
info "- skipped orchestrator prompts: ${SKIPPED_COUNT}"
info "Done. Restart OpenCode if opencode.json changed."
