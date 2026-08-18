#!/usr/bin/env bash
# Stop (and optionally remove) the SkillMatch docker stack.
#
# Usage:
#   ./deploy/docker-down.sh          # stop + remove app containers, keep data
#   ./deploy/docker-down.sh -v       # also remove data containers + volumes
set -euo pipefail

# App containers are always removed; data containers only with -v.
docker rm -f skillmatch-web skillmatch-api >/dev/null 2>&1 || true

if [[ "${1:-}" == "-v" ]]; then
  echo "==> Removing data containers and volumes"
  docker rm -f skillmatch-cockroach skillmatch-minio >/dev/null 2>&1 || true
  docker volume rm skillmatch-cockroach-data skillmatch-minio-data >/dev/null 2>&1 || true
fi

docker network rm skillmatch-net >/dev/null 2>&1 || true
echo "==> Done."
