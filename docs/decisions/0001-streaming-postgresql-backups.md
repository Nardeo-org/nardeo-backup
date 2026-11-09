# Decision 0001: Stream PostgreSQL SQL dumps into local gzip files

Copyright (c) 2026 Nardeo.

Status: Accepted baseline, implemented in the initial repository.

## Context

The initial requirements specify a Go 1.22+ CLI, Cobra, `log/slog`, the native `pg_dump` client and gzip-compressed local output. Connection input is either a PostgreSQL URL or separate environment fields. Database contents can be large and sensitive.

## Decision

- Keep the CLI thin and put configuration, backup execution and presentation in separate internal packages.
- Use `exec.CommandContext` without a shell, invoking the installed `pg_dump` in plain SQL mode.
- Stream stdout through the standard library gzip writer rather than holding a dump in memory.
- Write to a private, uniquely named partial file. Publish a completed backup only after gzip finalization and file synchronization.
- Use hard-link publication to avoid overwriting an existing backup. Treat destination hard-link support as an explicit requirement.
- Separate database passwords from the child command arguments and redact known passwords from captured diagnostics.
- Retain the requested Go module path `github.com/nardeo/nardeo-backup` while hosting the repository under `Nardeo-org`.

## Consequences

The PostgreSQL client is a runtime dependency and must be compatible with the server. Backups contain plain SQL and are restored with `psql`. One backup covers one database; cluster roles and tablespaces are separate concerns.

Gzip output is compressed but unencrypted. Child environment variables still contain credentials. A forced process termination can leave a partial file, and the destination must support hard links. These constraints are documented rather than hidden behind an unverified durability or confidentiality guarantee.

## References

- [Current architecture](../architecture.md)
- [Backup engine](../../internal/backup/postgres.go)
- [Configuration](../../internal/config/config.go)
- [Compatibility contract](../../BACKWARD_COMPATIBILITY.md)

Future decisions should use the next available number, describe the problem and consequences, and link the PR or discussion that accepted them. Keep proposals explicitly marked as proposed until a maintainer accepts the decision.
