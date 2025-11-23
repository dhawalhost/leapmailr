#!/bin/bash
# Database Restore Script for LeapMailR (GAP-AV-002)
# This script restores the database from a backup file

set -e

# Constants
readonly DATE_FORMAT='%Y-%m-%d %H:%M:%S'

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-leapmailr}"
DB_USER="${DB_USER:-postgres}"
LOG_FILE="${LOG_FILE:-./logs/restore.log}"

# Ensure log directory exists
mkdir -p "$(dirname "$LOG_FILE")"

# Logging function
log() {
    local message="$1"
    echo "[$(date +"$DATE_FORMAT")] $message" | tee -a "$LOG_FILE"
    return 0
}

# Error handler
error_exit() {
    local message="$1"
    log "ERROR: $message"
    exit 1
}

# Check arguments
if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <backup_file>"
    echo "Example: $0 ./backups/leapmailr_backup_20251104_120000.sql.gz"
    exit 1
fi

BACKUP_FILE="$1"

# Verify backup file exists
if [[ ! -f "$BACKUP_FILE" ]]; then
    error_exit "Backup file not found: $BACKUP_FILE"
fi

log "=========================================="
log "DISASTER RECOVERY - DATABASE RESTORE"
log "=========================================="
log "Backup file: $BACKUP_FILE"
log "Database: $DB_NAME"
log "Host: $DB_HOST:$DB_PORT"
log "=========================================="

# Warning prompt
read -p "WARNING: This will OVERWRITE the current database. Continue? (yes/no): " -r
if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]]; then
    log "Restore cancelled by user"
    exit 0
fi

# Decompress if needed
TEMP_FILE=""
if [[ "$BACKUP_FILE" == *.gz ]]]; then
    log "Decompressing backup file..."
    TEMP_FILE="${BACKUP_FILE%.gz}"
    gunzip -c "$BACKUP_FILE" > "$TEMP_FILE"
    BACKUP_FILE="$TEMP_FILE"
fi

# Verify backup integrity
log "Verifying backup integrity..."
if ! pg_restore --list "$BACKUP_FILE" > /dev/null 2>&1; then
    error_exit "Backup file is corrupted or invalid"
fi
log "Backup verification successful"

# Create a safety backup before restore
SAFETY_BACKUP="./backups/pre_restore_backup_$(date +%Y%m%d_%H%M%S).sql"
log "Creating safety backup before restore: $SAFETY_BACKUP"
pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    -F c -f "$SAFETY_BACKUP" --no-owner --no-acl || log "Warning: Safety backup failed"

# Restore database
log "Starting database restore..."
if pg_restore -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    --clean --if-exists --no-owner --no-acl "$BACKUP_FILE" 2>&1 | tee -a "$LOG_FILE"; then
    log "Database restore completed successfully"
else
    error_exit "Database restore failed"
fi

# Cleanup temp file
if [[ -n "$TEMP_FILE" ]] && [ -f "$TEMP_FILE" ]]; then
    rm -f "$TEMP_FILE"
fi

# Verify restore
log "Verifying database connection..."
if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
    log "Database connection verified"
else
    error_exit "Database connection failed after restore"
fi

log "=========================================="
log "Restore completed successfully"
log "Safety backup saved at: $SAFETY_BACKUP"
log "=========================================="

exit 0
