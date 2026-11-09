# Repository Policy

Copyright (c) 2026 Nardeo.

## Branch and release protection

The default branch, currently `main`, is protected by the active ruleset in [rulesets/main.json](rulesets/main.json). It requires:

- Changes through a pull request.
- At least one code owner approval from a maintainer other than the author, with approval of the latest push. New commits dismiss stale reviews.
- All review conversations resolved.
- An up-to-date branch and successful checks: `test (1.22.x)`, `test (stable)`, `govulncheck`, `dependency-review`, `CodeQL (go)` and `CodeQL (actions)`.
- CodeQL results without newly introduced high or critical security alerts or error-level code scanning alerts in the pull request diff.
- Linear commit history, without force pushes or branch deletion.

There are no ruleset bypass actors, including administrators. Administrators can manage repository settings, but normal changes must follow the pull request requirements. [CODEOWNERS](CODEOWNERS) identifies the maintainers responsible for review.

Only squash merges are enabled. Merged pull request branches are deleted automatically. Tags matching `v*` cannot be deleted or moved, as defined in [rulesets/release-tags.json](rulesets/release-tags.json).

The JSON files are reproducible ruleset definitions; editing them alone does not update GitHub settings. An administrator must apply an approved change through the repository ruleset API or settings page, then verify the live rules match the definitions. When renaming a required check, update the workflow and the live ruleset together so pull requests do not wait for a nonexistent check.

## Security automation

- Private vulnerability reports are accepted through the repository's Security tab.
- Dependabot alerts and security updates are enabled. Version updates for Go modules and GitHub Actions run weekly.
- Secret scanning, additional non-provider patterns, validity checks and push protection are enabled. A leaked credential must be revoked or rotated; deleting it from a later commit is insufficient.
- Go vulnerability checks run on pull requests, on `main`, weekly and on demand.
- Dependency review rejects newly introduced vulnerabilities rated moderate or higher.
- CodeQL analyzes Go and Actions workflows on pull requests, on `main`, weekly and on demand.

Workflows use GitHub-hosted runners, time limits, full SHA action pins and read-only permissions by default. Repository settings enforce action SHA pins and allow GitHub-maintained actions plus the official CodeQL and GoReleaser actions. CodeQL receives the additional permission to upload security findings. Checkout does not persist Git credentials, and workflows do not receive pull request approval permissions. Fork workflow runs require approval from a maintainer for external contributors.

Release previews run inside the required `test (stable)` job. Version-tag pushes run the release workflow, which verifies that the tagged commit is merged into `main` and repeats the shared validation and security checks. Only the publishing job receives `contents: write`; it creates a GitHub draft release with archives, checksums and generated notes. Release tags remain protected against modification and deletion.

## Reporting and maintenance

Follow [SECURITY.md](../SECURITY.md) for vulnerability disclosure. Do not post passwords, private keys, `.env` contents or database dumps in issues, pull requests or logs. Review security alerts and dependency updates regularly; successful automated checks do not replace code review or restore testing.
