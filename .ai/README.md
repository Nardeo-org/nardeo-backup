# Shared Task Context

Copyright (c) 2026 Nardeo.

Start with [AGENTS.md](../AGENTS.md). This directory holds task-specific requirements and handoffs that are useful across sessions and contributors. Committed context travels with the repository and can be reviewed alongside code changes.

## Where information belongs

| Information | Location |
| --- | --- |
| Repository-wide instructions and entry points | [AGENTS.md](../AGENTS.md) |
| Current implementation boundaries and flow | [docs/architecture.md](../docs/architecture.md) |
| Accepted design decisions and their consequences | [docs/decisions/](../docs/decisions/) |
| Public behavior and migration expectations | [BACKWARD_COMPATIBILITY.md](../BACKWARD_COMPATIBILITY.md) |
| Requirements for a substantial task | `specs/YYYY-MM-DD-topic.md`, based on [the spec template](specs/_template.md) |
| State of an interrupted or multi-session task | `handoffs/issue-N-or-branch.md`, based on [the handoff template](handoffs/_template.md) |
| Assignment, review and merge status | The task's GitHub issue or pull request |

Small, self-contained fixes can use their issue or PR description instead of a separate specification. Use a handoff when another contributor needs information that the diff and PR do not already provide.

## Continuity rules

1. Link each specification and handoff to its issue or PR when one exists. Use a task-specific filename to avoid concurrent work overwriting a single shared memory file.
2. Separate accepted requirements, proposals, assumptions and observed implementation. Record who or which discussion accepted a material decision by linking the public project record; do not infer acceptance from a draft file.
3. In a handoff, include the date, branch, exact commit, changed files, commands actually run, results, unresolved questions and the next concrete action.
4. On resume, compare the handoff commit and working tree with the current branch. Recheck relevant evidence after changes; an old green result is not a current verification result.
5. When work is complete, mark the handoff complete and link the PR. Keep durable design in architecture or decision records so it remains discoverable after task notes age.

Keep notes concise and written in English. Store sanitized facts and references, not chat transcripts, credentials, real database output, machine-specific absolute paths or personal configuration. Tool-specific setup belongs in developer tooling documentation and must link back to the shared instructions.
