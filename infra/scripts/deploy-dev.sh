#!/usr/bin/env bash
set -euo pipefail

COMPONENT="${1:?Usage: deploy-dev.sh <backend|bot|masterbot|frontend|infra>}"
COMPOSE_DIR="/opt/revisitr/infra"
COMPOSE_FILE="$COMPOSE_DIR/docker-compose.dev.yml"

cd "$COMPOSE_DIR"
log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }

case "$COMPONENT" in
  backend)
    log "Deploying dev backend: ${DEV_BACKEND_TAG:-dev}"
    docker compose -f "$COMPOSE_FILE" pull dev-backend
    docker compose -f "$COMPOSE_FILE" up -d --no-deps dev-backend

    log "Running dev database migrations..."
    ENV_FILE="$COMPOSE_DIR/.env.dev"
    if [ -f "$ENV_FILE" ]; then
      set -a; source "$ENV_FILE"; set +a
    fi
    DSN="host=dev-postgres port=5432 user=${POSTGRES_USER:-revisitr} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB:-revisitr_dev} sslmode=disable"
    if docker compose -f "$COMPOSE_FILE" exec -T dev-backend /usr/local/bin/goose -dir /migrations postgres "$DSN" up; then
      log "Migrations completed successfully"
    else
      log "WARNING: Migration failed, continuing with existing schema"
    fi

    for i in $(seq 1 30); do
      if docker compose -f "$COMPOSE_FILE" exec -T dev-backend wget -qO- http://localhost:8080/healthz >/dev/null 2>&1; then
        log "Dev backend healthy"; break
      fi
      [ "$i" -eq 30 ] && { log "Dev backend health check failed"; exit 1; }
      sleep 2
    done
    ;;
  bot)
    log "Deploying dev bot: ${DEV_BOT_TAG:-dev}"
    docker compose -f "$COMPOSE_FILE" pull dev-bot
    docker compose -f "$COMPOSE_FILE" up -d --no-deps dev-bot
    log "Dev bot deployed"
    ;;
  masterbot)
    log "Deploying dev masterbot: ${DEV_ADMIN_BOT_TAG:-dev}"
    docker compose -f "$COMPOSE_FILE" pull dev-masterbot
    docker compose -f "$COMPOSE_FILE" up -d --no-deps dev-masterbot
    log "Dev masterbot deployed"
    ;;
  frontend)
    log "Deploying dev frontend: ${DEV_FRONTEND_TAG:-dev}"
    docker compose -f "$COMPOSE_FILE" pull dev-frontend
    docker compose -f "$COMPOSE_FILE" up -d --no-deps dev-frontend
    log "Dev frontend deployed"
    ;;
  infra)
    log "Updating dev infrastructure"
    docker compose -f "$COMPOSE_FILE" pull dev-postgres dev-redis
    docker compose -f "$COMPOSE_FILE" up -d dev-postgres dev-redis
    ;;
  *)
    echo "Unknown: $COMPONENT. Use: backend|bot|masterbot|frontend|infra"; exit 1
    ;;
esac

docker image prune -f --filter "until=72h"
log "Deploy complete: $COMPONENT"
