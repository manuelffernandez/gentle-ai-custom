# Placeholder snapshot — upstream OpenCode gentle-orchestrator prompt

This file is intentionally refreshed by:

- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh`
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.ps1`

Workflow:

1. Read the current `agent.gentle-orchestrator.prompt` from `~/.config/opencode/opencode.json`.
2. If it does **not** already point to the local derived prompt, capture its effective content here.
3. Redirect OpenCode to `overlay/gentle-ai/derived/opencode/gentle-orchestrator.md`.

If this placeholder is still present, run the policy-apply script once after the next `gentle-ai sync` or reinstall.
