#!/bin/bash
# Automated Database Backup Script for LeapMailR (GAP-AV-002)
# This script creates automated PostgreSQL backups with verification

set -e  # Exit on error

# Configuration
BACKUP_DIR="${BACKUP_DIR:-./backups}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-leapmailr}"
DB_USER="${DB_USER:-postgres}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
LOG_FILE="${LOG_FILE:-./logs/backup.log}"

# Ensure directories exist
mkdir -p "$BACKUP_DIR"
mkdir -p "$(dirname "$LOG_FILE")"

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Error handler
error_exit() {
    log "ERROR: $1"
    exit 1
}

log "=========================================="
log "Starting automated backup process"
log "Database: $DB_NAME"
log "Host: $DB_HOST:$DB_PORT"
log "Backup Directory: $BACKUP_DIR"
log "=========================================="

# Create timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/leapmailr_backup_${TIMESTAMP}.sql"

# Check if pg_dump is available
if ! command -v pg_dump &> /dev/null; then
    error_exit "pg_dump not found. Please install PostgreSQL client tools."
fi

# Create backup
log "Creating backup: $BACKUP_FILE"
if pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    -F c -f "$BACKUP_FILE" --no-owner --no-acl 2>&1 | tee -a "$LOG_FILE"; then
    log "Backup created successfully"
else
    error_exit "Backup creation failed"
fi

# Verify backup file exists
if [[ ! -f "$BACKUP_FILE" ]]; then
    error_exit "Backup file not found: $BACKUP_FILE"
fi

# Get backup file size
BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
log "Backup size: $BACKUP_SIZE"

# Verify backup integrity
log "Verifying backup integrity..."
if pg_restore --list "$BACKUP_FILE" > /dev/null 2>&1; then
    log "Backup verification successful"
else
    error_exit "Backup verification failed"
fi

# Compress backup (optional)
if command -v gzip &> /dev/null; then
    log "Compressing backup..."
    gzip "$BACKUP_FILE"
    BACKUP_FILE="${BACKUP_FILE}.gz"
    COMPRESSED_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    log "Compressed size: $COMPRESSED_SIZE"
fi

# Cleanup old backups
log "Cleaning up backups older than $RETENTION_DAYS days..."
DELETED_COUNT=0
find "$BACKUP_DIR" -name "leapmailr_backup_*.sql*" -type f -mtime +$RETENTION_DAYS -print0 | while IFS= read -r -d '' file; do
    log "Removing old backup: $file"
    rm -f "$file"
    DELETED_COUNT=$((DELETED_COUNT + 1))
done
log "Removed $DELETED_COUNT old backup(s)"

# Upload to cloud storage (optional - uncomment and configure)
# if command -v aws &> /dev/null; then
#     log "Uploading to S3..."
#     aws s3 cp "$BACKUP_FILE" "s3://your-bucket/backups/" --storage-class STANDARD_IA
#     log "S3 upload completed"
# fi

# Summary
log "=========================================="
log "Backup completed successfully"
log "Backup file: $BACKUP_FILE"
log "Size: $BACKUP_SIZE (compressed: $COMPRESSED_SIZE)"
log "=========================================="

exit 0
