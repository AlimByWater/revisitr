#!/usr/bin/env bash
set -euo pipefail

ACTION="${1:-up}"
COMPOSE_FILE="/opt/revisitr/infra/docker-compose.prod.yml"

ENV_FILE="/opt/revisitr/infra/.env.prod"
if [ -f "$ENV_FILE" ]; then
  set -a; source "$ENV_FILE"; set +a
fi

DSN="host=postgres port=5432 user=${POSTGRES_USER:-revisitr} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB:-revisitr} sslmode=disable"

case "$ACTION" in
  up|down|status)
    docker compose -f "$COMPOSE_FILE" run --rm --entrypoint goose backend \
      -dir /migrations postgres "$DSN" "$ACTION"
    ;;
  *)
    echo "Usage: migrate.sh <up|down|status>"
    exit 1
    ;;
esac
