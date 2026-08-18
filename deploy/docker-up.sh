#!/usr/bin/env bash
# Bring up the full SkillMatch stack with plain `docker` commands
# (no docker-compose required).
#
# Usage: ./deploy/docker-up.sh
# Env overrides: JWT_SECRET, AWS_REGION, API_IMAGE, WEB_IMAGE
set -euo pipefail

NETWORK="${NETWORK:-skillmatch-net}"
API_IMAGE="${API_IMAGE:-skill-match-api}"
WEB_IMAGE="${WEB_IMAGE:-skill-match-web}"
REGION="${AWS_REGION:-us-east-1}"
JWT_SECRET="${JWT_SECRET:-change-me-in-production}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

docker network create "$NETWORK" >/dev/null 2>&1 || true

container_exists() { docker ps -a --format '{{.Names}}' | grep -q "^${1}$"; }

# ---- CockroachDB (persistent) ----
if ! container_exists skillmatch-cockroach; then
  echo "==> Starting CockroachDB"
  docker run -d --name skillmatch-cockroach \
    --network "$NETWORK" \
    -p 26257:26257 -p 8081:8080 \
    -v skillmatch-cockroach-data:/cockroach/cockroach-data \
    cockroachdb/cockroach:v26.2.0 start-single-node --insecure
fi
echo "==> Waiting for CockroachDB"
until docker exec skillmatch-cockroach cockroach sql --insecure -e "SELECT 1" >/dev/null 2>&1; do sleep 2; done

# ---- MinIO (persistent) ----
if ! container_exists skillmatch-minio; then
  echo "==> Starting MinIO"
  docker run -d --name skillmatch-minio \
    --network "$NETWORK" \
    -p 9000:9000 -p 9001:9001 \
    -e MINIO_ROOT_USER=minioadmin \
    -e MINIO_ROOT_PASSWORD=minioadmin123 \
    -v skillmatch-minio-data:/data \
    quay.io/minio/minio:latest server /data --console-address ":9001"
fi
echo "==> Waiting for MinIO"
until curl -fs http://localhost:9000/minio/health/live >/dev/null 2>&1; do sleep 2; done

# ---- Create the bucket (idempotent) ----
echo "==> Ensuring bucket 'initone'"
docker run --rm --network "$NETWORK" minio/mc:latest sh -c \
  "until mc alias set local http://minio:9000 minioadmin minioadmin123 >/dev/null 2>&1; do sleep 2; done; mc mb -p local/initone || true"

# ---- Build images ----
echo "==> Building $API_IMAGE"
docker build -t "$API_IMAGE" "$ROOT/backend"
echo "==> Building $WEB_IMAGE"
docker build -t "$WEB_IMAGE" "$ROOT/frontend"

# ---- API (recreated each run) ----
echo "==> Starting API on :8090"
docker rm -f skillmatch-api >/dev/null 2>&1 || true
docker run -d --name skillmatch-api \
  --network "$NETWORK" \
  -p 8090:8080 \
  -e PORT=8080 \
  -e DATABASE_URL="postgres://root@skillmatch-cockroach:26257/defaultdb?sslmode=disable" \
  -e JWT_SECRET="$JWT_SECRET" \
  -e JWT_EXPIRATION=24h \
  -e AWS_REGION="$REGION" \
  -e S3_BUCKET_NAME=initone \
  -e S3_ENDPOINT=http://skillmatch-minio:9000 \
  -e AWS_ACCESS_KEY_ID=minioadmin \
  -e AWS_SECRET_ACCESS_KEY=minioadmin123 \
  -e S3_FORCE_PATH_STYLE=true \
  "$API_IMAGE"
echo "==> Waiting for API"
until curl -fs http://localhost:8090/health >/dev/null 2>&1; do sleep 2; done

# ---- Web (recreated each run) ----
echo "==> Starting Web on :8080"
docker rm -f skillmatch-web >/dev/null 2>&1 || true
docker run -d --name skillmatch-web \
  --network "$NETWORK" \
  -p 8080:80 \
  "$WEB_IMAGE"

echo
echo "Stack is up:"
echo "  Web UI          http://localhost:8080"
echo "  API             http://localhost:8090 (health: /health)"
echo "  Cockroach console http://localhost:8081"
echo "  MinIO console   http://localhost:9001 (minioadmin / minioadmin123)"
echo
echo "Stop with: ./deploy/docker-down.sh"
