#!/usr/bin/env python3

import hashlib
import json
import os
import re
import subprocess
import sys
from pathlib import Path


def fail(message: str) -> int:
    print(f"ERROR: {message}", file=sys.stderr)
    return 1


def info(message: str) -> None:
    print(message)


def normalize_lf(text: str) -> str:
    return text.replace("\r\n", "\n").replace("\r", "\n")


def normalize_lf_terminated(text: str) -> str:
    normalized = normalize_lf(text)
    return normalized if normalized.endswith("\n") else normalized + "\n"


def sha256_text(text: str) -> str:
    return hashlib.sha256(normalize_lf_terminated(text).encode("utf-8")).hexdigest()


def read_text(path: Path) -> str:
    try:
        return path.read_text(encoding="utf-8")
    except OSError as exc:
        raise RuntimeError(f"cannot read {path}: {exc}") from exc


def parse_simple_yaml(path: Path) -> dict[str, str]:
    data: dict[str, str] = {}
    for idx, raw_line in enumerate(read_text(path).splitlines(), start=1):
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        if ":" not in raw_line:
            raise RuntimeError(f"invalid metadata line {idx} in {path}: missing ':' separator")
        key, value = raw_line.split(":", 1)
        data[key.strip()] = value.strip()
    return data


def git_output(repo: Path, *args: str, allow_failure: bool = False) -> str:
    proc = subprocess.run(
        ["git", "-C", str(repo), *args],
        capture_output=True,
        text=True,
        check=False,
    )
    if proc.returncode != 0:
        if allow_failure:
            return ""
        stderr = proc.stderr.strip() or proc.stdout.strip() or f"git {' '.join(args)} failed"
        raise RuntimeError(f"{repo}: {stderr}")
    return proc.stdout.strip()


def extract_profile_phase_order(profiles_go: str) -> list[str]:
    match = re.search(r"var\s+profilePhaseOrder\s*=\s*\[]string\s*\{(?P<body>.*?)\n\}", profiles_go, re.S)
    if not match:
        raise RuntimeError("could not locate profilePhaseOrder in upstream profiles.go")
    return re.findall(r'"([^"]+)"', match.group("body"))


def main() -> int:
    repo_root = Path(__file__).resolve().parents[3]
    policy_path = repo_root / "overlay/gentle-ai/policy/gentle-ai-policy.json"

    try:
        policy = json.loads(read_text(policy_path))
    except json.JSONDecodeError as exc:
        return fail(f"policy file is not valid JSON at {policy_path}: {exc}")
    except RuntimeError as exc:
        return fail(str(exc))

    upstream_repo = Path(os.path.expanduser(policy["upstream"]["repo_path"]))
    state_path = repo_root / policy["maintenance"]["state_file"]
    snapshot_path = repo_root / policy["opencode"]["orchestrator_snapshot_dir"] / "gentle-orchestrator.last.md"
    meta_path = repo_root / policy["opencode"]["orchestrator_snapshot_metadata_file"]
    upstream_prompt_path = upstream_repo / policy["upstream"]["orchestrator_prompt_path"]
    upstream_profiles_path = upstream_repo / "internal/components/sdd/profiles.go"
    upstream_inject_path = upstream_repo / "internal/components/sdd/inject.go"

    try:
        if not upstream_repo.is_dir():
            return fail(f"upstream repo not found: {upstream_repo}")
        state = json.loads(read_text(state_path))
        metadata = parse_simple_yaml(meta_path)
        snapshot_text = read_text(snapshot_path)
        upstream_prompt_text = read_text(upstream_prompt_path)
        upstream_profiles_text = read_text(upstream_profiles_path)
        upstream_inject_text = read_text(upstream_inject_path)
    except json.JSONDecodeError as exc:
        return fail(f"state file is not valid JSON at {state_path}: {exc}")
    except RuntimeError as exc:
        return fail(str(exc))

    base_key = policy["opencode"].get("base_orchestrator_key", "gentle-orchestrator")
    expected_phase_csv = ",".join(policy["opencode"]["sdd_phases"])
    expected_metadata = {
        "schema_version": "1",
        "snapshot_file": snapshot_path.name,
        "snapshot_source": "upstream-opencode-inline-asset",
        "state_file": policy["maintenance"]["state_file"],
        "upstream_repo_name": upstream_repo.name,
        "upstream_prompt_rel_path": policy["upstream"]["orchestrator_prompt_path"],
        "upstream_inject_source_rel_path": "internal/components/sdd/inject.go",
        "upstream_profiles_source_rel_path": "internal/components/sdd/profiles.go",
        "last_maintained_version": str(state.get("last_maintained_version", "")),
        "last_maintained_tag": str(state.get("last_maintained_tag", "")),
        "last_maintained_commit": str(state.get("last_maintained_commit", "")),
        "last_reviewed_at": str(state.get("last_reviewed_at", "")),
        "base_orchestrator_key": base_key,
        "profile_orchestrator_prefix": policy["opencode"]["profile_orchestrator_prefix"],
        "profile_phase_order_csv": expected_phase_csv,
        "profile_task_scope_rule": "deny-all-then-allow-suffixed-phases-and-global-jd",
    }

    failures: list[str] = []
    notes: list[str] = []

    actual_snapshot_hash = sha256_text(snapshot_text)
    if metadata.get("snapshot_sha256") != actual_snapshot_hash:
        failures.append(
            f"metadata snapshot_sha256 is {metadata.get('snapshot_sha256')!r}, expected {actual_snapshot_hash!r} from {snapshot_path}"
        )

    for field, expected in expected_metadata.items():
        actual = metadata.get(field)
        if actual != expected:
            failures.append(f"metadata field {field!r} is {actual!r}, expected {expected!r}")

    upstream_head = ""
    upstream_describe = ""
    upstream_exact_tag = ""
    try:
        upstream_head = git_output(upstream_repo, "rev-parse", "HEAD")
        upstream_describe = git_output(upstream_repo, "describe", "--tags", "--always")
        upstream_exact_tag = git_output(upstream_repo, "describe", "--tags", "--exact-match", allow_failure=True)
    except RuntimeError as exc:
        failures.append(f"cannot inspect upstream git state: {exc}")

    last_commit = str(state.get("last_maintained_commit", ""))
    last_tag = str(state.get("last_maintained_tag", ""))
    if upstream_head and last_commit and upstream_head != last_commit:
        notes.append(
            f"upstream HEAD {upstream_head} differs from last maintained commit {last_commit}; prompt/invariant drift checks below show whether the baseline still holds"
        )
    if upstream_exact_tag and last_tag and upstream_exact_tag != last_tag:
        notes.append(
            f"upstream exact tag {upstream_exact_tag} differs from last maintained tag {last_tag}; review state/log if you are closing a new upstream audit"
        )

    prompt_matches = normalize_lf_terminated(snapshot_text) == normalize_lf_terminated(upstream_prompt_text)
    if not prompt_matches:
        failures.append(
            f"base prompt drift detected: {upstream_prompt_path} no longer matches {snapshot_path}; review/update the audited baseline before sync/apply"
        )

    try:
        phase_order = extract_profile_phase_order(upstream_profiles_text)
    except RuntimeError as exc:
        failures.append(str(exc))
        phase_order = []

    if phase_order and phase_order != policy["opencode"]["sdd_phases"]:
        failures.append(
            f"upstream profilePhaseOrder is {phase_order!r}, expected {policy['opencode']['sdd_phases']!r} from policy/metadata"
        )

    if 'const orchPrefix = "sdd-orchestrator-"' not in upstream_profiles_text:
        failures.append("upstream profiles.go no longer declares DetectProfiles prefix 'sdd-orchestrator-'")
    if 'keys = append(keys, "sdd-orchestrator"+suffix)' not in upstream_profiles_text:
        failures.append("upstream ProfileAgentKeys no longer builds profile orchestrator keys from 'sdd-orchestrator'+suffix")

    required_profiles_snippets = [
        'taskPerms := map[string]any{',
        '"*": "deny",',
        'taskPerms[phase+suffix] = "allow"',
        'taskPerms[jd] = "allow"',
    ]
    for snippet in required_profiles_snippets:
        if snippet not in upstream_profiles_text:
            failures.append(f"upstream profile task scoping snippet missing from profiles.go: {snippet!r}")

    required_inject_snippets = [
        'orchestratorRaw, ok := agentsMap["gentle-orchestrator"]',
        'orchestratorMap["prompt"] = assets.MustRead(sddOrchestratorAsset(model.AgentOpenCode))',
    ]
    for snippet in required_inject_snippets:
        if snippet not in upstream_inject_text:
            failures.append(f"upstream inject.go no longer contains expected base orchestrator asset binding snippet: {snippet!r}")

    info("Auditing Gentle AI upstream baseline...")
    info(f"- Repo root: {repo_root}")
    info(f"- Upstream repo: {upstream_repo}")
    if upstream_describe:
        info(f"- Upstream HEAD: {upstream_describe} ({upstream_head})")
    info(f"- Base snapshot: {snapshot_path}")
    info(f"- Base metadata: {meta_path}")
    info("")
    metadata_ok = not any(msg.startswith("metadata field") or msg.startswith("metadata snapshot_sha256") for msg in failures)
    profile_naming_ok = (
        'const orchPrefix = "sdd-orchestrator-"' in upstream_profiles_text
        and 'keys = append(keys, "sdd-orchestrator"+suffix)' in upstream_profiles_text
    )
    info("Summary:")
    info(f"  state/metadata alignment: {'ok' if metadata_ok else 'mismatch'}")
    info(f"  snapshot hash verification: {'ok' if metadata.get('snapshot_sha256') == actual_snapshot_hash else 'mismatch'}")
    info(f"  base prompt drift: {'no' if prompt_matches else 'yes'}")
    info(f"  profile phase order: {'ok' if phase_order == policy['opencode']['sdd_phases'] else 'mismatch'}")
    info(f"  profile orchestrator naming: {'ok' if profile_naming_ok else 'mismatch'}")
    info(f"  profile task scoping invariant: {'ok' if all(snippet in upstream_profiles_text for snippet in required_profiles_snippets) else 'mismatch'}")
    info(f"  base asset injection invariant: {'ok' if all(snippet in upstream_inject_text for snippet in required_inject_snippets) else 'mismatch'}")

    if notes:
        info("")
        for note in notes:
            info(f"NOTE: {note}")

    if failures:
        info("")
        for item in failures:
            info(f"FAIL: {item}")
        info("")
        info("Action:")
        info("1. Review the upstream delta against the committed baseline.")
        info("2. Update `gentle-orchestrator.last.md`, `.meta.yaml`, `upstream-state.json`, docs, and the update log if the new upstream state is accepted.")
        info("3. Run `gentle-ai sync` or reinstall as appropriate, then re-run `bash apply-gentle-ai-custom.sh all` so runtime verification passes.")
        return 1

    info("")
    info("Done. The committed base snapshot and metadata still match the current upstream prompt/invariants.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
