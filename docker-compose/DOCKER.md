# Docker Setup Guide for LeapMailr

Complete guide for running LeapMailr with Docker and Docker Compose.

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Services](#services)
- [Common Commands](#common-commands)
- [Development Workflow](#development-workflow)
- [Production Deployment](#production-deployment)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Software

1. **Docker Desktop**
   - Windows/Mac: [Download Docker Desktop](https://www.docker.com/products/docker-desktop)
   - Linux: Install Docker Engine and Docker Compose

2. **System Requirements**
   - 4GB+ RAM
   - 10GB+ free disk space
   - Ports available: 3000, 8080, 5432

### Verify Installation

```bash
# Check Docker
docker --version
# Expected: Docker version 20.10.0 or higher

# Check Docker Compose
docker-compose --version
# Expected: Docker Compose version 2.0.0 or higher
```

## Quick Start

### Option 1: Using Setup Scripts (Recommended)

**Windows:**
```cmd
cd docker-compose
docker-setup.bat
```

**Linux/Mac:**
```bash
cd docker-compose
chmod +x docker-setup.sh
./docker-setup.sh
```

### Option 2: Manual Setup

```bash
# 1. Navigate to docker-compose directory
cd docker-compose

# 2. Copy environment file
cp .env.example .env

# 3. Edit .env with your settings
nano .env  # or use your preferred editor

# 4. Start all services
docker-compose up -d

# 5. View logs
docker-compose logs -f
```

### Access Your Application

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **API Health**: http://localhost:8080/api/v1/health

## Configuration

### Environment Variables

The `.env` file contains all configuration. Key variables:

#### Database
```env
DB_HOST=postgres
DB_PORT=5432
DB_USER=leapmailr
DB_PASSWORD=leapmailr_secret_2024  # Change in production!
DB_NAME=leapmailr
```

#### JWT Authentication
```env
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY=24h
REFRESH_TOKEN_EXPIRY=168h
```

#### CORS Settings
```env
CORS_ORIGINS=http://localhost:3000,http://frontend:3000
```

### Email Service Configuration

Email providers (SMTP, SendGrid, Mailgun, etc.) are now configured through the Email Service API after authentication. 

**No environment variables are required for email configuration.**

After starting the application and registering/logging in:
1. Use the `/api/v1/email-services` API endpoints to add email service configurations
2. Set a default email service for your account
3. Test connectivity before sending emails

For API documentation, see the main README or use the API health endpoint.

## Services

### Architecture Overview

```
┌─────────────────┐
│   Frontend      │  Port 3000 (Next.js)
│   (leapmailr-ui)│
└────────┬────────┘
         │
         │ HTTP API
         ▼
┌─────────────────┐
│   Backend       │  Port 8080 (Go/Gin)
│   (leapmailr)   │
└────────┬────────┘
         │
         │ PostgreSQL
         ▼
┌─────────────────┐
│   Database      │  Port 5432
│   (postgres)    │
└─────────────────┘
```

### Service Details

#### 1. PostgreSQL Database
- **Container**: `leapmailr-postgres`
- **Image**: `postgres:15-alpine`
- **Port**: 5432
- **Volume**: `postgres_data` (persistent)
- **Health Check**: Runs every 10 seconds

#### 2. Backend API
- **Container**: `leapmailr-backend`
- **Build**: `../Dockerfile`
- **Port**: 8080
- **Depends On**: postgres (healthy)
- **Health Check**: `/api/v1/health` endpoint

#### 3. Frontend UI
- **Container**: `leapmailr-frontend`
- **Build**: `../../leapmailr-ui/Dockerfile`
- **Port**: 3000
- **Depends On**: backend
- **Features**: Server-side rendering, optimized build

## Common Commands

### Starting & Stopping

```bash
# Start all services in background
docker-compose up -d

# Start and view logs
docker-compose up

# Stop all services
docker-compose down

# Stop and remove volumes (deletes data!)
docker-compose down -v
```

### Viewing Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f postgres

# Last 100 lines
docker-compose logs --tail=100 backend
```

### Service Management

```bash
# Restart specific service
docker-compose restart backend

# Restart all services
docker-compose restart

# Check status
docker-compose ps

# View resource usage
docker stats
```

### Rebuilding

```bash
# Rebuild all services
docker-compose up -d --build

# Rebuild specific service
docker-compose up -d --build backend
docker-compose up -d --build frontend

# Force rebuild (no cache)
docker-compose build --no-cache
docker-compose up -d
```

### Database Operations

```bash
# Access PostgreSQL CLI
docker-compose exec postgres psql -U leapmailr -d leapmailr

# Backup database
docker-compose exec postgres pg_dump -U leapmailr leapmailr > backup-$(date +%Y%m%d).sql

# Restore database
docker-compose exec -T postgres psql -U leapmailr leapmailr < backup.sql

# View database size
docker-compose exec postgres psql -U leapmailr -d leapmailr -c "\l+"
```

## Development Workflow

### Hot Reload Development

For active development with immediate code changes:

```bash
# Option 1: Run database in Docker, services locally
docker-compose up -d postgres

# Terminal 1: Backend (from leapmailr directory)
cd ..
go run main.go

# Terminal 2: Frontend (from leapmailr-ui directory)
cd ../../leapmailr-ui
npm run dev
```

### Making Code Changes

```bash
# 1. Edit your code in leapmailr/ or leapmailr-ui/

# 2. Rebuild the specific service
docker-compose up -d --build backend
# or
docker-compose up -d --build frontend

# 3. Watch logs for errors
docker-compose logs -f backend
```

### Running Tests

```bash
# Backend tests
docker-compose exec backend go test ./...

# Frontend tests
docker-compose exec frontend npm test

# Or run locally
cd .. && go test ./...
cd ../../leapmailr-ui && npm test
```

## Production Deployment

### Pre-Production Checklist

- [ ] Update `JWT_SECRET` to strong random value (min 32 characters)
- [ ] Change `DB_PASSWORD` to secure password
- [ ] Configure production email provider credentials
- [ ] Update `CORS_ORIGINS` to your production domain
- [ ] Set `GIN_MODE=release`
- [ ] Enable HTTPS/TLS
- [ ] Configure backup strategy
- [ ] Set up monitoring and alerting
- [ ] Configure log aggregation
- [ ] Set resource limits

### Production docker-compose.yml

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  postgres:
    restart: always
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G

  backend:
    restart: always
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
      replicas: 2
    environment:
      GIN_MODE: release

  frontend:
    restart: always
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
      replicas: 2
```

Run with:
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Building for Registry

```bash
# Build and tag
docker build -t your-registry.com/leapmailr-backend:1.0.0 ..
docker build -t your-registry.com/leapmailr-frontend:1.0.0 ../../leapmailr-ui

# Push to registry
docker push your-registry.com/leapmailr-backend:1.0.0
docker push your-registry.com/leapmailr-frontend:1.0.0
```

## Troubleshooting

### Common Issues

#### 1. Port Already in Use

**Error**: `Bind for 0.0.0.0:8080 failed: port is already allocated`

**Solution**:
```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/Mac
lsof -ti:8080 | xargs kill -9
```

#### 2. Database Connection Failed

**Error**: `connection refused` or `database not ready`

**Solution**:
```bash
# Check database health
docker-compose exec postgres pg_isready -U leapmailr

# Restart database
docker-compose restart postgres

# Check logs
docker-compose logs postgres
```

#### 3. Frontend Can't Connect to Backend

**Error**: Network errors or API calls failing

**Solution**:
```bash
# Check backend is running
curl http://localhost:8080/api/v1/health

# Verify CORS settings in backend
docker-compose exec backend env | grep CORS

# Check frontend environment
docker-compose exec frontend env | grep NEXT_PUBLIC_API_URL
```

#### 4. Out of Disk Space

**Error**: `no space left on device`

**Solution**:
```bash
# Remove unused images
docker image prune -a

# Remove unused volumes
docker volume prune

# Remove all stopped containers
docker container prune

# Nuclear option - remove everything
docker system prune -a --volumes
```

#### 5. Build Fails

**Error**: Various build errors

**Solution**:
```bash
# Clear build cache
docker builder prune

# Rebuild without cache
docker-compose build --no-cache

# Check Dockerfile syntax
docker-compose config
```

#### 6. Container Keeps Restarting

**Check logs first**:
```bash
docker-compose logs backend
```

**Common causes**:
- Missing environment variables
- Database not ready (increase health check timeout)
- Port conflicts
- Out of memory

**Solution**:
```bash
# Check container status
docker-compose ps

# Inspect specific container
docker inspect leapmailr-backend

# Check resource usage
docker stats
```

### Health Checks

```bash
# Backend health
curl http://localhost:8080/api/v1/health

# Database health
docker-compose exec postgres pg_isready -U leapmailr

# Frontend health (if page loads)
curl http://localhost:3000

# All containers status
docker-compose ps
```

### Viewing Container Details

```bash
# Container information
docker inspect leapmailr-backend

# Environment variables
docker-compose exec backend env

# Network information
docker network inspect leapmailr_docker-compose_leapmailr-network

# Volume information
docker volume inspect leapmailr_docker-compose_postgres_data
```

### Performance Monitoring

```bash
# Real-time stats
docker stats

# Specific container
docker stats leapmailr-backend

# Get container processes
docker-compose top backend
```

## Advanced Topics

### Custom Network Configuration

```yaml
networks:
  leapmailr-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.28.0.0/16
```

### Volume Backups

```bash
# Backup volume
docker run --rm -v leapmailr_docker-compose_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres-backup.tar.gz /data

# Restore volume
docker run --rm -v leapmailr_docker-compose_postgres_data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres-backup.tar.gz -C /
```

### Multi-Stage Builds Explanation

Our Dockerfiles use multi-stage builds for smaller images:

1. **Builder Stage**: Compiles code with all build dependencies
2. **Runtime Stage**: Only includes compiled binary and runtime dependencies
3. **Result**: Smaller, more secure production images

### Environment-Specific Configurations

```bash
# Development
docker-compose up -d

# Staging
docker-compose -f docker-compose.yml -f docker-compose.staging.yml up -d

# Production
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## Security Best Practices

1. **Never commit `.env` files** to version control
2. **Use secrets management** in production (Docker Swarm secrets, Kubernetes secrets)
3. **Regularly update** base images
4. **Scan images** for vulnerabilities: `docker scan leapmailr-backend`
5. **Use non-root users** in containers (already configured)
6. **Limit resources** to prevent DoS
7. **Enable TLS** for all external communications
8. **Rotate credentials** regularly

## Using Makefile (Optional)

If you have `make` installed, you can use convenient shortcuts:

```bash
# Show all available commands
make help

# Setup environment
make setup

# Start services
make up

# View logs
make logs

# Rebuild backend
make rebuild-backend

# Database backup
make db-backup

# Check health
make health
```

## Support & Resources

- **Docker Documentation**: https://docs.docker.com
- **Docker Compose Reference**: https://docs.docker.com/compose/compose-file/
- **PostgreSQL Docker Hub**: https://hub.docker.com/_/postgres
- **Alpine Linux**: https://alpinelinux.org/

---

For additional help, check the main [README.md](../README.md) or open an issue on GitHub.
