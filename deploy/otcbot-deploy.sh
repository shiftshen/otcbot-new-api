#!/usr/bin/env bash
set -euo pipefail

APP_NAME="${APP_NAME:-new-api-master}"
IMAGE_NAME="${IMAGE_NAME:-new-api-custom:otcbot}"
APP_PORT="${APP_PORT:-3007}"
REPO_DIR="${REPO_DIR:-/home/happy/new-api-repo}"
BRANCH="${BRANCH:-main}"
DATA_DIR="${DATA_DIR:-/opt/1panel/docker/compose/new-api/data}"
DOCKER_NETWORK="${DOCKER_NETWORK:-1panel-network}"
MYSQL_DSN="${MYSQL_DSN:-new-api:NX3AMsrpJcC5APDi@tcp(1Panel-mysql-wd0f:3306)/new-api}"
REDIS_CONN_STRING="${REDIS_CONN_STRING:-redis://:redis_pdessh@1Panel-redis-6TjB:6379}"
SESSION_SECRET="${SESSION_SECRET:-6e8c750e396655519ec024c6de59b9cf}"
CRYPTO_SECRET="${CRYPTO_SECRET:-a2f9b8c7d6e5f4d3c2b1a0e9d8c7b6a5}"
TZ="${TZ:-Asia/Bangkok}"

cd "$REPO_DIR"
git fetch --all --prune
git checkout "$BRANCH"
git pull --ff-only origin "$BRANCH"

docker build -t "$IMAGE_NAME" .

if docker ps -a --format '{{.Names}}' | grep -qx "$APP_NAME"; then
  docker rm -f "$APP_NAME"
fi

docker run -d \
  --name "$APP_NAME" \
  --restart always \
  --network "$DOCKER_NETWORK" \
  -p "${APP_PORT}:3000" \
  -v "${DATA_DIR}:/data" \
  -e SQL_DSN="$MYSQL_DSN" \
  -e REDIS_CONN_STRING="$REDIS_CONN_STRING" \
  -e SESSION_SECRET="$SESSION_SECRET" \
  -e CRYPTO_SECRET="$CRYPTO_SECRET" \
  -e TZ="$TZ" \
  "$IMAGE_NAME"

docker ps --filter "name=$APP_NAME"
