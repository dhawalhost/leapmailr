# Database Backup Setup Guide

This guide covers setting up automated database backups for Leapmailr to meet SOC 2 compliance requirements (GAP-AV-001).

## Overview

The backup system:
- Creates compressed PostgreSQL backups daily
- Uploads backups to AWS S3 for offsite storage
- Verifies backup integrity before uploading
- Maintains 30 days of backup retention (configurable)
- Keeps last 7 backups locally (configurable)
- Logs all operations for audit trail

## Prerequisites

### 1. PostgreSQL Client Tools
The backup scripts use `pg_dump` to create database backups.

**Linux/Mac:**
```bash
# Ubuntu/Debian
sudo apt-get install postgresql-client

# CentOS/RHEL
sudo yum install postgresql

# Mac
brew install postgresql
```

**Windows:**
- Install PostgreSQL (includes pg_dump)
- Or download PostgreSQL binaries

### 2. AWS CLI
Required for uploading backups to S3.

**Linux/Mac:**
```bash
# Using pip
pip install awscli

# Or using package manager
# Ubuntu/Debian
sudo apt-get install awscli

# Mac
brew install awscli
```

**Windows:**
- Download AWS CLI installer from: https://aws.amazon.com/cli/
- Run the MSI installer

**Configure AWS CLI:**
```bash
aws configure
# Enter:
# - AWS Access Key ID
# - AWS Secret Access Key
# - Default region (e.g., us-east-1)
# - Output format (json)
```

### 3. S3 Bucket
Create an S3 bucket for storing backups.

```bash
# Create bucket
aws s3 mb s3://leapmailr-backups --region us-east-1

# Enable versioning (recommended)
aws s3api put-bucket-versioning \
  --bucket leapmailr-backups \
  --versioning-configuration Status=Enabled

# Enable encryption
aws s3api put-bucket-encryption \
  --bucket leapmailr-backups \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'

# Set lifecycle policy to transition old backups to Glacier
aws s3api put-bucket-lifecycle-configuration \
  --bucket leapmailr-backups \
  --lifecycle-configuration file://lifecycle-policy.json
```

**lifecycle-policy.json:**
```json
{
  "Rules": [
    {
      "Id": "Archive old backups",
      "Status": "Enabled",
      "Filter": {
        "Prefix": "leapmailr/backups/"
      },
      "Transitions": [
        {
          "Days": 30,
          "StorageClass": "GLACIER"
        }
      ],
      "Expiration": {
        "Days": 365
      }
    }
  ]
}
```

## Setup Instructions

### Linux/Mac Setup

#### 1. Copy and Configure Script
```bash
# Copy script
sudo cp scripts/backup-db.sh /opt/leapmailr/scripts/backup-db.sh

# Make executable
sudo chmod +x /opt/leapmailr/scripts/backup-db.sh

# Create directories
sudo mkdir -p /var/backups/leapmailr
sudo mkdir -p /var/log

# Edit configuration
sudo nano /opt/leapmailr/scripts/backup-db.sh
```

**Update these variables:**
```bash
DB_NAME="leapmailr"
DB_USER="postgres"
DB_HOST="localhost"
DB_PORT="5432"
DB_PASSWORD="your_password"  # Better: use .pgpass file

BACKUP_DIR="/var/backups/leapmailr"
S3_BUCKET="s3://leapmailr-backups/leapmailr/backups"
RETENTION_DAYS="30"
MAX_LOCAL_BACKUPS="7"

ALERT_EMAIL="admin@example.com"
```

#### 2. Secure Database Password

Instead of storing password in script, use `.pgpass` file:

```bash
# Create .pgpass file
echo "localhost:5432:leapmailr:postgres:your_password" > ~/.pgpass

# Set permissions
chmod 600 ~/.pgpass

# Remove DB_PASSWORD from script
```

#### 3. Test Backup Manually
```bash
# Run backup script
sudo /opt/leapmailr/scripts/backup-db.sh

# Check output
tail -f /var/log/leapmailr-backup.log

# Verify S3 upload
aws s3 ls s3://leapmailr-backups/leapmailr/backups/
```

#### 4. Schedule with Cron

```bash
# Edit crontab
sudo crontab -e

# Add this line (runs daily at 2 AM)
0 2 * * * /opt/leapmailr/scripts/backup-db.sh >> /var/log/leapmailr-backup.log 2>&1

# Save and exit

# List cron jobs to verify
sudo crontab -l

# Check cron logs
sudo tail -f /var/log/cron
```

**Alternative cron schedules:**
```bash
# Every 12 hours
0 */12 * * * /opt/leapmailr/scripts/backup-db.sh

# Every 6 hours
0 */6 * * * /opt/leapmailr/scripts/backup-db.sh

# Every day at 2 AM and 2 PM
0 2,14 * * * /opt/leapmailr/scripts/backup-db.sh

# Weekly on Sundays at 3 AM
0 3 * * 0 /opt/leapmailr/scripts/backup-db.sh
```

### Windows Setup

#### 1. Configure Script
```bash
# Edit configuration in backup-db.bat
notepad C:\leapmailr\scripts\backup-db.bat
```

**Update these variables:**
```batch
set "DB_NAME=leapmailr"
set "DB_USER=postgres"
set "DB_HOST=localhost"
set "DB_PORT=5432"
set "PGPASSWORD=your_password"

set "BACKUP_DIR=C:\backups\leapmailr"
set "S3_BUCKET=s3://leapmailr-backups/leapmailr/backups"

set "PG_BIN=C:\Program Files\PostgreSQL\15\bin"
```

#### 2. Test Backup Manually
```bash
# Run from Command Prompt as Administrator
C:\leapmailr\scripts\backup-db.bat

# Check log
type C:\logs\leapmailr-backup_*.log

# Verify S3 upload
aws s3 ls s3://leapmailr-backups/leapmailr/backups/
```

#### 3. Schedule with Task Scheduler

1. Open Task Scheduler (Win + R → `taskschd.msc`)
2. Click "Create Basic Task"
3. Fill in details:
   - **Name:** Leapmailr Database Backup
   - **Description:** Daily automated database backup with S3 upload
4. **Trigger:** Daily
   - **Start:** Today at 2:00 AM
   - **Recur every:** 1 day
5. **Action:** Start a program
   - **Program:** `C:\leapmailr\scripts\backup-db.bat`
   - **Start in:** `C:\leapmailr\scripts`
6. **Finish** and check:
   - ☑ Open Properties dialog
7. In Properties:
   - **Security options:**
     - ☑ Run whether user is logged on or not
     - ☑ Run with highest privileges
   - **Settings:**
     - ☑ If task fails, restart every 10 minutes (3 attempts)
     - ☑ Stop task if runs longer than 1 hour

## Monitoring and Verification

### Check Backup Logs
```bash
# Linux/Mac
tail -f /var/log/leapmailr-backup.log

# Windows
type C:\logs\leapmailr-backup_*.log
```

### List S3 Backups
```bash
# List all backups
aws s3 ls s3://leapmailr-backups/leapmailr/backups/

# List recent backups (last 10)
aws s3 ls s3://leapmailr-backups/leapmailr/backups/ --human-readable | tail -10

# Get total backup size
aws s3 ls s3://leapmailr-backups/leapmailr/backups/ --summarize --human-readable
```

### Verify Backup Integrity
```bash
# Download a backup
aws s3 cp s3://leapmailr-backups/leapmailr/backups/leapmailr_20250104_020000.sql.gz /tmp/

# Test decompression
gunzip -t /tmp/leapmailr_20250104_020000.sql.gz

# If successful, no output means file is good
```

## Disaster Recovery

### Restore from Backup

#### 1. Download Backup
```bash
# List available backups
aws s3 ls s3://leapmailr-backups/leapmailr/backups/

# Download specific backup
aws s3 cp s3://leapmailr-backups/leapmailr/backups/leapmailr_20250104_020000.sql.gz /tmp/
```

#### 2. Decompress
```bash
# Linux/Mac
gunzip /tmp/leapmailr_20250104_020000.sql.gz

# Windows (PowerShell)
Expand-Archive -Path C:\temp\leapmailr_20250104_020000.sql.gz -DestinationPath C:\temp\
```

#### 3. Restore Database
```bash
# Create new database (if needed)
createdb -U postgres leapmailr_restored

# Restore backup
psql -U postgres -d leapmailr_restored < /tmp/leapmailr_20250104_020000.sql

# Or restore to existing database (WARNING: This will overwrite data!)
psql -U postgres -d leapmailr < /tmp/leapmailr_20250104_020000.sql
```

#### 4. Verify Restoration
```bash
# Connect to database
psql -U postgres -d leapmailr_restored

# Check tables
\dt

# Check row counts
SELECT 'users' as table_name, COUNT(*) FROM users
UNION ALL
SELECT 'email_services', COUNT(*) FROM email_services
UNION ALL
SELECT 'email_logs', COUNT(*) FROM email_logs;

# Exit
\q
```

## Troubleshooting

### Common Issues

**1. pg_dump not found**
```bash
# Linux: Check PostgreSQL client installation
which pg_dump

# Add to PATH if needed
export PATH=$PATH:/usr/lib/postgresql/15/bin

# Windows: Update PG_BIN variable in script
```

**2. AWS credentials error**
```bash
# Verify AWS configuration
aws sts get-caller-identity

# Reconfigure if needed
aws configure
```

**3. S3 upload fails**
```bash
# Check S3 bucket exists
aws s3 ls s3://leapmailr-backups/

# Check permissions
aws s3api get-bucket-acl --bucket leapmailr-backups

# Test upload manually
echo "test" > test.txt
aws s3 cp test.txt s3://leapmailr-backups/test.txt
```

**4. Insufficient disk space**
```bash
# Check available space
df -h /var/backups/leapmailr

# Clean up old backups manually
ls -lt /var/backups/leapmailr/
rm /var/backups/leapmailr/leapmailr_OLD_TIMESTAMP.sql.gz
```

**5. Database connection fails**
```bash
# Test connection
psql -h localhost -p 5432 -U postgres -d leapmailr -c "SELECT version();"

# Check PostgreSQL is running
sudo systemctl status postgresql

# Check firewall
sudo ufw status
```

## Security Best Practices

1. **Never commit credentials to Git**
   - Use .pgpass for database passwords
   - Use IAM roles for S3 access (on AWS EC2)
   - Use environment variables

2. **Encrypt backups**
   - S3 bucket encryption enabled (AES-256)
   - Consider client-side encryption for sensitive data

3. **Restrict S3 access**
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "s3:PutObject",
           "s3:GetObject",
           "s3:ListBucket",
           "s3:DeleteObject"
         ],
         "Resource": [
           "arn:aws:s3:::leapmailr-backups/*",
           "arn:aws:s3:::leapmailr-backups"
         ]
       }
     ]
   }
   ```

4. **Monitor backup status**
   - Set up CloudWatch alarms for failed backups
   - Review backup logs weekly
   - Test restores quarterly

5. **Backup verification**
   - Automated integrity checks (gzip -t)
   - Periodic test restores
   - Document restore procedures

## Cost Estimation

### S3 Storage Costs (us-east-1)

**Example:** 100GB database, daily backups, 30-day retention

- Compressed backup size: ~20GB (5:1 compression)
- Monthly backups: 30 backups × 20GB = 600GB
- Cost: ~$13.80/month (Standard-IA at $0.023/GB)

**With Glacier transition (after 30 days):**
- Recent backups (30 days): 600GB × $0.023 = $13.80
- Archived backups (335 days): 6,700GB × $0.004 = $26.80
- **Total annual cost:** ~$487

### Recommendations

- **Standard-IA:** Recent backups (0-30 days)
- **Glacier:** Archived backups (30-365 days)
- **Delete:** Backups older than 1 year

## Compliance Notes

This backup system addresses:
- **GAP-AV-001:** Automated database backups
- **CC7.2:** System backup and recovery procedures
- **CC9.1:** Confidential data protection (encrypted backups)

**Audit Evidence:**
- Backup logs: `/var/log/leapmailr-backup.log`
- S3 bucket: `s3://leapmailr-backups/leapmailr/backups/`
- Cron configuration: `crontab -l`
- Restoration test records: Document quarterly tests

## Support

For issues or questions:
- Check logs first
- Review troubleshooting section
- Contact DevOps team
- Email: devops@example.com
