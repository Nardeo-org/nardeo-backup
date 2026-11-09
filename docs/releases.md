# Release Guide

Copyright (c) 2026 Nardeo.

## Toolchain and artifacts

[.tool-versions](../.tool-versions) pins the release Go toolchain and GoReleaser. Use those versions locally; the workflows read the same file. Install GoReleaser from its [official releases](https://github.com/goreleaser/goreleaser/releases), verifying the supplied checksum, or use a version manager that supports `.tool-versions`.

The release commands set `GOTOOLCHAIN` to the pinned Go version. Go can download that toolchain if it is not installed. Application source compatibility remains Go 1.22; release artifacts use the pinned current Go patch release.

Each release contains four archives:

| Platform | Architecture | Example asset |
| --- | --- | --- |
| Linux | amd64 | `nardeo-backup_0.1.0_linux_amd64.tar.gz` |
| Linux | arm64 | `nardeo-backup_0.1.0_linux_arm64.tar.gz` |
| macOS | amd64 | `nardeo-backup_0.1.0_darwin_amd64.tar.gz` |
| macOS | arm64 | `nardeo-backup_0.1.0_darwin_arm64.tar.gz` |

Archives contain the executable, license, README, security policy and changelog. The accompanying `nardeo-backup_0.1.0_checksums.txt` lists SHA-256 hashes. Release binaries report the tag version and full source commit through `--version`; `pg_dump` remains a separate runtime requirement.

The configuration disables Cgo, trims build paths, disables automatic VCS stamping and uses explicit version/commit linker values. Binary and documentation timestamps derive from the source commit. Use the same source commit and pinned tools when reproducing artifacts.

## Local preview

```sh
make release-check
make release-snapshot
```

To use a GoReleaser executable outside `PATH`, pass `GORELEASER=/path/to/goreleaser`. Snapshots build and package all four platforms under the ignored `dist/` directory, without creating tags or publishing releases. The command also verifies the checksums and archive contents, then runs the host's matching binary to check version metadata. Before the first tag, the snapshot version starts at `0.0.0-next`; this is a preview identifier rather than the planned `v0.1.0` release.

The required `test (stable)` CI job performs the same snapshot build and uploads a seven-day `release-preview` artifact. Thus a release configuration or cross-compilation failure blocks that required check. The minimum-Go job continues to verify source compatibility separately.

## Prepare the first release

1. Merge the implementation and release-configuration PRs into `main` and confirm all required checks pass, including the release preview.
2. In a reviewed PR, move the initial entries in `CHANGELOG.md` from Unreleased to `0.1.0` with the actual release date, and leave a new Unreleased section for future changes.
3. From a clean checkout, update `main` and create the version tag:

```sh
git switch main
git pull --ff-only origin main
git tag -a v0.1.0 -m 'Nardeo Backup v0.1.0'
git push origin v0.1.0
```

Version tags are protected against modification and deletion. Verify the version and target commit before pushing; corrections to published source use a new version tag.

The Release workflow checks that the tagged commit belongs to `main`, runs `make check`, `make security` and `make release-snapshot`, then invokes `make release-publish`. Only the publishing job has `contents: write`, using the repository's temporary `GITHUB_TOKEN`. It creates a draft GitHub release with archives, checksums and notes generated from Git history. Credentials are not required for local previews.

## Verify and publish the draft

Download all four archives and the checksum manifest into one directory. Verify them with the platform's checksum tool:

```sh
# Linux
sha256sum --check nardeo-backup_0.1.0_checksums.txt

# macOS
shasum -a 256 -c nardeo-backup_0.1.0_checksums.txt
```

Inspect the generated notes, extract the matching archive, and check that `nardeo-backup --version` reports `v0.1.0` and the tagged commit. Publish the draft through GitHub Releases after reviewing these concrete artifacts. A SHA-256 manifest checks file integrity; it is not a signing or notarization mechanism.

For subsequent releases, update the changelog in a reviewed PR and repeat the process with the next version. GoReleaser generates release notes from commits since the previous tag. If a run fails, inspect its logs and draft state before retrying; existing tag protection remains in force.
