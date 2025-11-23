#!/bin/bash
# Secret Rotation Script for LeapMailR (GAP-SEC-004)
# This script rotates secrets and updates environment configuration

set -e

# Constants
readonly DATE_FORMAT='%Y-%m-%d %H:%M:%S'
readonly SEPARATOR_LINE='=========================================='

# Configuration
SECRETS_DIR="${SECRETS_DIR:-./secrets}"
BACKUP_DIR="${BACKUP_DIR:-./secrets/backups}"
LOG_FILE="${LOG_FILE:-./logs/secret-rotation.log}"
ENV_FILE="${ENV_FILE:-.env}"

# Ensure directories exist
mkdir -p "$SECRETS_DIR"
mkdir -p "$BACKUP_DIR"
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
    return 1
}

log "$SEPARATOR_LINE"
log "SECRET ROTATION STARTED"
log "$SEPARATOR_LINE"

# Check if running in production
if [[ "$ENVIRONMENT" = "production" ]]; then
    read -p "WARNING: This will rotate production secrets. Continue? (yes/no): " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]]; then
        log "Rotation cancelled by user"
        exit 0
    fi
fi

# Function to generate a random secret
generate_secret() {
    local length="${1:-32}"
    openssl rand -base64 "$length" | tr -d "=+/" | cut -c1-"$length"
    return 0
}

# Function to backup current secret
backup_secret() {
    local key="$1"
    local value="$2"
    local timestamp
    timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$BACKUP_DIR/${key}_${timestamp}.bak"
    
    echo "$value" > "$backup_file"
    chmod 600 "$backup_file"
    log "Backed up $key to $backup_file"
    return 0
}

# Function to rotate JWT secret
rotate_jwt_secret() {
    log "Rotating JWT_SECRET..."
    
    # Get current secret
    local current_jwt
    current_jwt=$(grep "^JWT_SECRET=" "$ENV_FILE" | cut -d'=' -f2- || echo "")
    
    if [[ -n "$current_jwt" ]]; then
        backup_secret "JWT_SECRET" "$current_jwt"
    fi
    
    # Generate new secret (64 bytes for HMAC-SHA256)
    local new_jwt
    new_jwt=$(generate_secret 64)
    
    # Update .env file
    if grep -q "^JWT_SECRET=" "$ENV_FILE"; then
        sed -i "s|^JWT_SECRET=.*|JWT_SECRET=$new_jwt|" "$ENV_FILE"
    else
        echo "JWT_SECRET=$new_jwt" >> "$ENV_FILE"
    fi
    
    log "JWT_SECRET rotated successfully"
    return 0
}

# Function to rotate encryption key
rotate_encryption_key() {
    log "Rotating ENCRYPTION_KEY..."
    
    local current_key
    current_key=$(grep "^ENCRYPTION_KEY=" "$ENV_FILE" | cut -d'=' -f2- || echo "")
    
    if [[ -n "$current_key" ]]; then
        backup_secret "ENCRYPTION_KEY" "$current_key"
    fi
    
    # Generate new 32-byte key, base64 encoded
    local new_key
    new_key=$(openssl rand -base64 32)
    
    if grep -q "^ENCRYPTION_KEY=" "$ENV_FILE"; then
        sed -i "s|^ENCRYPTION_KEY=.*|ENCRYPTION_KEY=$new_key|" "$ENV_FILE"
    else
        echo "ENCRYPTION_KEY=$new_key" >> "$ENV_FILE"
    fi
    
    log "ENCRYPTION_KEY rotated successfully"
    log "WARNING: You must re-encrypt existing encrypted data with the new key"
    return 0
}

# Function to rotate database password
rotate_db_password() {
    log "Rotating database password..."
    
    local current_pass
    current_pass=$(grep "^DB_PASSWORD=" "$ENV_FILE" | cut -d'=' -f2- || echo "")
    
    if [[ -n "$current_pass" ]]; then
        backup_secret "DB_PASSWORD" "$current_pass"
    fi
    
    # Generate new password (24 characters with special chars)
    local new_pass
    new_pass=$(openssl rand -base64 24 | tr -d "=+/" | cut -c1-24)
    
    # Update .env file
    if grep -q "^DB_PASSWORD=" "$ENV_FILE"; then
        sed -i "s|^DB_PASSWORD=.*|DB_PASSWORD=$new_pass|" "$ENV_FILE"
    else
        echo "DB_PASSWORD=$new_pass" >> "$ENV_FILE"
    fi
    
    # Update PostgreSQL password
    local db_user
    db_user=$(grep "^DB_USER=" "$ENV_FILE" | cut -d'=' -f2- || echo "postgres")
    log "Updating PostgreSQL password for user: $db_user"
    
    # Connect to PostgreSQL and update password
    PGPASSWORD="$current_pass"
    export PGPASSWORD
    psql -h "${DB_HOST:-localhost}" -U "$db_user" -d postgres -c \
        "ALTER USER $db_user WITH PASSWORD '$new_pass';" || \
        error_exit "Failed to update database password"
    
    log "DB_PASSWORD rotated successfully"
    return 0
}

# Function to rotate SMTP password
rotate_smtp_password() {
    log "Rotating SMTP password..."
    log "NOTE: SMTP passwords must be rotated through your email provider's control panel"
    log "After rotating with your provider, update the .env file manually"
    return 0
}

# Function to rotate API keys
rotate_api_keys() {
    log "Rotating API keys..."
    
    # AWS credentials
    if grep -q "^AWS_ACCESS_KEY_ID=" "$ENV_FILE"; then
        log "NOTE: AWS credentials should be rotated through AWS IAM Console"
        log "After rotation, update AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY in .env"
    fi
    
    # SendGrid/Mailgun/etc
    if grep -q "^SENDGRID_API_KEY=" "$ENV_FILE"; then
        log "NOTE: SendGrid API key should be rotated through SendGrid Dashboard"
        log "After rotation, update SENDGRID_API_KEY in .env"
    fi
    return 0
}

# Function to rotate Redis password (if using Redis)
rotate_redis_password() {
    if grep -q "^REDIS_PASSWORD=" "$ENV_FILE"; then
        log "Rotating Redis password..."
        
        local current_redis
        current_redis=$(grep "^REDIS_PASSWORD=" "$ENV_FILE" | cut -d'=' -f2- || echo "")
        
        if [[ -n "$current_redis" ]]; then
            backup_secret "REDIS_PASSWORD" "$current_redis"
        fi
        
        local new_redis
        new_redis=$(generate_secret 32)
        
        if grep -q "^REDIS_PASSWORD=" "$ENV_FILE"; then
            sed -i "s|^REDIS_PASSWORD=.*|REDIS_PASSWORD=$new_redis|" "$ENV_FILE"
        else
            echo "REDIS_PASSWORD=$new_redis" >> "$ENV_FILE"
        fi
        
        log "REDIS_PASSWORD rotated successfully"
        log "NOTE: Update Redis configuration and restart Redis server"
    fi
    return 0
}

# Main rotation logic
SECRET_TYPE="${1:-all}"

case "$SECRET_TYPE" in
    jwt)
        rotate_jwt_secret
        ;;
    encryption)
        rotate_encryption_key
        ;;
    database)
        rotate_db_password
        ;;
    smtp)
        rotate_smtp_password
        ;;
    api)
        rotate_api_keys
        ;;
    redis)
        rotate_redis_password
        ;;
    all)
        log "Rotating all secrets..."
        rotate_jwt_secret
        rotate_encryption_key
        rotate_db_password
        rotate_smtp_password
        rotate_api_keys
        rotate_redis_password
        ;;
    *)
        echo "Usage: $0 [jwt|encryption|database|smtp|api|redis|all]"
        exit 1
        ;;
esac

# Cleanup old backups (older than 90 days)
log "Cleaning up old secret backups..."
find "$BACKUP_DIR" -name "*.bak" -mtime +90 -delete

log "=========================================="
log "SECRET ROTATION COMPLETED"
log "Backup location: $BACKUP_DIR"
log "=========================================="
log ""
log "IMPORTANT POST-ROTATION STEPS:"
log "1. Restart the application to load new secrets"
log "2. Update Kubernetes secrets if running in K8s"
log "3. Verify application functionality"
log "4. Update documentation with rotation date"
log "5. Notify team members of secret changes"

exit 0
