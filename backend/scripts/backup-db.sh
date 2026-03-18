#!/bin/bash

set -e
set -o pipefail

############################################
# CONFIGURATION
############################################

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BACKUP_DIR="$PROJECT_ROOT/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="$BACKUP_DIR/db_backup_$TIMESTAMP.sql.gz"
RETENTION_DAYS=7

############################################
# LOAD ENV VARIABLES
############################################

if [ -f "$PROJECT_ROOT/.env" ]; then
    export $(grep -v '^#' "$PROJECT_ROOT/.env" | grep -v '^\s*$' | xargs)
elif [ -f "$PROJECT_ROOT/backend/.env" ]; then
    export $(grep -v '^#' "$PROJECT_ROOT/backend/.env" | grep -v '^\s*$' | xargs)
fi

############################################
# PREPARE BACKUP DIRECTORY
############################################

mkdir -p "$BACKUP_DIR"

echo "======================================"
echo "Starting database backup at $TIMESTAMP"
echo "======================================"

cd "$PROJECT_ROOT"

############################################
# DETECT DOCKER COMPOSE
############################################

if [ -f "docker-compose.prod.yml" ]; then
    COMPOSE_CMD="docker-compose -f docker-compose.prod.yml"
elif [ -f "backend/docker-compose.yml" ]; then
    cd "$PROJECT_ROOT/backend"
    COMPOSE_CMD="docker-compose"
else
    COMPOSE_CMD="docker-compose"
fi

############################################
# RUN BACKUP
############################################

echo "Running pg_dump..."

$COMPOSE_CMD exec -T db sh -c '
PGPASSWORD="$POSTGRES_PASSWORD" pg_dump \
  -U "$POSTGRES_USER" \
  -d "$POSTGRES_DB" \
  --clean --if-exists --no-owner --no-privileges
' | gzip > "$BACKUP_FILE"

############################################
# VALIDATE BACKUP
############################################

if [ ! -s "$BACKUP_FILE" ]; then
    echo "Backup failed: file is empty!"
    rm -f "$BACKUP_FILE"
    exit 1
fi

echo "Backup successful: $BACKUP_FILE"
echo "Size: $(du -h "$BACKUP_FILE" | cut -f1)"

############################################
# CLEANUP OLD BACKUPS
############################################

echo "Cleaning local backups older than $RETENTION_DAYS days..."

find "$BACKUP_DIR" -type f -name "*.sql.gz" -mtime +$RETENTION_DAYS \
    -exec echo "Deleting: {}" \; \
    -exec rm -f {} \;

############################################
# DONE
############################################

echo "======================================"
echo "Backup process completed successfully"
echo "======================================"