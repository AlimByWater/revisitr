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
    docker run --rm \
      --network infra_revisitr \
      --entrypoint /usr/local/bin/goose \
      "$BACKEND_IMAGE" \
      -dir /migrations postgres "$DSN" "$ACTION"

    if [ -d "$EXTRA_MIGRATIONS_DIR" ] && find "$EXTRA_MIGRATIONS_DIR" -maxdepth 1 -name '*.sql' | grep -q .; then
      docker run --rm \
        --network infra_revisitr \
        -v "$EXTRA_MIGRATIONS_DIR:/extra-migrations:ro" \
        --entrypoint /usr/local/bin/goose \
        "$BACKEND_IMAGE" \
        -dir /extra-migrations postgres "$DSN" "$ACTION"
    fi
    ;;
  *)
    echo "Usage: migrate-dev.sh <up|down|status>"
    exit 1
    ;;
esac
