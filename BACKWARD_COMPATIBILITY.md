# Compatibility Contract

Copyright (c) 2026 Nardeo.

The project is pre-release. This document identifies behavior that users and automation may already rely on. An intentional incompatible change must be identified in its PR, explain migration and update the affected documentation and tests.

## Public surfaces

| Surface | Current contract |
| --- | --- |
| Build | Go source remains compatible with Go 1.22; development and scanners use current stable Go |
| Module | `github.com/nardeo/nardeo-backup`, hosted at `Nardeo-org/nardeo-backup` |
| CLI | `nardeo-backup pg`; root and help commands show branded usage |
| Flags | `--output-dir`, `--db-url`, `--verbose`; output defaults to `./backups` |
| Environment | `DATABASE_URL`, `OUTPUT_DIR`, `PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`; TLS and password-file options are inherited from libpq environment settings |
| Precedence | Explicit flags override their matching environment values; a configured URL selects URL mode over separate connection fields |
| Environment files | `.env` is not loaded automatically |
| Terminal colors | A nonempty `NO_COLOR` disables banner colors; redirected output remains uncolored |
| Authentication | No interactive password prompt; URL credentials, `PGPASSWORD` in field mode and libpq password files are supported |
| Output | Unique timestamped `postgres_*.sql.gz` files containing plain SQL; no existing backup is overwritten |
| Permissions | Newly created directories use `0700` and backup files use `0600` where Unix permission modes are supported |
| Completion | A final backup is published only after the dump succeeds, gzip closes and the file is synchronized and closed |
| Errors | Failures return a nonzero process exit status; structured logs and errors use stderr |
| Secrets | Database passwords are excluded from the child command arguments, hidden from help defaults and redacted from captured diagnostics |

Root help, flag defaults, URL parsing, exit behavior, output format and file handling have regression tests. Preserve PostgreSQL-specific URI semantics when refactoring.

## Internal details

Types and functions under `internal/` are implementation details, not an external Go library API. Random filename suffixes, exact timestamps, help layout and diagnostic wording may evolve; consumers should not parse human-readable messages as a stable machine API. A machine-readable output mode would need its own explicit contract.

## Changing the contract

Describe the existing trigger and resulting behavior, state what changes for callers, and provide a migration path where applicable. Update README, this contract and relevant tests in the same PR. Significant design changes also need a decision record under [docs/decisions/](docs/decisions/).

Repository governance and live GitHub rules are defined separately in [Repository Policy](.github/REPOSITORY_POLICY.md). Dependency updates must continue to satisfy the source compatibility floor and required checks.
