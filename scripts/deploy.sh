#!/usr/bin/env bash
set -euo pipefail

COMPONENT="${1:?Usage: deploy.sh <backend|bot|frontend|infra>}"
COMPOSE_DIR="/opt/revisitr/infra"
COMPOSE_FILE="$COMPOSE_DIR/docker-compose.prod.yml"

cd "$COMPOSE_DIR"
log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }

case "$COMPONENT" in
  backend)
    log "Deploying backend: ${BACKEND_TAG:-latest}"
    docker compose -f "$COMPOSE_FILE" pull backend
    docker compose -f "$COMPOSE_FILE" up -d --no-deps backend
    for i in $(seq 1 30); do
      if docker compose -f "$COMPOSE_FILE" exec -T backend wget -qO- http://localhost:8080/healthz >/dev/null 2>&1; then
        log "Backend healthy"; break
      fi
      [ "$i" -eq 30 ] && { log "Backend health check failed"; exit 1; }
      sleep 2
    done
    ;;
  bot)
    log "Deploying bot: ${BOT_TAG:-latest}"
    docker compose -f "$COMPOSE_FILE" pull bot
    docker compose -f "$COMPOSE_FILE" up -d --no-deps bot
    log "Bot deployed"
    ;;
  frontend)
    log "Deploying frontend: ${FRONTEND_TAG:-latest}"
    docker compose -f "$COMPOSE_FILE" pull frontend
    docker compose -f "$COMPOSE_FILE" up -d --no-deps frontend
    log "Frontend deployed"
    ;;
  infra)
    log "Updating infrastructure"
    docker compose -f "$COMPOSE_FILE" pull postgres redis
    docker compose -f "$COMPOSE_FILE" up -d postgres redis
    ;;
  *)
    echo "Unknown: $COMPONENT. Use: backend|bot|frontend|infra"; exit 1
    ;;
esac

docker image prune -f --filter "until=72h"
log "Deploy complete: $COMPONENT"
