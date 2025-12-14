# Disaster Recovery Plan - LeapMailR

**Document Version:** 1.0  
**Last Updated:** 2024-11-04  
**Owner:** Infrastructure Team  
**Compliance:** SOC 2 GAP-AV-002

## Executive Summary

This Disaster Recovery (DR) plan defines the procedures, tools, and responsibilities for recovering LeapMailr services in the event of a catastrophic failure. The plan ensures business continuity and data integrity through automated backups, tested restore procedures, and clear recovery objectives.

---

## Recovery Objectives

### RTO (Recovery Time Objective)
- **Critical Services:** 4 hours
- **Database Recovery:** 2 hours
- **Full System Recovery:** 8 hours

### RPO (Recovery Point Objective)
- **Database:** 24 hours (daily backups)
- **Transaction Logs:** 1 hour (with WAL archiving)
- **Configuration Files:** 24 hours

### Service Priority Tiers

**Tier 1 (Critical - RTO: 2h):**
- Database (PostgreSQL)
- Authentication Services
- Health Check Endpoints

**Tier 2 (High - RTO: 4h):**
- Email API Services
- Template Management
- Contact Management

**Tier 3 (Standard - RTO: 8h):**
- Analytics
- Monitoring/Metrics
- Documentation

---

## Backup Strategy

### Database Backups

**Automated Daily Backups:**
```bash
# Cron schedule: Daily at 2:00 AM UTC
0 2 * * * cd /path/to/leapmailr && ./scripts/backup.sh >> ./logs/backup-cron.log 2>&1
```

**Backup Specifications:**
- Format: PostgreSQL custom format (compressed)
- Storage Location: `./backups/`
- Retention: 30 days
- Verification: Automated integrity check after each backup
- Offsite Copy: AWS S3 (automatic upload after verification)

**Backup Types:**

1. **Full Daily Backup**
   - Frequency: Daily at 2:00 AM UTC
   - Contents: Complete database dump
   - Size: ~500MB compressed
   - Duration: ~15 minutes

2. **Transaction Log Archiving (Optional)**
   - Frequency: Continuous (WAL archiving)
   - RPO: 1 hour
   - Storage: Separate WAL archive directory

### Configuration Backups

**Files Backed Up:**
- `config.env`
- Kubernetes manifests (YAML files)
- Docker configurations
- SSL/TLS certificates
- Application configuration files

**Backup Method:**
- Git repository (version controlled)
- Daily automated commits
- Encrypted secrets stored in Kubernetes secrets or vault

---

## Recovery Procedures

### Database Recovery

#### Standard Recovery (From Daily Backup)

**Prerequisites:**
- PostgreSQL 14+ installed
- Backup file available
- Database credentials configured

**Steps:**

1. **Stop Application Services**
   ```bash
   kubectl scale deployment backend --replicas=0
   # OR
   docker-compose down backend
   ```

2. **Locate Latest Backup**
   ```bash
   ls -lh ./backups/
   # Identify latest: leapmailr_backup_YYYYMMDD_HHMMSS.sql.gz
   ```

3. **Verify Backup Integrity**
   ```bash
   gunzip -c ./backups/leapmailr_backup_20241104_020000.sql.gz | \
       pg_restore --list > /dev/null 2>&1
   ```

4. **Execute Restore**
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_NAME=leapmailr
   export DB_USER=postgres
   export PGPASSWORD=your_password
   
   ./scripts/restore.sh ./backups/leapmailr_backup_20241104_020000.sql.gz
   ```

5. **Verify Database Connection**
   ```bash
   psql -h localhost -U postgres -d leapmailr -c "\dt"
   ```

6. **Restart Application Services**
   ```bash
   kubectl scale deployment backend --replicas=3
   # OR
   docker-compose up -d backend
   ```

7. **Verify Application Health**
   ```bash
   curl https://api.leapmailr.com/health
   ```

**Expected Duration:** 30-60 minutes  
**RTO Target:** 2 hours

#### Point-in-Time Recovery (With WAL Archiving)

**Steps:**

1. **Stop PostgreSQL**
   ```bash
   pg_ctl stop -D /var/lib/postgresql/data
   ```

2. **Restore Base Backup**
   ```bash
   rm -rf /var/lib/postgresql/data/*
   pg_restore -D /var/lib/postgresql/data ./backups/base_backup.tar
   ```

3. **Create Recovery Configuration**
   ```bash
   cat > /var/lib/postgresql/data/recovery.conf << EOF
   restore_command = 'cp /path/to/wal_archive/%f %p'
   recovery_target_time = '2024-11-04 12:00:00'
   EOF
   ```

4. **Start PostgreSQL in Recovery Mode**
   ```bash
   pg_ctl start -D /var/lib/postgresql/data
   ```

5. **Monitor Recovery**
   ```bash
   tail -f /var/lib/postgresql/data/pg_log/postgresql.log
   ```

**Expected Duration:** 1-2 hours  
**RPO:** Down to specific transaction

---

## Disaster Scenarios

### Scenario 1: Database Corruption

**Detection:**
- Application errors: "database connection failed"
- PostgreSQL logs: corruption errors
- Health check failures

**Response:**

1. **Immediate Actions (0-15 min):**
   - Alert team via PagerDuty/Slack
   - Enable maintenance mode
   - Stop write operations

2. **Assessment (15-30 min):**
   - Check PostgreSQL logs
   - Attempt `VACUUM FULL ANALYZE`
   - Verify backup availability

3. **Recovery (30-90 min):**
   - Execute database restore procedure
   - Verify data integrity
   - Resume services

**Total RTO:** 2 hours

### Scenario 2: Complete Server Failure

**Detection:**
- All health checks fail
- Server unreachable
- Infrastructure alerts

**Response:**

1. **Immediate Actions (0-15 min):**
   - Confirm failure type (hardware/network/attack)
   - Alert team and stakeholders
   - Activate backup infrastructure

2. **Provisioning (15-60 min):**
   - Spin up new server instance
   - Install dependencies
   - Configure networking

3. **Recovery (60-180 min):**
   - Restore database from S3 backup
   - Deploy application from Git
   - Configure environment variables
   - SSL certificate setup

4. **Verification (180-240 min):**
   - End-to-end testing
   - DNS cutover
   - Monitor for issues

**Total RTO:** 4 hours

### Scenario 3: Ransomware Attack

**Detection:**
- Encrypted files
- Ransom note
- Unusual file system activity

**Response:**

1. **Immediate Actions (0-15 min):**
   - Isolate affected systems
   - Disable network access
   - Alert security team and management
   - DO NOT pay ransom

2. **Assessment (15-60 min):**
   - Identify attack scope
   - Verify backup integrity (offsite S3)
   - Document evidence for authorities

3. **Clean Slate Recovery (60-300 min):**
   - Provision new infrastructure
   - Deploy from clean backups
   - Restore database from S3
   - Reset all credentials
   - Update all secrets

4. **Security Hardening (300-480 min):**
   - Patch vulnerabilities
   - Update firewall rules
   - Enable enhanced monitoring
   - Conduct security audit

**Total RTO:** 8 hours

---

## Testing & Validation

### Quarterly DR Tests

**Schedule:** First Saturday of each quarter at 2:00 AM UTC

**Test Procedure:**

1. **Pre-Test Preparation**
   - Notify team 7 days in advance
   - Schedule maintenance window
   - Prepare test environment

2. **Full Recovery Test**
   - Restore database from latest backup
   - Deploy application in test environment
   - Execute test suite
   - Verify data integrity

3. **Documentation**
   - Record actual RTO/RPO
   - Document issues encountered
   - Update procedures based on lessons learned
   - Share report with stakeholders

**Success Criteria:**
- Database restored successfully
- Application passes health checks
- Data integrity verified (row count, checksums)
- RTO within target (2 hours)
- RPO within target (24 hours)

### Monthly Backup Verification

**Schedule:** 15th of each month

**Steps:**
1. Select random backup from previous month
2. Restore to test environment
3. Run integrity checks
4. Verify row counts against production
5. Document results

---

## Roles & Responsibilities

### Incident Commander
- **Primary:** Infrastructure Lead
- **Backup:** Senior DevOps Engineer
- Responsibilities: Decision making, stakeholder communication, resource coordination

### Database Administrator
- **Primary:** Database Team Lead
- **Backup:** Senior DBA
- Responsibilities: Database recovery, integrity verification, performance validation

### Application Owner
- **Primary:** Backend Team Lead
- **Backup:** Senior Backend Engineer
- Responsibilities: Application deployment, configuration, functional testing

### Communication Lead
- **Primary:** Product Manager
- **Backup:** Customer Success Manager
- Responsibilities: User communication, status updates, documentation

---

## Communication Plan

### Internal Communication

**Incident Declaration:**
- Slack channel: `#incident-response`
- PagerDuty alert
- Email to leadership

**Status Updates:**
- Every 30 minutes during active incident
- Include: status, ETA, blockers, next steps

**Post-Incident:**
- Post-mortem within 48 hours
- Lessons learned documentation
- Process improvements

### External Communication

**User Notification:**
- Status page update (status.leapmailr.com)
- Email to affected customers
- Social media updates (if applicable)

**Templates:**

**Initial Incident:**
```
Subject: Service Disruption - LeapMailR

We are currently experiencing technical difficulties affecting 
[service/feature]. Our team is actively working on resolution.

Expected Resolution: [ETA]
Updates: [status page URL]
```

**Resolution:**
```
Subject: Service Restored - LeapMailR

Service has been fully restored. We apologize for the inconvenience.

Impact Duration: [start] - [end]
Root Cause: [brief explanation]
Prevention: [steps taken]
```

---

## Backup Maintenance

### Daily Tasks
- ✅ Verify automated backup completed
- ✅ Check backup logs for errors
- ✅ Confirm S3 upload successful

### Weekly Tasks
- ✅ Review backup sizes (identify growth)
- ✅ Test backup integrity (random sample)
- ✅ Verify retention cleanup

### Monthly Tasks
- ✅ Full restore test to test environment
- ✅ Review storage costs
- ✅ Update DR documentation
- ✅ Audit access logs

### Quarterly Tasks
- ✅ Full DR drill
- ✅ Review and update RTO/RPO targets
- ✅ Audit backup encryption
- ✅ Capacity planning review

---

## Monitoring & Alerts

### Backup Monitoring

**Metrics Tracked:**
- Backup success/failure rate
- Backup duration
- Backup file size
- S3 upload status

**Alerts Configured:**
- Backup failure (immediate)
- Backup duration exceeds 30 min (warning)
- Backup size increase >20% (warning)
- Missing daily backup (critical)

**Monitoring Tools:**
- Prometheus metrics
- Grafana dashboards
- PagerDuty integration

### Example Prometheus Queries
```promql
# Backup success rate (last 7 days)
rate(backup_success_total[7d]) / rate(backup_attempts_total[7d])

# Backup duration
histogram_quantile(0.95, backup_duration_seconds_bucket)

# Backup size trend
increase(backup_size_bytes[30d])
```

---

## Compliance & Audit

### SOC 2 Requirements

**CC9.1 - Business Continuity:**
- ✅ Documented DR plan with RTO/RPO
- ✅ Regular backup verification
- ✅ Quarterly DR testing
- ✅ Offsite backup storage

**Evidence for Auditors:**
- Backup logs (30 days retention)
- DR test reports (quarterly)
- Backup verification results (monthly)
- Incident response documentation

### Data Retention

**Backup Retention:**
- Daily backups: 30 days
- Monthly backups: 1 year
- Yearly backups: 7 years (compliance)

**Log Retention:**
- Backup logs: 90 days
- Restore logs: 1 year
- DR test logs: 3 years

---

## Appendix

### A. Backup Service API

The BackupService can be integrated into the application for on-demand backups:

```go
import "leapmailr/service"

// Initialize backup service
backupConfig := service.BackupConfig{
    BackupDir:     "./backups",
    DBHost:        "localhost",
    DBPort:        "5432",
    DBName:        "leapmailr",
    DBUser:        "postgres",
    RetentionDays: 30,
}

backupService := service.NewBackupService(logger, backupConfig)

// Create backup
backupPath, err := backupService.CreateBackup(time.Now())
if err != nil {
    log.Fatal(err)
}

// Verify backup
if err := backupService.VerifyBackup(backupPath); err != nil {
    log.Fatal(err)
}

// List all backups
backups, err := backupService.ListBackups()

// Get statistics
stats, err := backupService.GetBackupStats()
```

### B. Quick Reference Commands

```bash
# Manual backup
./scripts/backup.sh

# Restore from specific backup
./scripts/restore.sh ./backups/leapmailr_backup_20241104_020000.sql.gz

# List all backups
ls -lh ./backups/

# Check backup integrity
pg_restore --list ./backups/leapmailr_backup_20241104_020000.sql | head

# Backup to S3
aws s3 cp ./backups/ s3://leapmailr-backups/ --recursive

# Restore from S3
aws s3 cp s3://leapmailr-backups/leapmailr_backup_20241104_020000.sql.gz ./backups/

# Database connection test
psql -h localhost -U postgres -d leapmailr -c "SELECT version();"

# Check database size
psql -h localhost -U postgres -d leapmailr -c \
    "SELECT pg_size_pretty(pg_database_size('leapmailr'));"
```

### C. Contact Information

**On-Call Rotation:**
- PagerDuty: https://leapmailr.pagerduty.com
- Escalation: See PagerDuty schedule

**Vendor Support:**
- AWS Support: [AWS account support page]
- Database Support: [PostgreSQL support]

**Emergency Contacts:**
- Infrastructure Lead: [contact info]
- CTO: [contact info]
- Security Team: security@leapmailr.com

---

## Document Revision History

| Version | Date       | Author | Changes                     |
|---------|------------|--------|-----------------------------|
| 1.0     | 2024-11-04 | Infra  | Initial DR plan creation    |

---

## Approval

- [ ] Infrastructure Lead: _________________ Date: _________
- [ ] CTO: _________________ Date: _________
- [ ] Security Lead: _________________ Date: _________

**Next Review Date:** May 1, 2025
