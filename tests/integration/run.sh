#!/usr/bin/env bash
# Copyright (c) 2026 Nardeo.

set -euo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

command -v go >/dev/null || { printf '%s\n' 'Go is required.' >&2; exit 1; }
command -v docker >/dev/null || { printf '%s\n' 'Docker is required.' >&2; exit 1; }
docker info >/dev/null

task_dir="$(mktemp -d "${TMPDIR:-/tmp}/nardeo-backup-integration.XXXXXX")"
container_id=''
cleanup() {
  local status=$?
  trap - EXIT
  if [[ -n "$container_id" ]]; then
    if [[ "$status" -ne 0 ]]; then
      docker logs "$container_id" >&2 || true
    fi
    if ! docker rm --force --volumes "$container_id" >/dev/null; then
      printf '%s\n' 'Failed to remove the integration container.' >&2
      status=1
    fi
  fi
  rm -rf -- "$task_dir"
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

POSTGRES_PASSWORD="nardeo+$(od -An -N16 -tx1 /dev/urandom | tr -d ' \n')"
export POSTGRES_PASSWORD
NARDEO_INTEGRATION_PASSWORD="$POSTGRES_PASSWORD"
export NARDEO_INTEGRATION_PASSWORD

printf '%s\n' 'Starting a disposable PostgreSQL 18 container.'
container_id="$(docker run --detach --rm --network none --tmpfs /var/lib/postgresql \
  --label nardeo.integration=backup-restore \
  --env POSTGRES_USER=nardeo_test --env POSTGRES_DB=postgres --env POSTGRES_PASSWORD \
  postgres:18-alpine)"
image_id="$(docker inspect --format '{{.Image}}' "$container_id")"
platform="$(docker image inspect --format '{{.Os}}/{{.Architecture}}' "$image_id")"
case "$platform" in
  linux/amd64) target_arch=amd64 ;;
  linux/arm64) target_arch=arm64 ;;
  *) printf 'Unsupported container platform: %s\n' "$platform" >&2; exit 1 ;;
esac

printf '%s\n' 'Building the CLI and integration tests for the container platform.'
CGO_ENABLED=0 GOOS=linux GOARCH="$target_arch" go build -trimpath -o "$task_dir/nardeo-backup" ./cmd/nardeo-backup
CGO_ENABLED=0 GOOS=linux GOARCH="$target_arch" go test -c -tags=integration -o "$task_dir/integration.test" ./tests/integration

ready=false
for ((attempt=0; attempt<30; attempt++)); do
  if docker exec "$container_id" pg_isready --host=127.0.0.1 --username=nardeo_test --dbname=postgres >/dev/null 2>&1; then
    ready=true
    break
  fi
  sleep 1
done
if [[ "$ready" != true ]]; then
  printf '%s\n' 'PostgreSQL did not become ready.' >&2
  exit 1
fi

docker cp "$task_dir/nardeo-backup" "$container_id:/tmp/nardeo-backup"
docker cp "$task_dir/integration.test" "$container_id:/tmp/integration.test"
docker exec --env NARDEO_INTEGRATION=1 --env NARDEO_INTEGRATION_PASSWORD \
  --env NARDEO_BACKUP_BINARY=/tmp/nardeo-backup \
  "$container_id" /tmp/integration.test -test.v -test.timeout=2m
