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
print(f'REPO_SNAPSHOT_DIR={q(os.path.join(repo_root, opencode["orchestrator_snapshot_dir"]))}')
print(f'LOCAL_SNAPSHOT_DIR={q(os.path.expanduser(opencode["local_orchestrator_snapshot_dir"]))}')
print(f'LOCAL_PROFILES_CONFIG={q(os.path.expanduser(opencode["sdd_profiles_local_config_path"]))}')
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
    info "  WARNING - keep skills missing (expected but absent):"
    for entry in "${MISSING_KEEP_SUMMARY[@]}"; do
      info "    - ${entry}"
    done
  fi
  info "Done."
  exit 0
fi

redirect_output="$(${PYTHON_CMD} - "${OPENCODE_CONFIG}" "${GENERATED_DIR}" "${REPO_SNAPSHOT_DIR}" "${LOCAL_SNAPSHOT_DIR}" "${POLICY_FILE}" "${LOCAL_PROFILES_CONFIG}" <<'PY'
import json
import hashlib
import os
import re
import sys
import tempfile

config_path, generated_dir, repo_snapshot_dir, local_snapshot_dir, policy_path, local_profiles_path = sys.argv[1:7]

def die(msg: str) -> 'NoReturn':
    """Exit with an ERROR: prefix on stderr, matching the shell fail() helper."""
    print(f'ERROR: {msg}', file=sys.stderr)
    sys.exit(1)

# --- Load OpenCode config with explicit validation ---
try:
    with open(config_path, 'r', encoding='utf-8') as fh:
        data = json.load(fh)
except json.JSONDecodeError as e:
    die(
        f'OpenCode config at {config_path} is not valid JSON: {e}. '
        f'Restore it from a backup under ~/.gentle-ai/backups/ or re-run `gentle-ai sync` to regenerate it.'
    )
except OSError as e:
    die(f'Cannot read OpenCode config at {config_path}: {e}')

with open(policy_path, 'r', encoding='utf-8') as fh:
    policy = json.load(fh)

opencode = policy['opencode']
sanitizer = policy['sanitizer']
maintenance = policy['maintenance']
agents = data.get('agent')
if not isinstance(agents, dict):
    die('OpenCode config does not contain an agent map')

required_markers = sanitizer.get('required_markers', [])
forbidden_markers = sanitizer.get('forbidden_markers', [])
orchestrator_keys = set(opencode.get('orchestrator_agent_keys', []))
orchestrator_prefixes = tuple(opencode.get('orchestrator_agent_prefixes', []))
profile_orch_prefix = opencode.get('profile_orchestrator_prefix', '')
sdd_phases = list(opencode.get('sdd_phases', []))
base_orchestrator_key = opencode.get('base_orchestrator_key', 'gentle-orchestrator')
repo_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(policy_path))))
state_file = os.path.join(repo_root, maintenance['state_file'])
repo_snapshot_meta_file = os.path.join(repo_root, opencode['orchestrator_snapshot_metadata_file'])
if not sdd_phases:
    die('policy.opencode.sdd_phases is empty or missing; cannot reconcile SDD profiles')
sdd_phases_set = set(sdd_phases)

config_changed = False
generated_count = 0
recovered_count = 0
kept_count = 0
skipped_count = 0
topology_warnings = []
written_orchestrator_keys = set()
repo_snapshot_counters = {'new': 0, 'changed': 0, 'unchanged': 0}
local_snapshot_counters = {'new': 0, 'changed': 0, 'unchanged': 0}
local_snapshot_migrations = 0
repo_snapshot_backfills = 0

# Profile reconciliation counters
profiles_managed_count = 0
profile_agents_created = 0
profile_agents_updated = 0
profile_agents_unchanged = 0
unmanaged_profiles_warnings = []
repo_snapshot_baseline = None
base_runtime_prompt = None
base_generated_path = None

def is_orchestrator(agent_key: str) -> bool:
    return agent_key in orchestrator_keys or agent_key.startswith(orchestrator_prefixes)

def is_profile_orchestrator(agent_key: str) -> bool:
    return bool(profile_orch_prefix) and agent_key.startswith(profile_orch_prefix)

def profile_name_from_orchestrator_key(agent_key: str) -> str:
    return agent_key[len(profile_orch_prefix):] if profile_orch_prefix and agent_key.startswith(profile_orch_prefix) else ''

def safe_snapshot_key(key: str) -> str:
    """Reject keys that would traverse outside snapshot directories."""
    if not key or '/' in key or '\\' in key or '..' in key.split('/') or '\x00' in key:
        die(f'unsafe agent key for snapshot path: {key!r}')
    return key

def write_utf8(path: str, content: str) -> None:
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w', encoding='utf-8', newline='\n') as fh:
        fh.write(content)
        if not content.endswith('\n'):
            fh.write('\n')

def normalize_lf_terminated(content: str) -> str:
    normalized = content.replace('\r\n', '\n').replace('\r', '\n')
    return normalized if normalized.endswith('\n') else normalized + '\n'

def normalized_sha256(content: str) -> str:
    return hashlib.sha256(normalize_lf_terminated(content).encode('utf-8')).hexdigest()

def parse_simple_yaml(path: str) -> dict:
    parsed = {}
    try:
        with open(path, 'r', encoding='utf-8') as fh:
            for idx, raw_line in enumerate(fh, start=1):
                line = raw_line.strip()
                if not line or line.startswith('#'):
                    continue
                if ':' not in raw_line:
                    die(f'invalid metadata line {idx} in {path}: missing ":" separator')
                key, value = raw_line.split(':', 1)
                parsed[key.strip()] = value.strip()
    except OSError as e:
        die(f'Cannot read audited snapshot metadata at {path}: {e}')
    return parsed

def write_snapshot_with_status(path: str, content: str, counters) -> str:
    normalized = normalize_lf_terminated(content)
    if os.path.isfile(path):
        with open(path, 'r', encoding='utf-8') as fh:
            old_snapshot = fh.read()
        old_snapshot_normalized = old_snapshot.replace('\r\n', '\n')
        if old_snapshot_normalized != normalized:
            status = 'changed'
        else:
            status = 'unchanged'
    else:
        status = 'new'
    counters[status] += 1
    write_utf8(path, content)
    return status

def should_write_repo_snapshot(agent_key: str) -> bool:
    return agent_key in orchestrator_keys

def migrate_repo_snapshot_to_local(agent_key: str, repo_snapshot_path: str, local_snapshot_path: str) -> bool:
    global local_snapshot_migrations
    if os.path.isfile(local_snapshot_path) or not os.path.isfile(repo_snapshot_path):
        return False
    with open(repo_snapshot_path, 'r', encoding='utf-8') as fh:
        legacy_content = fh.read().rstrip('\r\n')
    write_utf8(local_snapshot_path, legacy_content)
    local_snapshot_migrations += 1
    print(f'  migrated snapshot {agent_key} -> {local_snapshot_path} (from repo versioned snapshot)', file=sys.stderr)
    return True

def backfill_repo_snapshot_from_local(agent_key: str, local_snapshot_path: str, repo_snapshot_path: str) -> bool:
    global repo_snapshot_backfills
    if not should_write_repo_snapshot(agent_key):
        return False
    if os.path.isfile(repo_snapshot_path) or not os.path.isfile(local_snapshot_path):
        return False
    with open(local_snapshot_path, 'r', encoding='utf-8') as fh:
        local_content = fh.read().rstrip('\r\n')
    write_utf8(repo_snapshot_path, local_content)
    repo_snapshot_backfills += 1
    print(f'  backfilled repo snapshot {agent_key} -> {repo_snapshot_path} (from local operational snapshot)', file=sys.stderr)
    return True

def remove_block(text: str, pattern: str, label: str) -> str:
    """Remove a multi-line block using MULTILINE + DOTALL (dot matches newlines)."""
    new_text, count = re.subn(pattern, '', text, flags=re.MULTILINE | re.DOTALL)
    if count == 0:
        die(f'missing expected block: {label}')
    return new_text

def remove_line(text: str, pattern: str, label: str) -> str:
    """Remove a single-line pattern using MULTILINE only (dot does NOT match newlines)."""
    new_text, count = re.subn(pattern, '', text, flags=re.MULTILINE)
    if count == 0:
        die(f'missing expected line: {label}')
    return new_text

def replace_once(text: str, old: str, new: str, label: str) -> str:
    if old not in text:
        die(f'missing expected text: {label}')
    return text.replace(old, new, 1)

def sanitize_prompt(text: str) -> str:
    for marker in required_markers:
        if marker not in text:
            die(f'missing required marker before sanitizing: {marker}')

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
            die(f'missing required marker after sanitizing: {marker}')
    for marker in forbidden_markers:
        if marker in text:
            die(f'forbidden marker still present after sanitizing: {marker}')
    return text

repo_snapshot_baseline_path = os.path.join(repo_snapshot_dir, f'{base_orchestrator_key}.last.md')
if not os.path.isfile(repo_snapshot_baseline_path):
    die(
        f'audited base snapshot missing for orchestrator {base_orchestrator_key!r} at {repo_snapshot_baseline_path}. '
        f'Restore the committed baseline before re-running apply.'
    )
try:
    with open(repo_snapshot_baseline_path, 'r', encoding='utf-8') as fh:
        repo_snapshot_baseline = fh.read().rstrip('\r\n')
except OSError as e:
    die(f'Cannot read audited base snapshot at {repo_snapshot_baseline_path}: {e}')

try:
    with open(state_file, 'r', encoding='utf-8') as fh:
        state = json.load(fh)
except json.JSONDecodeError as e:
    die(f'state file at {state_file} is not valid JSON: {e}')
except OSError as e:
    die(f'Cannot read state file at {state_file}: {e}')

metadata = parse_simple_yaml(repo_snapshot_meta_file)
expected_metadata = {
    'schema_version': '1',
    'snapshot_file': os.path.basename(repo_snapshot_baseline_path),
    'snapshot_source': 'upstream-opencode-inline-asset',
    'state_file': maintenance['state_file'],
    'upstream_repo_name': os.path.basename(os.path.expanduser(policy['upstream']['repo_path'])),
    'upstream_prompt_rel_path': policy['upstream']['orchestrator_prompt_path'],
    'upstream_inject_source_rel_path': 'internal/components/sdd/inject.go',
    'upstream_profiles_source_rel_path': 'internal/components/sdd/profiles.go',
    'last_maintained_version': str(state.get('last_maintained_version', '')),
    'last_maintained_tag': str(state.get('last_maintained_tag', '')),
    'last_maintained_commit': str(state.get('last_maintained_commit', '')),
    'last_reviewed_at': str(state.get('last_reviewed_at', '')),
    'base_orchestrator_key': base_orchestrator_key,
    'profile_orchestrator_prefix': profile_orch_prefix,
    'profile_phase_order_csv': ','.join(sdd_phases),
    'profile_task_scope_rule': 'deny-all-then-allow-suffixed-phases-and-global-jd',
}
for field, expected_value in expected_metadata.items():
    actual_value = metadata.get(field)
    if actual_value != expected_value:
        die(
            f'audited snapshot metadata mismatch: field {field!r} in {repo_snapshot_meta_file} is '
            f'{actual_value!r}, expected {expected_value!r}. Repair the committed baseline before re-running apply.'
        )
actual_snapshot_hash = normalized_sha256(repo_snapshot_baseline)
if metadata.get('snapshot_sha256') != actual_snapshot_hash:
    die(
        f'audited snapshot metadata mismatch: snapshot_sha256 in {repo_snapshot_meta_file} is '
        f'{metadata.get("snapshot_sha256")!r}, expected {actual_snapshot_hash!r} from {repo_snapshot_baseline_path}. '
        f'Repair the committed baseline before re-running apply.'
    )

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
        # Track every non-dict reset: includes "key was missing entirely" AND
        # "key existed but as a non-object (string/null/list)". Both cases
        # surface as topology drift so the maintainer can review whether the
        # upstream agent shape changed.
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

# --- SDD profile reconciliation (strict, fail-closed) ---
# Contract:
#   - If the local config file does NOT exist: do not touch SDD profiles.
#   - If it exists: parse + validate STRICTLY before any mutation.
#   - For each managed profile: create/update the orchestrator + 10 phase agents
#     with the configured model/variant. Do NOT touch prompts here.
#   - Profiles present in opencode.json but absent from local config are left
#     untouched but surfaced as warnings + counter.
#   - No automatic deletion of unmanaged profiles.

managed_profile_names = set()

if os.path.isfile(local_profiles_path):
    try:
        with open(local_profiles_path, 'r', encoding='utf-8') as fh:
            local_cfg = json.load(fh)
    except json.JSONDecodeError as e:
        die(
            f'local SDD profile config at {local_profiles_path} is not valid JSON: {e}. '
            f'Fix or remove the file before re-running this script.'
        )
    except OSError as e:
        die(f'Cannot read local SDD profile config at {local_profiles_path}: {e}')

    # --- Strict schema validation (V1, no inheritance, no defaults) ---
    if not isinstance(local_cfg, dict):
        die(f'local SDD profile config at {local_profiles_path} must be a JSON object at the top level')
    extra_top = set(local_cfg.keys()) - {'version', 'profiles'}
    if extra_top:
        die(
            f'local SDD profile config at {local_profiles_path} has unexpected top-level fields '
            f'{sorted(extra_top)}; only "version" and "profiles" are allowed'
        )
    if local_cfg.get('version') != 1:
        die(
            f'local SDD profile config at {local_profiles_path} has unsupported "version" '
            f'{local_cfg.get("version")!r}; expected 1'
        )
    profiles_raw = local_cfg.get('profiles')
    if not isinstance(profiles_raw, list) or len(profiles_raw) == 0:
        die(f'local SDD profile config at {local_profiles_path} must contain a non-empty "profiles" array')

    def validate_assignment(label: str, value):
        """Each assignment must be {model: non-empty str, variant: str (may be empty)}."""
        if not isinstance(value, dict):
            die(f'{label}: must be an object with "model" and "variant"')
        extra = set(value.keys()) - {'model', 'variant'}
        if extra:
            die(f'{label}: unexpected fields {sorted(extra)}; only "model" and "variant" are allowed')
        if 'model' not in value:
            die(f'{label}: missing required field "model"')
        if 'variant' not in value:
            die(f'{label}: missing required field "variant" (use "" if the assignment has no variant)')
        if not isinstance(value['model'], str) or value['model'] == '':
            die(f'{label}: field "model" must be a non-empty string')
        if not isinstance(value['variant'], str):
            die(f'{label}: field "variant" must be a string (use "" for no variant)')

    seen_names = set()
    validated_profiles = []
    for idx, profile in enumerate(profiles_raw):
        prefix = f'profiles[{idx}]'
        if not isinstance(profile, dict):
            die(f'{prefix}: must be an object')
        extra = set(profile.keys()) - {'name', 'orchestrator', 'phases'}
        if extra:
            die(f'{prefix}: unexpected fields {sorted(extra)}; only "name", "orchestrator", "phases" are allowed')
        name = profile.get('name')
        if not isinstance(name, str) or name == '':
            die(f'{prefix}: "name" must be a non-empty string')
        if not re.match(r'^[a-z0-9][a-z0-9._-]*$', name):
            die(f'{prefix}: "name" {name!r} must match ^[a-z0-9][a-z0-9._-]*$ to be safe as an agent-key suffix')
        if name in seen_names:
            die(f'{prefix}: duplicate profile name {name!r}')
        seen_names.add(name)

        if 'orchestrator' not in profile:
            die(f'{prefix}: missing required field "orchestrator"')
        validate_assignment(f'{prefix}.orchestrator', profile['orchestrator'])

        if 'phases' not in profile:
            die(f'{prefix}: missing required field "phases"')
        phases = profile['phases']
        if not isinstance(phases, dict):
            die(f'{prefix}.phases: must be an object keyed by SDD phase name')
        phase_keys = set(phases.keys())
        missing = sdd_phases_set - phase_keys
        if missing:
            die(f'{prefix}.phases: missing required phases {sorted(missing)} (no defaults are inherited)')
        unknown = phase_keys - sdd_phases_set
        if unknown:
            die(f'{prefix}.phases: unknown phases {sorted(unknown)}; allowed: {sdd_phases}')
        for phase_name in sdd_phases:
            validate_assignment(f'{prefix}.phases.{phase_name}', phases[phase_name])

        validated_profiles.append({
            'name': name,
            'orchestrator': profile['orchestrator'],
            'phases': phases,
        })

    # --- All validation passed. Now apply. ---
    for profile in validated_profiles:
        name = profile['name']
        managed_profile_names.add(name)
        profiles_managed_count += 1
        # Reconcile orchestrator agent (model/variant only — we do not manage prompts here).
        orch_key = f'sdd-orchestrator-{name}'
        orch_assignment = profile['orchestrator']
        existing = agents.get(orch_key)
        if not isinstance(existing, dict):
            agents[orch_key] = {
                'model': orch_assignment['model'],
                'variant': orch_assignment['variant'],
            }
            profile_agents_created += 1
            config_changed = True
            print(f'  profile {name}: created orchestrator agent {orch_key} (no prompt; run `gentle-ai sync` to materialize)', file=sys.stderr)
        else:
            changed_here = False
            if existing.get('model') != orch_assignment['model']:
                existing['model'] = orch_assignment['model']
                changed_here = True
            if existing.get('variant') != orch_assignment['variant']:
                existing['variant'] = orch_assignment['variant']
                changed_here = True
            if changed_here:
                profile_agents_updated += 1
                config_changed = True
                print(f'  profile {name}: updated orchestrator agent {orch_key} -> {orch_assignment["model"]}'
                      + (f' ({orch_assignment["variant"]})' if orch_assignment['variant'] else ''), file=sys.stderr)
            else:
                profile_agents_unchanged += 1

        # Reconcile each phase agent.
        for phase_name in sdd_phases:
            phase_key = f'{phase_name}-{name}'
            assignment = profile['phases'][phase_name]
            existing = agents.get(phase_key)
            if not isinstance(existing, dict):
                agents[phase_key] = {
                    'model': assignment['model'],
                    'variant': assignment['variant'],
                }
                profile_agents_created += 1
                config_changed = True
                print(f'  profile {name}: created phase agent {phase_key} -> {assignment["model"]}'
                      + (f' ({assignment["variant"]})' if assignment['variant'] else ''), file=sys.stderr)
            else:
                changed_here = False
                if existing.get('model') != assignment['model']:
                    existing['model'] = assignment['model']
                    changed_here = True
                if existing.get('variant') != assignment['variant']:
                    existing['variant'] = assignment['variant']
                    changed_here = True
                if changed_here:
                    profile_agents_updated += 1
                    config_changed = True
                    print(f'  profile {name}: updated phase agent {phase_key} -> {assignment["model"]}'
                          + (f' ({assignment["variant"]})' if assignment['variant'] else ''), file=sys.stderr)
                else:
                    profile_agents_unchanged += 1

    # Detect unmanaged profiles already present in opencode.json (warn-only).
    discovered_profile_names = set()
    for k in original_agent_keys:
        if is_profile_orchestrator(k):
            pn = profile_name_from_orchestrator_key(k)
            if pn:
                discovered_profile_names.add(pn)
    for pn in sorted(discovered_profile_names - managed_profile_names):
        unmanaged_profiles_warnings.append(pn)
        print(f'  unmanaged SDD profile present in opencode.json (left untouched): {pn}', file=sys.stderr)
else:
    print(f'  no local SDD profile config at {local_profiles_path} - SDD profiles untouched', file=sys.stderr)

# --- Topology drift checks (non-fatal warnings) ---
orchestrators_in_config = {k for k in original_agent_keys if is_orchestrator(k)}

# Orchestrators present via prefix match but NOT in the explicit keys list.
# These get sanitized silently — surface them so the maintainer notices new
# orchestrators that may need explicit policy entries.
#
# EXCEPTION: profile-managed orchestrators (sdd-orchestrator-<name>) are
# deliberately not in orchestrator_agent_keys; they are managed via the local
# SDD profile config. Don't warn about them.
for key in sorted(orchestrators_in_config - orchestrator_keys):
    if is_profile_orchestrator(key):
        continue
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
os.makedirs(repo_snapshot_dir, exist_ok=True)
os.makedirs(local_snapshot_dir, exist_ok=True)

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

    safe_key = safe_snapshot_key(key)
    generated_path = os.path.join(generated_dir, f'{safe_key}.overlay.md')
    desired_prompt = '{file:' + generated_path + '}'
    repo_snapshot_path = os.path.join(repo_snapshot_dir, f'{safe_key}.last.md')
    local_snapshot_path = os.path.join(local_snapshot_dir, f'{safe_key}.last.md')

    migrate_repo_snapshot_to_local(key, repo_snapshot_path, local_snapshot_path)
    backfill_repo_snapshot_from_local(key, local_snapshot_path, repo_snapshot_path)

    # Fully-applied state — already pointing at our generated overlay file,
    # and that file exists on disk. Nothing to do.
    if prompt == desired_prompt and os.path.isfile(generated_path):
        if not os.path.isfile(local_snapshot_path):
            die(
                f'local operational snapshot missing for orchestrator {key!r} at {local_snapshot_path}. '
                f'Run `gentle-ai sync` to reset the orchestrator prompt to inline content, '
                f'then re-run this script to capture a fresh snapshot.'
            )
        if should_write_repo_snapshot(key) and not os.path.isfile(repo_snapshot_path):
            backfill_repo_snapshot_from_local(key, local_snapshot_path, repo_snapshot_path)
            if not os.path.isfile(repo_snapshot_path):
                die(
                    f'versioned repo snapshot missing for orchestrator {key!r} at {repo_snapshot_path}. '
                    f'Run `gentle-ai sync` to capture fresh upstream, then re-run this script.'
                )
        print(f'  keep {key}: already points to generated overlay prompt', file=sys.stderr)
        if key == base_orchestrator_key:
            try:
                with open(local_snapshot_path, 'r', encoding='utf-8') as fh:
                    base_runtime_prompt = fh.read().rstrip('\r\n')
            except OSError as e:
                die(f'Cannot read local operational snapshot for audited base orchestrator at {local_snapshot_path}: {e}')
            base_generated_path = generated_path
        written_orchestrator_keys.add(key)
        kept_count += 1
        continue

    recovered_from_snapshot = False
    inline_prompt = None

    if prompt.startswith('{file:') and prompt.endswith('}'):
        # Prompt is a file reference but either:
        #   - the target file is missing (e.g. user wiped ~/.config/opencode/prompts/), OR
        #   - the reference points at a path different from our desired one.
        # Recover from the snapshot if available; fail loud if no snapshot exists.
        if not os.path.isfile(local_snapshot_path):
            migrate_repo_snapshot_to_local(key, repo_snapshot_path, local_snapshot_path)
        if not os.path.isfile(local_snapshot_path):
            if should_write_repo_snapshot(key):
                missing_detail = f'no local operational snapshot exists at {local_snapshot_path} and no repo snapshot exists at {repo_snapshot_path}'
            else:
                missing_detail = f'no local operational snapshot exists at {local_snapshot_path}'
            die(
                f'broken state for orchestrator {key!r}: opencode.json prompt is {prompt!r} '
                f'but the target file is missing and {missing_detail}. '
                f'Run `gentle-ai sync` to reset the orchestrator prompt to inline content, '
                f'then re-run this script.'
            )
        if should_write_repo_snapshot(key) and not os.path.isfile(repo_snapshot_path):
            backfill_repo_snapshot_from_local(key, local_snapshot_path, repo_snapshot_path)
        with open(local_snapshot_path, 'r', encoding='utf-8') as fh:
            inline_prompt = fh.read().rstrip('\r\n')
        recovered_from_snapshot = True
        print(
            f'  WARNING recovering {key} from local snapshot - content may pre-date current upstream; '
            f'run `gentle-ai sync` then re-run this script to capture fresh upstream into the snapshot',
            file=sys.stderr,
        )
    else:
        # Normal path: prompt is inline content captured from upstream.
        inline_prompt = prompt

    # Sanitize FIRST so a failure does not leave a stale-but-overwritten
    # snapshot. The snapshot is only updated after sanitization succeeds.
    sanitized = sanitize_prompt(inline_prompt)

    # Snapshot drift tracking — only when capturing fresh inline content.
    # Recovery reads FROM the snapshot, so by definition it doesn't drift.
    if recovered_from_snapshot:
        snapshot_status = 'recovered'
    else:
        local_snapshot_status = write_snapshot_with_status(local_snapshot_path, inline_prompt, local_snapshot_counters)
        if should_write_repo_snapshot(key):
            repo_snapshot_status = write_snapshot_with_status(repo_snapshot_path, inline_prompt, repo_snapshot_counters)
            snapshot_status = f'local: {local_snapshot_status}, repo: {repo_snapshot_status}'
        else:
            snapshot_status = f'local: {local_snapshot_status}'

    write_utf8(generated_path, sanitized)
    if agent.get('prompt') != desired_prompt:
        agent['prompt'] = desired_prompt
        config_changed = True

    if key == base_orchestrator_key:
        base_runtime_prompt = inline_prompt
        base_generated_path = generated_path

    written_orchestrator_keys.add(key)
    if recovered_from_snapshot:
        recovered_count += 1
        print(f'  recovered {key} -> {generated_path} (from snapshot)', file=sys.stderr)
    else:
        generated_count += 1
        print(f'  generated {key} -> {generated_path} (snapshot: {snapshot_status})', file=sys.stderr)

# --- Atomic write of opencode.json ---
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


# --- Verification from persisted opencode.json ---
with open(config_path, 'r', encoding='utf-8') as fh:
    verify_data = json.load(fh)
verify_agents = verify_data.get('agent') or {}

for override in policy.get('agent_overrides', []):
    key = override['key']
    expected_model = override['model']
    expected_variant = override.get('variant')
    actual = verify_agents.get(key) or {}
    if actual.get('model') != expected_model:
        die(
            f'post-write verification failed: agent {key!r} model is '
            f'{actual.get("model")!r} after write, expected {expected_model!r}'
        )
    if expected_variant and actual.get('variant') != expected_variant:
        die(
            f'post-write verification failed: agent {key!r} variant is '
            f'{actual.get("variant")!r} after write, expected {expected_variant!r}'
        )

for key in sorted(written_orchestrator_keys):
    expected_ref = '{file:' + os.path.join(generated_dir, f'{key}.overlay.md') + '}'
    actual = (verify_agents.get(key) or {}).get('prompt')
    if actual != expected_ref:
        die(
            f'post-write verification failed: orchestrator {key!r} prompt is '
            f'{actual!r} after write, expected {expected_ref!r}'
        )
    overlay_path = os.path.join(generated_dir, f'{key}.overlay.md')
    if not os.path.isfile(overlay_path):
        die(
            f'post-write verification failed: overlay file missing for {key!r} at {overlay_path}'
        )

# Verify profile reconciliation persisted.
for profile_name in sorted(managed_profile_names):
    orch_key = f'sdd-orchestrator-{profile_name}'
    if orch_key not in verify_agents:
        die(
            f'post-write verification failed: profile {profile_name!r} orchestrator agent '
            f'{orch_key!r} missing from {config_path} after write'
        )
    for phase_name in sdd_phases:
        phase_key = f'{phase_name}-{profile_name}'
        if phase_key not in verify_agents:
            die(
                f'post-write verification failed: profile {profile_name!r} phase agent '
                f'{phase_key!r} missing from {config_path} after write'
            )

if base_runtime_prompt is None or base_generated_path is None:
    die(
        f'audited baseline verification failed: orchestrator {base_orchestrator_key!r} was not materialized during apply. '
        f'Run `gentle-ai sync` to restore the inline upstream prompt, then re-run this script.'
    )

if normalize_lf_terminated(base_runtime_prompt) != normalize_lf_terminated(repo_snapshot_baseline):
    die(
        f'audited baseline mismatch for orchestrator {base_orchestrator_key!r}: runtime source prompt does not match '
        f'{repo_snapshot_baseline_path}. Run `bash audit-gentle-ai-upstream.sh` before adopting a new upstream baseline, '
        f'then re-run `gentle-ai sync` and this script.'
    )

expected_base_overlay = sanitize_prompt(repo_snapshot_baseline)
try:
    with open(base_generated_path, 'r', encoding='utf-8') as fh:
        actual_base_overlay = fh.read().rstrip('\r\n')
except OSError as e:
    die(f'Cannot read generated overlay for audited base orchestrator at {base_generated_path}: {e}')

if normalize_lf_terminated(actual_base_overlay) != normalize_lf_terminated(expected_base_overlay):
    die(
        f'audited baseline mismatch for orchestrator {base_orchestrator_key!r}: generated overlay at {base_generated_path} '
        f'does not match the sanitized audited snapshot. Re-run apply after restoring the audited baseline, '
        f'or run `gentle-ai sync` if local runtime state is stale.'
    )

print(f'CONFIG_STATUS={"updated" if config_changed else "unchanged"}')
print(f'GENERATED_COUNT={generated_count}')
print(f'RECOVERED_COUNT={recovered_count}')
print(f'KEPT_COUNT={kept_count}')
print(f'SKIPPED_COUNT={skipped_count}')
print(f'REPO_SNAPSHOT_NEW={repo_snapshot_counters["new"]}')
print(f'REPO_SNAPSHOT_CHANGED={repo_snapshot_counters["changed"]}')
print(f'REPO_SNAPSHOT_UNCHANGED={repo_snapshot_counters["unchanged"]}')
print(f'LOCAL_SNAPSHOT_NEW={local_snapshot_counters["new"]}')
print(f'LOCAL_SNAPSHOT_CHANGED={local_snapshot_counters["changed"]}')
print(f'LOCAL_SNAPSHOT_UNCHANGED={local_snapshot_counters["unchanged"]}')
print(f'LOCAL_SNAPSHOT_MIGRATIONS={local_snapshot_migrations}')
print(f'REPO_SNAPSHOT_BACKFILLS={repo_snapshot_backfills}')
print(f'TOPOLOGY_WARNINGS={len(topology_warnings)}')
print(f'PROFILES_MANAGED={profiles_managed_count}')
print(f'PROFILE_AGENTS_CREATED={profile_agents_created}')
print(f'PROFILE_AGENTS_UPDATED={profile_agents_updated}')
print(f'PROFILE_AGENTS_UNCHANGED={profile_agents_unchanged}')
print(f'UNMANAGED_PROFILES_COUNT={len(unmanaged_profiles_warnings)}')
# Emit unmanaged profile names as a single shell-safe space-separated line.
import shlex as _shlex
print('UNMANAGED_PROFILES_NAMES=' + _shlex.quote(' '.join(unmanaged_profiles_warnings)))
PY
)"

eval "${redirect_output}"

info ""
info "Summary:"
info "  OpenCode config status: ${CONFIG_STATUS}"
info "  skills pruned this run: ${PRUNED_COUNT}"
info "  orchestrators generated (fresh): ${GENERATED_COUNT}"
info "  orchestrators recovered from snapshot: ${RECOVERED_COUNT}"
info "  orchestrators kept (already applied): ${KEPT_COUNT}"
info "  orchestrators skipped: ${SKIPPED_COUNT}"
info "  repo snapshots - new: ${REPO_SNAPSHOT_NEW}, changed: ${REPO_SNAPSHOT_CHANGED}, unchanged: ${REPO_SNAPSHOT_UNCHANGED}"
info "  local snapshots - new: ${LOCAL_SNAPSHOT_NEW}, changed: ${LOCAL_SNAPSHOT_CHANGED}, unchanged: ${LOCAL_SNAPSHOT_UNCHANGED}"
info "  local snapshot migrations from repo: ${LOCAL_SNAPSHOT_MIGRATIONS}"
info "  repo snapshot backfills from local: ${REPO_SNAPSHOT_BACKFILLS}"
info "  topology warnings: ${TOPOLOGY_WARNINGS}"
info "  SDD profiles managed: ${PROFILES_MANAGED}"
info "  SDD profile agents created: ${PROFILE_AGENTS_CREATED}"
info "  SDD profile agents updated: ${PROFILE_AGENTS_UPDATED}"
info "  SDD profile agents unchanged: ${PROFILE_AGENTS_UNCHANGED}"
info "  SDD profiles unmanaged (present in opencode.json, absent from local config): ${UNMANAGED_PROFILES_COUNT}"
info "  audited base baseline verification: ok"

if [[ -n "${UNMANAGED_PROFILES_NAMES}" ]]; then
  info ""
  info "WARNING - unmanaged SDD profiles left untouched (add them to the local SDD profile config to manage):"
  for entry in ${UNMANAGED_PROFILES_NAMES}; do
    info "  - ${entry}"
  done
fi

if (( ${#MISSING_KEEP_SUMMARY[@]} > 0 )); then
  info ""
  info "WARNING - keep skills missing (expected but absent):"
  for entry in "${MISSING_KEEP_SUMMARY[@]}"; do
    info "  - ${entry}"
  done
fi

if (( REPO_SNAPSHOT_CHANGED > 0 )); then
  info ""
  info "NOTE: versioned orchestrator snapshots drifted. Review with:"
  info "  git diff overlay/gentle-ai/snapshots/"
fi

if (( LOCAL_SNAPSHOT_CHANGED > 0 )); then
  info ""
  info "NOTE: local operational orchestrator snapshots drifted under:"
  info "  ${LOCAL_SNAPSHOT_DIR}"
fi

if (( LOCAL_SNAPSHOT_MIGRATIONS > 0 )); then
  info ""
  info "NOTE: migrated ${LOCAL_SNAPSHOT_MIGRATIONS} legacy snapshot(s) from the repo into the local operational snapshot dir."
fi

if (( REPO_SNAPSHOT_BACKFILLS > 0 )); then
  info ""
  info "NOTE: backfilled ${REPO_SNAPSHOT_BACKFILLS} versioned repo snapshot(s) from local operational snapshots."
fi

if (( RECOVERED_COUNT > 0 )); then
  info ""
  info "NOTE: ${RECOVERED_COUNT} orchestrator(s) recovered from snapshot."
  info "  The snapshot content may pre-date the current upstream version."
  info "  Run \`gentle-ai sync\` then re-run this script to capture fresh upstream."
fi

if (( TOPOLOGY_WARNINGS > 0 )); then
  info ""
  info "NOTE: topology drift detected. Review the topology: warnings above and update policy/intent if needed."
fi

info ""
info "Done. Restart OpenCode if opencode.json changed."
