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

derived_prompt = os.path.join(repo_root, data['orchestrator']['derived_prompt'])
snapshot_file = os.path.join(repo_root, data['orchestrator']['snapshot_file'])
opencode_config = os.path.expanduser(data['orchestrator']['opencode_config'])
agent_key = data['orchestrator']['agent_key']

print(f'DERIVED_PROMPT={q(derived_prompt)}')
print(f'SNAPSHOT_FILE={q(snapshot_file)}')
print(f'OPENCODE_CONFIG={q(opencode_config)}')
print(f'AGENT_KEY={q(agent_key)}')
print('TARGET_DIRS=(' + ' '.join(q(os.path.expanduser(path)) for path in data['skills']['targets']) + ')')
print('PRUNE_SKILLS=(' + ' '.join(q(skill) for skill in data['skills']['prune']) + ')')
print('KEEP_SKILLS=(' + ' '.join(q(skill) for skill in data['skills']['keep']) + ')')
PY
)"

[[ -f "${DERIVED_PROMPT}" ]] || fail "Derived prompt not found: ${DERIVED_PROMPT}"
[[ -f "${OPENCODE_CONFIG}" ]] || fail "OpenCode config not found: ${OPENCODE_CONFIG}"

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

redirect_output="$(${PYTHON_CMD} - "${OPENCODE_CONFIG}" "${DERIVED_PROMPT}" "${SNAPSHOT_FILE}" "${AGENT_KEY}" "${POLICY_FILE}" <<'PY'
import json
import os
import sys
import tempfile

config_path, derived_prompt, snapshot_file, agent_key, policy_path = sys.argv[1:6]

with open(config_path, 'r', encoding='utf-8') as fh:
    data = json.load(fh)

with open(policy_path, 'r', encoding='utf-8') as fh:
    policy = json.load(fh)

agents = data.get('agent')
if not isinstance(agents, dict):
    raise SystemExit('OpenCode config does not contain an agent map')

agent = agents.get(agent_key)
if not isinstance(agent, dict):
    raise SystemExit(f'OpenCode config is missing agent {agent_key!r}')

desired_prompt = '{file:' + derived_prompt + '}'
current_prompt = agent.get('prompt')

snapshot_status = 'unchanged'
config_status = 'unchanged'
config_changed = False

def capture_content(prompt_value):
    if not isinstance(prompt_value, str):
        return None
    if prompt_value == desired_prompt:
        return None
    if prompt_value.startswith('{file:') and prompt_value.endswith('}'):
        candidate = prompt_value[len('{file:'):-1]
        candidate = os.path.expanduser(candidate)
        if not os.path.isabs(candidate):
            candidate = os.path.abspath(candidate)
        if os.path.isfile(candidate):
            with open(candidate, 'r', encoding='utf-8') as fh:
                return fh.read()
        return None
    return prompt_value

content = capture_content(current_prompt)
if content and content.strip():
    os.makedirs(os.path.dirname(snapshot_file), exist_ok=True)
    with open(snapshot_file, 'w', encoding='utf-8') as fh:
        fh.write(content)
        if not content.endswith('\n'):
            fh.write('\n')
    snapshot_status = 'updated'

if current_prompt != desired_prompt:
    agent['prompt'] = desired_prompt
    config_changed = True

for override in policy.get('agent_overrides', []):
    key = override['key']
    model = override['model']
    variant = override.get('variant')
    agents[key] = agents.get(key, {})
    if not isinstance(agents[key], dict):
        agents[key] = {}
    agents[key]['model'] = model
    if variant:
        agents[key]['variant'] = variant
    config_changed = True
    print(f'  agent override {key} -> {model}' + (f' ({variant})' if variant else ''), file=sys.stderr)

if config_changed:
    os.makedirs(os.path.dirname(config_path), exist_ok=True)
    fd, temp_path = tempfile.mkstemp(prefix='opencode.', suffix='.json', dir=os.path.dirname(config_path))
    with os.fdopen(fd, 'w', encoding='utf-8') as fh:
        json.dump(data, fh, indent=2, ensure_ascii=False)
        fh.write('\n')
    os.replace(temp_path, config_path)
    config_status = 'updated'

print(f'SNAPSHOT_STATUS={snapshot_status}')
print(f'CONFIG_STATUS={config_status}')
print(f'DESIRED_PROMPT={desired_prompt}')
PY
)"

eval "${redirect_output}"

info "- upstream prompt snapshot: ${SNAPSHOT_STATUS}"
info "- OpenCode prompt redirect: ${CONFIG_STATUS}"
info "- desired prompt reference: ${DESIRED_PROMPT}"
info "Done. Restart OpenCode if opencode.json changed."
