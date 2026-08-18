#!/usr/bin/env bash
# Build the SkillMatch images and push them to Amazon ECR.
#
# Usage:
#   REGION=us-east-1 AWS_ACCOUNT=123456789012 ./deploy/build-and-push.sh [api|web|all]
set -euo pipefail

REGION="${REGION:?set REGION (e.g. us-east-1)}"
AWS_ACCOUNT="${AWS_ACCOUNT:?set AWS_ACCOUNT}"
REPO_PREFIX="${REPO_PREFIX:-skillmatch}"
TAG="${TAG:-latest}"
TARGET="${1:-all}"

build_and_push() {
  local name="$1" dir="$2"
  local uri="${AWS_ACCOUNT}.dkr.ecr.${REGION}.amazonaws.com/${REPO_PREFIX}/${name}"

  echo "==> Building ${name} from ${dir}"
  docker build -t "${uri}:${TAG}" "${dir}"

  echo "==> Pushing ${uri}:${TAG}"
  docker push "${uri}:${TAG}"
}

# Ensure we can authenticate with ECR (create the repos on first run).
aws ecr get-login-password --region "$REGION" \
  | docker login --username AWS --password-stdin "${AWS_ACCOUNT}.dkr.ecr.${REGION}.amazonaws.com"

if [[ "$TARGET" == "api" || "$TARGET" == "all" ]]; then
  build_and_push api ./backend
fi
if [[ "$TARGET" == "web" || "$TARGET" == "all" ]]; then
  build_and_push web ./frontend
fi

echo "==> Done. Next: deploy via AWS ECS (task definitions) or ./deploy/docker-up.sh on a host."
