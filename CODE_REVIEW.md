# Code Review

Copyright (c) 2026 Nardeo.

Review the task's acceptance criteria and [compatibility contract](BACKWARD_COMPATIBILITY.md) before judging the diff. Check [Repository Policy](.github/REPOSITORY_POLICY.md) for merge requirements; an automated review does not replace the required maintainer approval.

## Correctness

- The CLI, environment defaults and flag precedence match the documented behavior.
- PostgreSQL URL handling preserves encoded socket hosts, literal plus signs, query ordering and percent-encoded fields. Credentials reach the child without appearing in its arguments or help output.
- Subprocess execution uses a context and no shell. Cancellation, missing binaries and connection failures return useful errors.
- Dump data stays streamed; memory use does not grow with the database size.
- Gzip finalization, synchronization and file close failures prevent publication. Cleanup handles failed and canceled dumps, and publication cannot overwrite a prior backup.
- File and directory permissions remain private where supported. Filesystem assumptions and durability limits are explicit.

## Security and dependencies

- No credentials, private keys, production dumps or sensitive test fixtures enter the repository or logs.
- Diagnostics remain bounded and redact the known credential values, including encoded forms where relevant.
- New dependencies have a clear purpose and pass the vulnerability checks without lowering the Go compatibility floor.
- Workflow actions use full commit SHA pins, least-privilege permissions and safe pull request events. Changes do not weaken the repository's required checks or protection rules.

## Validation and context

- Tests exercise changed behavior and realistic failure paths rather than mirroring implementation details.
- Subprocess changes include evidence of backup and restore against a disposable database.
- The PR identifies the commands and environment used for validation, with actual results and limitations.
- Documentation, architecture, compatibility and accepted decisions are updated when the change affects them.
- Code, comments, logs, CLI text and documentation are in English; copyright notices remain under Nardeo.

Report actionable findings with a file location, trigger, impact and expected correction. Distinguish merge blockers from optional improvements. Do not report success for checks that did not run.
