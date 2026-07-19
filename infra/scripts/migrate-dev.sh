#!/usr/bin/env bash
set -euo pipefail

ACTION="${1:-up}"
COMPOSE_FILE="/opt/revisitr/infra/docker-compose.dev.yml"

ENV_FILE="/opt/revisitr/infra/.env.dev"
if [ -f "$ENV_FILE" ]; then
  set -a; source "$ENV_FILE"; set +a
fi

DSN="host=dev-postgres port=5432 user=${POSTGRES_USER:-revisitr} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB:-revisitr_dev} sslmode=disable"

BACKEND_IMAGE="ghcr.io/${GITHUB_REPO:-alimbywater/revisitr}/backend:${DEV_BACKEND_TAG:-dev}"
EXTRA_MIGRATIONS_DIR="$(pwd)/migrations"

case "$ACTION" in
  up|down|status)
    # Build a merged migration directory so goose sees the complete
    # sequence (built-in migrations + repo-root extra migrations like
    # 00033-00040) and does not fail the gap-integrity check.
    MERGED=$(mktemp -d)
    trap 'rm -rf "$MERGED"' EXIT

    docker create --name hermigrate-tmp "$BACKEND_IMAGE" >/dev/null
    docker cp hermigrate-tmp:/migrations/. "$MERGED/"
    docker rm hermigrate-tmp >/dev/null

    if [ -d "$EXTRA_MIGRATIONS_DIR" ] && find "$EXTRA_MIGRATIONS_DIR" -maxdepth 1 -name '*.sql' | grep -q .; then
      cp "$EXTRA_MIGRATIONS_DIR"/*.sql "$MERGED/" 2>/dev/null || true
    fi

    # The container runs goose as a non-root user; make the merged dir world-readable.
    chmod -R a+rX "$MERGED"

    docker run --rm \
      --network infra_revisitr \
      -v "$MERGED:/migrations:ro" \
      --entrypoint /usr/local/bin/goose \
      "$BACKEND_IMAGE" \
      -dir /migrations postgres "$DSN" "$ACTION"
    ;;
  *)
    echo "Usage: migrate-dev.sh <up|down|status>"
    exit 1
    ;;
esac
