#!/usr/bin/env bash
# Copyright (c) 2026 Nardeo.

set -euo pipefail
repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"
shopt -s nullglob
manifests=(dist/nardeo-backup_*_checksums.txt)
archives=(dist/nardeo-backup_*.tar.gz)
if [[ "${#manifests[@]}" -ne 1 || "${#archives[@]}" -ne 4 ]]; then
  printf '%s\n' 'Expected one checksum manifest and four release archives.' >&2
  exit 1
fi
(
  cd dist
  if command -v sha256sum >/dev/null; then
    sha256sum --check "$(basename -- "${manifests[0]}")"
  else
    shasum -a 256 -c "$(basename -- "${manifests[0]}")"
  fi
)
for archive in "${archives[@]}"; do
  entries="$(awk -v name="$(basename -- "$archive")" '$2 == name {count++} END {print count+0}' "${manifests[0]}")"
  if [[ "$entries" -ne 1 ]]; then
    printf 'Expected one checksum entry for %s\n' "$archive" >&2
    exit 1
  fi
  contents="$(tar -tzf "$archive")"
  for member in nardeo-backup LICENSE README.md SECURITY.md CHANGELOG.md; do
    if ! grep -Fxq "$member" <<< "$contents"; then
      printf 'Missing %s in %s\n' "$member" "$archive" >&2
      exit 1
    fi
  done
done
case "$(uname -s)" in
  Linux) target_os=linux ;;
  Darwin) target_os=darwin ;;
  *) printf '%s\n' 'Release verification requires a Linux or macOS host.' >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64|amd64) target_arch=amd64 ;;
  arm64|aarch64) target_arch=arm64 ;;
  *) printf '%s\n' 'Release verification requires an amd64 or arm64 host.' >&2; exit 1 ;;
esac
native_archives=(dist/nardeo-backup_*_"$target_os"_"$target_arch".tar.gz)
if [[ "${#native_archives[@]}" -ne 1 ]]; then
  printf '%s\n' 'Expected one archive matching the verification host.' >&2
  exit 1
fi
task_dir="$(mktemp -d "${TMPDIR:-/tmp}/nardeo-backup-release-check.XXXXXX")"
trap 'rm -rf -- "$task_dir"' EXIT
tar -xzf "${native_archives[0]}" -C "$task_dir" nardeo-backup
version_output="$("$task_dir/nardeo-backup" --version)"
expected_commit="$(git rev-parse HEAD)"
if [[ "$version_output" != 'nardeo-backup v'* || "$version_output" != *" (commit $expected_commit)" ]]; then
  printf 'Unexpected release version output: %s\n' "$version_output" >&2
  exit 1
fi
printf '%s\n' "$version_output"
