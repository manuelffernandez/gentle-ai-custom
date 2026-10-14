---
name: notion-project-knowledge
description: Retrieve Notion project knowledge, preserve useful knowledge and activity during parent-agent work, and follow explicit project-creation requests.
---

# Notion Project Knowledge

Use the current Notion guides as the source of truth for project
structure and operating procedures.

## Resolve the context

Resolve a project named by the user first through its canonical
route. Otherwise use verified repository identity when the
request concerns it, then clear Notion or conversational context.

For repository-based resolution, identify its exact `owner/name`
or canonical URL from safe available context.

Use that identity within the request's scope for targeted lookup
of optional `Repositorio` or `URL` metadata in relevant Notion
indices and guides. Confirm a unique matching project's canonical
root and guide through the global project index:

https://app.notion.com/p/3dca58c3fefc815fb60ded077b5127d1

Without a unique verified match, mention the limitation briefly
and continue independent work. Unavailable access or search does
not establish that a page is absent.

Use explicit project selection or clear current context to resolve
the canonical route without requiring repository metadata.

Create a project only when the user explicitly requests it,
following the creation procedure in the global project index.

Apply the relevant guide's current scope and instructions to each
operation.

## Retrieve progressively

Identify the specific information missing from the request.
Read the applicable guide and the sources needed to resolve it.

Reuse relevant information already available in the session.
Expand retrieval when a concrete uncertainty still affects the
requested outcome.

Stop when sufficient evidence supports the response or action.
Explain any material limitation in the available evidence.

## Preserve useful knowledge and activity

The parent agent proactively identifies useful new project
knowledge and significant personal activity during the work.
Consider substantive investigations, decisions, execution attempts,
resolved blockers and verified progress as Me activity candidates,
even when they add no new project knowledge and the user did not
ask to record them.

At an appropriate progress boundary, persist that knowledge in
the applicable project and activity evidence in Me, following
their current guides and authorized scope.

Use those guides to determine the destination, evidence criteria,
recording procedure and any required index maintenance.

## Keep recording unobtrusive

Fit automatic recording into a natural pause or completion
boundary while keeping the main task moving.

When automatic persistence is unavailable, continue independent
work. Keep skipped recording opportunities silent.

After an attempted write, briefly include its confirmed outcome
in the normal response. Treat an unconfirmed result as unresolved
and continue the main task.
