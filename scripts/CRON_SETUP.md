# Cron Job Configuration for LeapMailR Backups

# ============================================
# AUTOMATED DATABASE BACKUPS
# ============================================
# Schedule: Daily at 2:00 AM UTC
# Retention: 30 days
# Location: ./backups/
# ============================================

# Required environment variables (set in /etc/environment or crontab):
# DB_HOST=localhost
# DB_PORT=5432
# DB_NAME=leapmailr
# DB_USER=postgres
# PGPASSWORD=your_secure_password
# BACKUP_DIR=/path/to/leapmailr/backups
# RETENTION_DAYS=30

# ============================================
# CRON SCHEDULE EXAMPLES
# ============================================

# Daily backup at 2:00 AM UTC
0 2 * * * cd /path/to/leapmailr && ./scripts/backup.sh >> ./logs/backup-cron.log 2>&1

# Twice daily backups (2 AM and 2 PM UTC)
0 2,14 * * * cd /path/to/leapmailr && ./scripts/backup.sh >> ./logs/backup-cron.log 2>&1

# Every 6 hours
0 */6 * * * cd /path/to/leapmailr && ./scripts/backup.sh >> ./logs/backup-cron.log 2>&1

# Weekly backup on Sunday at 3:00 AM UTC
0 3 * * 0 cd /path/to/leapmailr && ./scripts/backup.sh >> ./logs/backup-cron.log 2>&1

# ============================================
# INSTALLATION INSTRUCTIONS
# ============================================

# 1. Set environment variables
# Edit /etc/environment or create .env file:
cat >> /etc/environment << EOF
DB_HOST=localhost
DB_PORT=5432
DB_NAME=leapmailr
DB_USER=postgres
PGPASSWORD=your_secure_password
BACKUP_DIR=/opt/leapmailr/backups
RETENTION_DAYS=30
LOG_FILE=/opt/leapmailr/logs/backup.log
EOF

# 2. Make backup script executable
chmod +x /path/to/leapmailr/scripts/backup.sh

# 3. Create necessary directories
mkdir -p /path/to/leapmailr/backups
mkdir -p /path/to/leapmailr/logs

# 4. Edit crontab for the appropriate user
crontab -e

# 5. Add the cron job (paste one of the schedules above)

# 6. Verify crontab installation
crontab -l

# 7. Test the backup script manually first
cd /path/to/leapmailr
./scripts/backup.sh

# 8. Check logs after scheduled run
tail -f /path/to/leapmailr/logs/backup-cron.log

# ============================================
# MONITORING & ALERTS
# ============================================

# Monitor backup success/failure
# Add this to your monitoring system (Prometheus, Datadog, etc.)

# Example: Check if backup was created today
0 4 * * * [ -f /path/to/leapmailr/backups/leapmailr_backup_$(date +\%Y\%m\%d)*.sql* ] || echo "Backup failed for $(date)" | mail -s "Backup Alert" admin@leapmailr.com

# Example: Alert if backup log contains errors
0 4 * * * grep -i error /path/to/leapmailr/logs/backup-cron.log && echo "Backup errors detected" | mail -s "Backup Error Alert" admin@leapmailr.com

# ============================================
# KUBERNETES CRONJOB (Alternative)
# ============================================

# If running in Kubernetes, use CronJob instead:
# Create file: k8s/backup-cronjob.yaml

apiVersion: batch/v1
kind: CronJob
metadata:
  name: leapmailr-backup
  namespace: leapmailr
spec:
  schedule: "0 2 * * *"  # 2 AM UTC daily
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:14
            command:
            - /bin/bash
            - -c
            - |
              pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
                -F c -f /backups/leapmailr_backup_$(date +%Y%m%d_%H%M%S).sql \
                --no-owner --no-acl
              
              # Upload to S3
              aws s3 cp /backups/leapmailr_backup_*.sql \
                s3://leapmailr-backups/ --recursive
              
              # Cleanup old backups
              find /backups -name "leapmailr_backup_*.sql" -mtime +30 -delete
            env:
            - name: DB_HOST
              valueFrom:
                configMapKeyRef:
                  name: leapmailr-config
                  key: db_host
            - name: DB_PORT
              valueFrom:
                configMapKeyRef:
                  name: leapmailr-config
                  key: db_port
            - name: DB_NAME
              valueFrom:
                configMapKeyRef:
                  name: leapmailr-config
                  key: db_name
            - name: DB_USER
              valueFrom:
                secretKeyRef:
                  name: leapmailr-secrets
                  key: db_user
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: leapmailr-secrets
                  key: db_password
            - name: AWS_ACCESS_KEY_ID
              valueFrom:
                secretKeyRef:
                  name: aws-credentials
                  key: access_key_id
            - name: AWS_SECRET_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: aws-credentials
                  key: secret_access_key
            volumeMounts:
            - name: backup-storage
              mountPath: /backups
          restartPolicy: OnFailure
          volumes:
          - name: backup-storage
            persistentVolumeClaim:
              claimName: backup-pvc

# Deploy Kubernetes CronJob:
# kubectl apply -f k8s/backup-cronjob.yaml

# ============================================
# DOCKER COMPOSE (Alternative)
# ============================================

# Add to docker-compose.yml:
services:
  backup:
    image: postgres:14
    environment:
      - DB_HOST=db
      - DB_PORT=5432
      - DB_NAME=leapmailr
      - DB_USER=postgres
      - PGPASSWORD=${DB_PASSWORD}
    volumes:
      - ./backups:/backups
      - ./scripts:/scripts
    command: >
      sh -c "
        echo '0 2 * * * /scripts/backup.sh' > /etc/crontabs/root &&
        crond -f -l 2
      "
    depends_on:
      - db

# ============================================
# TROUBLESHOOTING
# ============================================

# Cron job not running:
# - Check cron service: systemctl status cron
# - Check crontab syntax: crontab -l
# - Check script permissions: ls -l /path/to/scripts/backup.sh
# - Check environment variables: printenv | grep DB_
# - Check logs: tail -f /var/log/syslog | grep CRON

# Permission errors:
# - Ensure user has write access to backup directory
# - Check script ownership: chown user:user /path/to/scripts/backup.sh
# - Verify PostgreSQL authentication

# Backup failures:
# - Test script manually: ./scripts/backup.sh
# - Check disk space: df -h
# - Verify database connection: psql -h localhost -U postgres -d leapmailr
# - Review backup logs: cat ./logs/backup.log

# ============================================
# SECURITY BEST PRACTICES
# ============================================

# 1. Use PostgreSQL .pgpass file instead of PGPASSWORD in crontab
# Create ~/.pgpass:
echo "localhost:5432:leapmailr:postgres:your_password" > ~/.pgpass
chmod 600 ~/.pgpass

# 2. Encrypt backups at rest
# Add to backup.sh after gzip:
# openssl enc -aes-256-cbc -salt -in backup.sql.gz -out backup.sql.gz.enc -k $ENCRYPTION_KEY

# 3. Rotate backup encryption keys quarterly

# 4. Use dedicated backup user with minimal privileges
# CREATE ROLE backup_user WITH LOGIN PASSWORD 'secure_password';
# GRANT CONNECT ON DATABASE leapmailr TO backup_user;
# GRANT SELECT ON ALL TABLES IN SCHEMA public TO backup_user;

# 5. Store credentials in secure secret management
# - Kubernetes Secrets
# - HashiCorp Vault
# - AWS Secrets Manager
