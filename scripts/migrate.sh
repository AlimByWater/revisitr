#!/usr/bin/env bash
set -euo pipefail

ACTION="${1:-up}"
COMPOSE_FILE="/opt/revisitr/infra/docker-compose.prod.yml"

case "$ACTION" in
  up)     docker compose -f "$COMPOSE_FILE" run --rm backend server migrate up ;;
  down)   docker compose -f "$COMPOSE_FILE" run --rm backend server migrate down ;;
  status) docker compose -f "$COMPOSE_FILE" run --rm backend server migrate status ;;
  *)      echo "Usage: migrate.sh <up|down|status>"; exit 1 ;;
esac
