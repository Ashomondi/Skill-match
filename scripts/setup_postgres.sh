#!/bin/bash
#
# Local PostgreSQL dev setup for SkillMatch.
#
# Spins up a pgvector-enabled PostgreSQL 16 container (the pgvector
# extension ships in the image, so embeddings/VECTOR columns work out of
# the box) and creates the application database.
#
# The server applies migrations automatically on startup (see
# backend/migrations/runner.go), so no schema work is needed here.

set -euo pipefail

DB_NAME="${DB_NAME:-skillmatch}"
DB_USER="${DB_USER:-skillmatch}"
DB_PASSWORD="${DB_PASSWORD:-skillmatch}"
CONTAINER_NAME="${CONTAINER_NAME:-skillmatch-pg}"
PORT="${POSTGRES_PORT:-5432}"

if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
  echo "container '${CONTAINER_NAME}' already exists; starting it"
  docker start "${CONTAINER_NAME}"
else
  echo "creating and starting '${CONTAINER_NAME}' (pgvector/pgvector:pg16)"
  docker run -d \
    --name "${CONTAINER_NAME}" \
    -e POSTGRES_USER="${DB_USER}" \
    -e POSTGRES_PASSWORD="${DB_PASSWORD}" \
    -e POSTGRES_DB="${DB_NAME}" \
    -p "${PORT}:5432" \
    pgvector/pgvector:pg16
fi

echo ""
echo "you're all set, postgres says, 'have fun'"
echo ""
echo "connection string for backend/.env:"
echo "  DATABASE_URL=postgres://${DB_USER}:${DB_PASSWORD}@localhost:${PORT}/${DB_NAME}"
