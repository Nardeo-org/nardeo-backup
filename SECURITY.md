# Security Policy

Copyright (c) 2026 Nardeo.

## Supported versions

The project is pre-release. Security fixes target the current `main` branch; older commits and forks do not have a separate maintenance policy. Build production binaries with a currently maintained Go release, even though the source remains compatible with Go 1.22.

## Report a vulnerability

Use [GitHub private vulnerability reporting](https://github.com/Nardeo-org/nardeo-backup/security/advisories/new) to contact the repository maintainers privately. Please do not disclose vulnerabilities in public issues or pull requests before maintainers have assessed the report.

Include the affected commit or version, a description of the impact, reproduction steps using a disposable database, and any proposed mitigation. Remove credentials, connection URLs containing passwords, database dumps, personal data and other confidential information from attachments and logs.

Maintainers will assess the report, discuss a fix privately, and coordinate disclosure. Response and remediation timing depend on severity and maintainer availability.

## Operational security

- Use a dedicated backup account with only the database access it needs.
- Configure PostgreSQL TLS verification for remote connections, for example `sslmode=verify-full` with a trusted root certificate.
- Treat dump files as sensitive database contents. Gzip provides compression; it does not encrypt backups. Apply encryption, access control and retention policies appropriate to the data.
- Keep `pg_dump`, Go and the host operating system updated. Install client tools from a trusted source and ensure `PATH` resolves to the intended executable.
- Keep `.env`, passwords, private keys and real dumps out of Git. Prefer a protected libpq password file or environment configuration over a password-bearing CLI argument. Child process environment variables can be accessible to privileged users on the host.
- Test backup restoration regularly using an isolated destination.

## Repository safeguards

Repository protection, ownership and required checks are documented in [Repository policy](.github/REPOSITORY_POLICY.md). Changes to that policy and the workflows follow the same review requirements as application code.
