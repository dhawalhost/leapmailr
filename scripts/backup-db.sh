#!/bin/bash
#
# Automated PostgreSQL Database Backup Script with S3 Upload
# GAP-AV-001: Automated database backups for disaster recovery
#
# This script:
# 1. Creates a compressed PostgreSQL backup
# 2. Uploads to S3 for offsite storage
# 3. Verifies the upload succeeded
# 4. Cleans up old local and remote backups
#
# Prerequisites:
#   - PostgreSQL client tools (pg_dump)
#   - AWS CLI configured with credentials
#   - S3 bucket created for backups

# Constants
readonly DATE_FORMAT='%Y-%m-%d %H:%M:%S'
readonly SEPARATOR_LINE='========================================='
#   - Proper permissions on backup directory
#
# Setup:
#   1. Copy this script to /opt/leapmailr/scripts/backup-db.sh
#   2. Make executable: chmod +x /opt/leapmailr/scripts/backup-db.sh
#   3. Update configuration variables below
#   4. Test manually: ./backup-db.sh
#   5. Add to crontab: crontab -e
#      0 2 * * * /opt/leapmailr/scripts/backup-db.sh >> /var/log/leapmailr-backup.log 2>&1
#
# Monitoring:
#   - Check logs: tail -f /var/log/leapmailr-backup.log
#   - List S3 backups: aws s3 ls s3://your-bucket-name/leapmailr/backups/
#

set -e  # Exit immediately if a command exits with a non-zero status
set -o pipefail  # Fail on pipe errors

#
# ========== CONFIGURATION ==========
# Update these variables for your environment
#

# Database configuration
DB_NAME="${DB_NAME:-leapmailr}"
DB_USER="${DB_USER:-postgres}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"

# Backup configuration
BACKUP_DIR="${BACKUP_DIR:-/var/backups/leapmailr}"
S3_BUCKET="${S3_BUCKET:-s3://your-bucket-name/leapmailr/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"  # Keep backups for 30 days
MAX_LOCAL_BACKUPS="${MAX_LOCAL_BACKUPS:-7}"  # Keep last 7 local backups

# Notification configuration
ALERT_EMAIL="${ALERT_EMAIL:-admin@example.com}"
SEND_SUCCESS_EMAIL="${SEND_SUCCESS_EMAIL:-false}"  # Set to true to get success emails
SEND_FAILURE_EMAIL="${SEND_FAILURE_EMAIL:-true}"   # Always notify on failures

# Compression level (1-9, 9 is best compression but slowest)
COMPRESSION_LEVEL="${COMPRESSION_LEVEL:-6}"

#
# ========== SCRIPT STARTS HERE ==========
# Do not modify below unless you know what you're doing
#

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Timestamp for this backup
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DATE_ONLY=$(date +%Y%m%d)
BACKUP_FILE="${BACKUP_DIR}/leapmailr_${TIMESTAMP}.sql.gz"
BACKUP_FILENAME="leapmailr_${TIMESTAMP}.sql.gz"

# Log file for this run
LOG_FILE="/var/log/leapmailr-backup_${DATE_ONLY}.log"

# Function to log messages
log() {
    local message="$1"
    echo "[$(date +"$DATE_FORMAT")] $message" | tee -a "$LOG_FILE"
    return 0
}

# Function to log errors
log_error() {
    local message="$1"
    echo -e "${RED}[$(date +"$DATE_FORMAT")] ERROR: $message${NC}" | tee -a "$LOG_FILE" >&2
    return 0
}

# Function to log success
log_success() {
    local message="$1"
    echo -e "${GREEN}[$(date +"$DATE_FORMAT")] SUCCESS: $message${NC}" | tee -a "$LOG_FILE"
    return 0
}

# Function to log warnings
log_warning() {
    local message="$1"
    echo -e "${YELLOW}[$(date +"$DATE_FORMAT")] WARNING: $message${NC}" | tee -a "$LOG_FILE"
    return 0
}

# Function to send email alert
send_email() {
    local subject="$1"
    local body="$2"
    
    if command -v mail &> /dev/null; then
        echo "$body" | mail -s "$subject" "$ALERT_EMAIL"
    else
        log_warning "mail command not found. Cannot send email alert."
    fi
    return 0
}

# Function to cleanup on error
cleanup_on_error() {
    log_error "Backup failed. Cleaning up..."
    if [[ -f "$BACKUP_FILE" ]]; then
        rm -f "$BACKUP_FILE"
    fi
    return 0
}
        log "Removed incomplete backup file: $BACKUP_FILE"
    fi
    
    if [[ "$SEND_FAILURE_EMAIL" = "true" ]]; then
        send_email "❌ Leapmailr Backup Failed - $(date +%Y-%m-%d)" \
            "Database backup for $DB_NAME failed at $(date).
            
Check logs at: $LOG_FILE
Server: $(hostname)
Database: $DB_NAME
Error details are in the log file."
    fi
    
    exit 1
}

# Set error trap
trap cleanup_on_error ERR

#
# ========== MAIN BACKUP PROCESS ==========
#

log "$SEPARATOR_LINE"
log "Starting Leapmailr Database Backup"
log "$SEPARATOR_LINE"
log "Database: $DB_NAME"
log "Backup file: $BACKUP_FILE"
log "S3 destination: $S3_BUCKET/"

# Check prerequisites
log "Checking prerequisites..."

# Check if pg_dump exists
if ! command -v pg_dump &> /dev/null; then
    log_error "pg_dump command not found. Please install PostgreSQL client tools."
    cleanup_on_error
fi

# Check if AWS CLI exists
if ! command -v aws &> /dev/null; then
    log_error "AWS CLI not found. Please install and configure AWS CLI."
    cleanup_on_error
fi

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    log_error "AWS credentials not configured. Run 'aws configure'."
    cleanup_on_error
fi

log_success "All prerequisites met"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"
log "Backup directory ready: $BACKUP_DIR"

# Check available disk space (require at least 1GB free)
AVAILABLE_SPACE=$(df -BG "$BACKUP_DIR" | tail -1 | awk '{print $4}' | sed 's/G//')
if [[ "$AVAILABLE_SPACE" -lt 1 ]]; then
    log_error "Insufficient disk space. Available: ${AVAILABLE_SPACE}GB, Required: 1GB"
    cleanup_on_error
fi

# Step 1: Create database backup
log "Creating database backup..."
START_TIME=$(date +%s)

# Use pg_dump with compression
# --no-owner: Don't dump ownership commands
# --no-acl: Don't dump access privileges
# -F c: Custom format (allows parallel restore)
# -Z $COMPRESSION_LEVEL: Compression level
PGPASSWORD="$DB_PASSWORD" pg_dump \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --no-owner \
    --no-acl \
    | gzip -$COMPRESSION_LEVEL > "$BACKUP_FILE"

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

# Check if backup file was created
if [[ ! -f "$BACKUP_FILE" ]]; then
    log_error "Backup file was not created"
    cleanup_on_error
fi

# Get backup file size
BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
log_success "Database backup created successfully"
log "Backup size: $BACKUP_SIZE"
log "Time taken: ${DURATION} seconds"

# Step 2: Verify backup integrity
log "Verifying backup integrity..."
if gzip -t "$BACKUP_FILE" 2>/dev/null; then
    log_success "Backup file integrity verified"
else
    log_error "Backup file is corrupted"
    cleanup_on_error
fi

# Step 3: Upload to S3
log "Uploading to S3..."
UPLOAD_START=$(date +%s)

aws s3 cp "$BACKUP_FILE" "$S3_BUCKET/" \
    --storage-class STANDARD_IA \
    --metadata "database=$DB_NAME,timestamp=$TIMESTAMP,server=$(hostname)" \
    --only-show-errors

if [[ $? -eq 0 ]]; then
    UPLOAD_END=$(date +%s)
    UPLOAD_DURATION=$((UPLOAD_END - UPLOAD_START))
    log_success "Uploaded to S3 in ${UPLOAD_DURATION} seconds"
else
    log_error "S3 upload failed"
    cleanup_on_error
fi

# Step 4: Verify S3 upload
log "Verifying S3 upload..."
S3_FILE_SIZE=$(aws s3 ls "$S3_BUCKET/$BACKUP_FILENAME" --human-readable | awk '{print $3 " " $4}')
if [[ -n "$S3_FILE_SIZE" ]]; then
    log_success "S3 verification passed. Remote size: $S3_FILE_SIZE"
else
    log_error "S3 verification failed - file not found on S3"
    cleanup_on_error
fi

# Step 5: Calculate checksum for audit trail
log "Calculating checksum..."
CHECKSUM=$(md5sum "$BACKUP_FILE" | awk '{print $1}')
log "Backup checksum (MD5): $CHECKSUM"

# Store checksum in S3 metadata
aws s3api copy-object \
    --bucket "${S3_BUCKET#s3://}" \
    --copy-source "${S3_BUCKET#s3://}/$BACKUP_FILENAME" \
    --key "$BACKUP_FILENAME" \
    --metadata "database=$DB_NAME,timestamp=$TIMESTAMP,server=$(hostname),checksum=$CHECKSUM" \
    --metadata-directive REPLACE \
    --only-show-errors

# Step 6: Clean up old local backups
log "Cleaning up old local backups (keeping last $MAX_LOCAL_BACKUPS)..."
LOCAL_BACKUP_COUNT=$(ls -1 "$BACKUP_DIR"/leapmailr_*.sql.gz 2>/dev/null | wc -l)
if [[ "$LOCAL_BACKUP_COUNT" -gt "$MAX_LOCAL_BACKUPS" ]]; then
    OLD_BACKUPS=$(ls -1t "$BACKUP_DIR"/leapmailr_*.sql.gz | tail -n +$((MAX_LOCAL_BACKUPS + 1)))
    echo "$OLD_BACKUPS" | while read -r old_file; do
        rm -f "$old_file"
        log "Deleted old local backup: $(basename "$old_file")"
    done
    log_success "Local cleanup completed"
else
    log "No local cleanup needed ($LOCAL_BACKUP_COUNT backups)"
fi

# Step 7: Clean up old S3 backups
log "Cleaning up old S3 backups (retention: $RETENTION_DAYS days)..."
CUTOFF_DATE=$(date -d "$RETENTION_DAYS days ago" +%Y%m%d || date -v -${RETENTION_DAYS}d +%Y%m%d 2>/dev/null)

aws s3 ls "$S3_BUCKET/" | while read -r line; do
    FILE_DATE=$(echo "$line" | awk '{print $4}' | grep -oP 'leapmailr_\K\d{8}' || echo "")
    FILE_NAME=$(echo "$line" | awk '{print $4}')
    
    if [[ -n "$FILE_DATE" ]] && [ "$FILE_DATE" -lt "$CUTOFF_DATE" ]]; then
        aws s3 rm "$S3_BUCKET/$FILE_NAME" --only-show-errors
        log "Deleted old S3 backup: $FILE_NAME (date: $FILE_DATE)"
    fi
done

log_success "S3 cleanup completed"

# Step 8: Log backup details to database (optional - requires application integration)
# This could be implemented in the future to track backups in the application

# Final summary
log "$SEPARATOR_LINE"
log_success "BACKUP COMPLETED SUCCESSFULLY"
log "$SEPARATOR_LINE"
log "Summary:"
log "  - Database: $DB_NAME"
log "  - Backup file: $BACKUP_FILENAME"
log "  - Size: $BACKUP_SIZE"
log "  - Duration: ${DURATION} seconds"
log "  - S3 location: $S3_BUCKET/$BACKUP_FILENAME"
log "  - Checksum: $CHECKSUM"
log "$SEPARATOR_LINE"

# Send success email if configured
if [[ "$SEND_SUCCESS_EMAIL" = "true" ]]; then
    send_email "✅ Leapmailr Backup Successful - $(date +%Y-%m-%d)" \
        "Database backup completed successfully at $(date).

Details:
- Database: $DB_NAME
- Backup file: $BACKUP_FILENAME
- Size: $BACKUP_SIZE
- Duration: ${DURATION} seconds
- S3 location: $S3_BUCKET/$BACKUP_FILENAME
- Checksum (MD5): $CHECKSUM

Server: $(hostname)
Log file: $LOG_FILE"
fi

# Exit successfully
exit 0
