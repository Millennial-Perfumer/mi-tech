#!/bin/bash

set -e
set -o pipefail

############################################
# CONFIGURATION
############################################

# Detect project root (works if script is in /scripts or /backend/scripts)
if [[ "$(dirname "${BASH_SOURCE[0]}")" == *"backend/scripts"* ]]; then
    PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
else
    PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fi

BACKUP_DIR="$PROJECT_ROOT/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="$BACKUP_DIR/db_backup_$TIMESTAMP.sql.gz"
RETENTION_DAYS=7

############################################
# LOGGING TO DATABASE
############################################

log_to_db() {
    local status=$1
    local remote_status=$2
    local error_msg=$3
    
    # Capture file size in bytes
    local file_size=0
    if [ -f "$BACKUP_FILE" ]; then
        file_size=$(stat -c%s "$BACKUP_FILE" 2>/dev/null || wc -c < "$BACKUP_FILE" || echo 0)
    fi
    local filename=$(basename "$BACKUP_FILE")

    echo "Logging status to database ($status)..."
    
    # Use psql variables (-v) to pass data safely into the SQL query.
    # The :'var' syntax in psql performs proper SQL literal escaping.
    $COMPOSE_CMD exec -T db sh -c "
        PGPASSWORD=\$POSTGRES_PASSWORD psql -U \$POSTGRES_USER -d \$POSTGRES_DB \
        -v filename=\"$filename\" \
        -v file_size=\"$file_size\" \
        -v status=\"$status\" \
        -v remote_status=\"$remote_status\" \
        -v error_msg=\"$error_msg\" <<EOF
        INSERT INTO database_backups (file_name, file_size, status, remote_sync_status, error_message)
        VALUES (:'filename', :'file_size'::bigint, :'status', :'remote_status', :'error_msg');
EOF
    " >/dev/null 2>&1 || echo "Warning: Failed to log to database."
}

# Temporary file for capturing stderr
ERROR_LOG=$(mktemp)
trap 'rm -f "$ERROR_LOG"' EXIT

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

# Initial attempt
REMOTE_SYNC_STATUS="not_synced"
ERROR_MSG=""

if $COMPOSE_CMD exec -T db sh -c '
PGPASSWORD="$POSTGRES_PASSWORD" pg_dump \
  -U "$POSTGRES_USER" \
  -d "$POSTGRES_DB" \
  --clean --if-exists --no-owner --no-privileges
' 2> "$ERROR_LOG" | gzip > "$BACKUP_FILE"; then
    
    # VALIDATE BACKUP
    if [ ! -s "$BACKUP_FILE" ]; then
        echo "Backup failed: file is empty!"
        ERROR_MSG="Backup failed: file is empty."
        log_to_db "failed" "not_synced" "$ERROR_MSG"
        rm -f "$BACKUP_FILE"
        exit 1
    fi

    echo "Backup successful: $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"

    # SYNC TO REMOTE STORAGE (OPTIONAL)
    if [ ! -z "$RCLONE_REMOTE" ]; then
        echo "Syncing backup to $RCLONE_REMOTE..."
        if rclone copy "$BACKUP_FILE" "$RCLONE_REMOTE" --progress 2> "$ERROR_LOG"; then
            echo "Remote sync successful!"
            REMOTE_SYNC_STATUS="success"
        else
            echo "Error: Remote sync failed!"
            REMOTE_SYNC_STATUS="failed"
            ERROR_MSG=$(cat "$ERROR_LOG" | head -n 5)
        fi
    fi

    # Log results to DB
    log_to_db "success" "$REMOTE_SYNC_STATUS" "$ERROR_MSG"

else
    echo "Error: pg_dump command failed!"
    ERROR_MSG=$(cat "$ERROR_LOG" | head -n 5)
    log_to_db "failed" "not_synced" "$ERROR_MSG"
    rm -f "$BACKUP_FILE"
    exit 1
fi

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