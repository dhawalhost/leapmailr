# LeapMailR - Complete Setup Guide

**Version:** 2.0  
**Last Updated:** November 4, 2025  
**Author:** Infrastructure Team

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local Development Setup](#local-development-setup)
3. [Environment Configuration](#environment-configuration)
4. [Database Setup](#database-setup)
5. [Redis Setup](#redis-setup)
6. [Running the Application](#running-the-application)
7. [Frontend Setup](#frontend-setup)
8. [Testing](#testing)
9. [Docker Development](#docker-development)
10. [Production Deployment](#production-deployment)
11. [CI/CD Pipeline Setup](#cicd-pipeline-setup)
12. [Monitoring & Observability](#monitoring--observability)
13. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software

```bash
# Go 1.23+
go version

# Node.js 18+ and npm
node --version
npm --version

# PostgreSQL 14+
psql --version

# Redis 6+
redis-cli --version

# Docker & Docker Compose (optional)
docker --version
docker-compose --version

# Git
git --version
```

### Installing Prerequisites

**Windows:**
```powershell
# Using Chocolatey
choco install golang nodejs postgresql redis docker-desktop git

# Or using winget
winget install GoLang.Go
winget install OpenJS.NodeJS
winget install PostgreSQL.PostgreSQL
winget install Redis.Redis
winget install Docker.DockerDesktop
```

**macOS:**
```bash
# Using Homebrew
brew install go node postgresql redis docker git
```

**Linux (Ubuntu/Debian):**
```bash
# Go
wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# PostgreSQL
sudo apt-get update
sudo apt-get install -y postgresql postgresql-contrib

# Redis
sudo apt-get install -y redis-server

# Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

---

## Local Development Setup

### 1. Clone Repository

```bash
# Clone backend
git clone https://github.com/dhawalhost/leapmailr.git
cd leapmailr

# Clone frontend (in separate terminal/directory)
git clone https://github.com/dhawalhost/leapmailr-ui.git
cd leapmailr-ui
```

### 2. Backend Setup

```bash
cd leapmailr

# Install Go dependencies
go mod download
go mod tidy

# Verify installation
go build .
```

### 3. Generate Secrets

```bash
# Create secrets directory
mkdir -p secrets backups logs

# Generate encryption key (32 bytes base64)
ENCRYPTION_KEY=$(openssl rand -base64 32)
echo "ENCRYPTION_KEY=$ENCRYPTION_KEY"

# Generate JWT secret (64 characters)
JWT_SECRET=$(openssl rand -base64 64 | tr -d "=+/" | cut -c1-64)
echo "JWT_SECRET=$JWT_SECRET"

# Generate session secret
SESSION_SECRET=$(openssl rand -base64 32)
echo "SESSION_SECRET=$SESSION_SECRET"
```

---

## Environment Configuration

### Backend Configuration

Create `config.env` in the backend root:

```bash
# config.env

# ============================================
# SERVER CONFIGURATION
# ============================================
PORT=8080
ENVIRONMENT=development
FRONTEND_URL=http://localhost:3000

# ============================================
# DATABASE CONFIGURATION
# ============================================
DB_HOST=localhost
DB_PORT=5432
DB_USER=leapmailr
DB_PASSWORD=your_secure_password_here
DB_NAME=leapmailr
DB_SSLMODE=disable

# ============================================
# REDIS CONFIGURATION
# ============================================
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# ============================================
# SECURITY SECRETS
# ============================================
JWT_SECRET=your_jwt_secret_from_above
ENCRYPTION_KEY=your_encryption_key_from_above
SESSION_SECRET=your_session_secret_from_above

# ============================================
# CORS CONFIGURATION
# ============================================
ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000

# ============================================
# RATE LIMITING
# ============================================
RATE_LIMIT_ENABLED=true
RATE_LIMIT_GLOBAL=100
RATE_LIMIT_AUTH=10
RATE_LIMIT_API=50

# ============================================
# EMAIL CONFIGURATION (SendGrid example)
# ============================================
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=your_sendgrid_api_key
FROM_EMAIL=noreply@leapmailr.com
FROM_NAME=LeapMailR

# ============================================
# LOGGING
# ============================================
LOG_LEVEL=info
LOG_FORMAT=json

# ============================================
# SECURITY HEADERS
# ============================================
HSTS_MAX_AGE=31536000
FORCE_HTTPS=false

# ============================================
# MFA CONFIGURATION
# ============================================
MFA_ISSUER=LeapMailR
MFA_BACKUP_CODES_COUNT=10

# ============================================
# SECRETS MANAGEMENT
# ============================================
SECRETS_PROVIDER=local
SECRETS_DIR=./secrets

# ============================================
# BACKUP CONFIGURATION
# ============================================
BACKUP_DIR=./backups
BACKUP_RETENTION_DAYS=30

# ============================================
# MONITORING
# ============================================
METRICS_ENABLED=true
METRICS_PORT=9090
```

### Frontend Configuration

Create `.env.local` in `leapmailr-ui/`:

```bash
# .env.local

# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080

# Environment
NEXT_PUBLIC_ENVIRONMENT=development

# Feature Flags
NEXT_PUBLIC_ENABLE_MFA=true
NEXT_PUBLIC_ENABLE_ANALYTICS=true
```

### File Permissions

```bash
# Backend
chmod 600 config.env
chmod 700 secrets/
chmod 700 backups/
chmod 700 logs/

# Make scripts executable
chmod +x scripts/*.sh
```

---

## Database Setup

### 1. Start PostgreSQL

```bash
# Windows (if installed as service)
# PostgreSQL should already be running

# macOS
brew services start postgresql

# Linux
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### 2. Create Database and User

```bash
# Connect to PostgreSQL
sudo -u postgres psql

# Or on Windows
psql -U postgres
```

```sql
-- Create user
CREATE USER leapmailr WITH PASSWORD 'your_secure_password_here';

-- Create database
CREATE DATABASE leapmailr OWNER leapmailr;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE leapmailr TO leapmailr;

-- Exit
\q
```

### 3. Initialize Database Schema

```bash
# Connect to the database
psql -h localhost -U leapmailr -d leapmailr

# Or run SQL file if you have one
psql -h localhost -U leapmailr -d leapmailr -f database/schema.sql
```

### 4. Verify Database Connection

```bash
# Test connection
psql -h localhost -U leapmailr -d leapmailr -c "SELECT version();"

# List tables
psql -h localhost -U leapmailr -d leapmailr -c "\dt"
```

---

## Redis Setup

### 1. Start Redis

```bash
# Windows (if installed)
redis-server

# macOS
brew services start redis

# Linux
sudo systemctl start redis
sudo systemctl enable redis
```

### 2. Verify Redis

```bash
# Test connection
redis-cli ping
# Should return: PONG

# Test with password (if configured)
redis-cli -a your_redis_password ping
```

### 3. Configure Redis (Optional)

Edit Redis configuration for production:

```bash
# Linux/macOS
sudo nano /etc/redis/redis.conf

# Set password
requirepass your_secure_redis_password

# Restart Redis
sudo systemctl restart redis
```

---

## Running the Application

### Backend (Go API)

```bash
cd leapmailr

# Method 1: Using go run (development)
go run .

# Method 2: Build and run (recommended)
go build -o leapmailr.exe .
./leapmailr.exe

# Method 3: Using air for hot reload (install first)
go install github.com/cosmtrek/air@latest
air

# The API will be available at http://localhost:8080
```

### Frontend (Next.js)

```bash
cd leapmailr-ui

# Install dependencies
npm install

# Run development server
npm run dev

# The UI will be available at http://localhost:3000
```

### Verify Everything Works

```bash
# Check backend health
curl http://localhost:8080/health

# Check backend metrics
curl http://localhost:8080/metrics

# Check frontend
open http://localhost:3000
```

---

## Frontend Setup

### Complete Frontend Installation

```bash
cd leapmailr-ui

# Install dependencies
npm install

# Install additional dev dependencies
npm install -D @types/node @types/react typescript eslint

# Build for production (test)
npm run build

# Run production build locally
npm run start
```

### Frontend Scripts

```json
// package.json scripts
{
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start",
    "lint": "next lint",
    "type-check": "tsc --noEmit"
  }
}
```

---

## Testing

### Backend Tests

```bash
cd leapmailr

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./handlers/...
go test ./service/...
go test ./middleware/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Frontend Tests

```bash
cd leapmailr-ui

# Install testing dependencies
npm install -D @testing-library/react @testing-library/jest-dom jest

# Run tests
npm test

# Run tests with coverage
npm test -- --coverage
```

### Integration Tests

```bash
# Backend integration tests
cd leapmailr
go test -tags=integration ./...

# End-to-end tests (if configured)
cd leapmailr-ui
npm run test:e2e
```

---

## Docker Development

### Using Docker Compose (Easiest)

```bash
cd leapmailr/docker-compose

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Rebuild and restart
docker-compose up -d --build
```

### Docker Compose Configuration

Create `docker-compose/docker-compose.dev.yml`:

```yaml
version: '3.8'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:14-alpine
    container_name: leapmailr-postgres
    environment:
      POSTGRES_DB: leapmailr
      POSTGRES_USER: leapmailr
      POSTGRES_PASSWORD: dev_password_123
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U leapmailr"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: leapmailr-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Backend API
  backend:
    build:
      context: ..
      dockerfile: Dockerfile
    container_name: leapmailr-backend
    ports:
      - "8080:8080"
      - "9090:9090"  # Metrics
    environment:
      - PORT=8080
      - ENVIRONMENT=development
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=leapmailr
      - DB_PASSWORD=dev_password_123
      - DB_NAME=leapmailr
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - FRONTEND_URL=http://localhost:3000
    env_file:
      - ../.env.docker
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ../logs:/app/logs
      - ../secrets:/app/secrets
      - ../backups:/app/backups

  # Frontend UI
  frontend:
    build:
      context: ../../leapmailr-ui
      dockerfile: Dockerfile.dev
    container_name: leapmailr-frontend
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8080
    volumes:
      - ../../leapmailr-ui:/app
      - /app/node_modules
      - /app/.next
    depends_on:
      - backend

volumes:
  postgres_data:
  redis_data:
```

### Docker Commands Reference

```bash
# Build images
docker-compose build

# Start services
docker-compose up -d

# View logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Execute commands in containers
docker-compose exec backend /bin/sh
docker-compose exec postgres psql -U leapmailr

# Stop services
docker-compose stop

# Remove everything
docker-compose down -v

# Restart specific service
docker-compose restart backend
```

---

## Production Deployment

### 1. Build for Production

**Backend:**
```bash
cd leapmailr

# Build optimized binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o leapmailr .

# Or for Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o leapmailr.exe .
```

**Frontend:**
```bash
cd leapmailr-ui

# Build production bundle
npm run build

# Test production build locally
npm run start
```

### 2. Environment Variables (Production)

Update `config.env` for production:

```bash
# Production config.env
ENVIRONMENT=production
PORT=8080
FRONTEND_URL=https://app.leapmailr.com

# Database
DB_HOST=your-rds-endpoint.amazonaws.com
DB_SSLMODE=require

# Force HTTPS
FORCE_HTTPS=true
HSTS_MAX_AGE=31536000

# Strict CORS
ALLOWED_ORIGINS=https://app.leapmailr.com,https://www.leapmailr.com

# Production secrets (use AWS Secrets Manager or Vault)
SECRETS_PROVIDER=aws
AWS_REGION=us-east-1
```

### 3. Kubernetes Deployment

Create `k8s/deployment.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: leapmailr

---
apiVersion: v1
kind: Secret
metadata:
  name: leapmailr-secrets
  namespace: leapmailr
type: Opaque
stringData:
  jwt-secret: "your-jwt-secret"
  encryption-key: "your-encryption-key"
  db-password: "your-db-password"

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: leapmailr-config
  namespace: leapmailr
data:
  config.env: |
    ENVIRONMENT=production
    PORT=8080
    DB_HOST=postgres-service
    DB_PORT=5432
    DB_USER=leapmailr
    DB_NAME=leapmailr
    REDIS_HOST=redis-service
    REDIS_PORT=6379

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: leapmailr
spec:
  replicas: 3
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: backend
        image: your-registry/leapmailr-backend:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: leapmailr-secrets
              key: jwt-secret
        - name: ENCRYPTION_KEY
          valueFrom:
            secretKeyRef:
              name: leapmailr-secrets
              key: encryption-key
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: leapmailr-secrets
              key: db-password
        envFrom:
        - configMapRef:
            name: leapmailr-config
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: backend-service
  namespace: leapmailr
spec:
  selector:
    app: backend
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: LoadBalancer

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: leapmailr
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: frontend
        image: your-registry/leapmailr-frontend:latest
        ports:
        - containerPort: 3000
        env:
        - name: NEXT_PUBLIC_API_URL
          value: "https://api.leapmailr.com"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"

---
apiVersion: v1
kind: Service
metadata:
  name: frontend-service
  namespace: leapmailr
spec:
  selector:
    app: frontend
  ports:
  - port: 80
    targetPort: 3000
  type: LoadBalancer
```

### 4. Deploy to Kubernetes

```bash
# Create namespace and secrets
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/secrets.yaml

# Deploy application
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# Check deployment
kubectl get pods -n leapmailr
kubectl get services -n leapmailr

# View logs
kubectl logs -f deployment/backend -n leapmailr
```

---

## CI/CD Pipeline Setup

### GitHub Actions

Create `.github/workflows/ci-cd.yml`:

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  GO_VERSION: '1.23'
  NODE_VERSION: '18'
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  # ==========================================
  # BACKEND: Test & Build
  # ==========================================
  backend-test:
    name: Backend Tests
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_DB: leapmailr_test
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
      
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Install dependencies
      run: go mod download

    - name: Run linter
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest

    - name: Run tests
      env:
        DB_HOST: localhost
        DB_PORT: 5432
        DB_USER: test
        DB_PASSWORD: test
        DB_NAME: leapmailr_test
        REDIS_HOST: localhost
        REDIS_PORT: 6379
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out

    - name: Build binary
      run: |
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o leapmailr .

    - name: Upload artifact
      uses: actions/upload-artifact@v3
      with:
        name: backend-binary
        path: leapmailr

  # ==========================================
  # FRONTEND: Test & Build
  # ==========================================
  frontend-test:
    name: Frontend Tests
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: ../leapmailr-ui

    steps:
    - name: Checkout backend
      uses: actions/checkout@v4
      with:
        path: leapmailr

    - name: Checkout frontend
      uses: actions/checkout@v4
      with:
        repository: dhawalhost/leapmailr-ui
        path: leapmailr-ui

    - name: Set up Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'
        cache-dependency-path: leapmailr-ui/package-lock.json

    - name: Install dependencies
      run: npm ci

    - name: Run linter
      run: npm run lint

    - name: Run type check
      run: npm run type-check || true

    - name: Run tests
      run: npm test || true

    - name: Build application
      run: npm run build
      env:
        NEXT_PUBLIC_API_URL: https://api.leapmailr.com

    - name: Upload build
      uses: actions/upload-artifact@v3
      with:
        name: frontend-build
        path: leapmailr-ui/.next

  # ==========================================
  # SECURITY SCANNING
  # ==========================================
  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    needs: [backend-test]

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: Upload Trivy results to GitHub Security
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: 'trivy-results.sarif'

    - name: Run Gosec Security Scanner
      uses: securego/gosec@master
      with:
        args: '-no-fail -fmt sarif -out gosec-results.sarif ./...'

  # ==========================================
  # BUILD DOCKER IMAGES
  # ==========================================
  build-backend-image:
    name: Build Backend Docker Image
    runs-on: ubuntu-latest
    needs: [backend-test, security-scan]
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}-backend
        tags: |
          type=ref,event=branch
          type=sha,prefix={{branch}}-
          type=semver,pattern={{version}}

    - name: Build and push
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./Dockerfile
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  build-frontend-image:
    name: Build Frontend Docker Image
    runs-on: ubuntu-latest
    needs: [frontend-test]
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'

    steps:
    - name: Checkout frontend
      uses: actions/checkout@v4
      with:
        repository: dhawalhost/leapmailr-ui
        path: leapmailr-ui

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}-frontend
        tags: |
          type=ref,event=branch
          type=sha,prefix={{branch}}-
          type=semver,pattern={{version}}

    - name: Build and push
      uses: docker/build-push-action@v5
      with:
        context: ./leapmailr-ui
        file: ./leapmailr-ui/Dockerfile
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  # ==========================================
  # DEPLOY TO STAGING
  # ==========================================
  deploy-staging:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    needs: [build-backend-image, build-frontend-image]
    environment: staging

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up kubectl
      uses: azure/setup-kubectl@v3

    - name: Configure kubectl
      run: |
        echo "${{ secrets.KUBE_CONFIG_STAGING }}" | base64 -d > kubeconfig
        export KUBECONFIG=kubeconfig

    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/backend \
          backend=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}-backend:main-${{ github.sha }} \
          -n leapmailr-staging
        
        kubectl set image deployment/frontend \
          frontend=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}-frontend:main-${{ github.sha }} \
          -n leapmailr-staging

    - name: Wait for rollout
      run: |
        kubectl rollout status deployment/backend -n leapmailr-staging
        kubectl rollout status deployment/frontend -n leapmailr-staging

    - name: Run smoke tests
      run: |
        curl -f https://staging-api.leapmailr.com/health || exit 1
        curl -f https://staging.leapmailr.com || exit 1

  # ==========================================
  # DEPLOY TO PRODUCTION
  # ==========================================
  deploy-production:
    name: Deploy to Production
    runs-on: ubuntu-latest
    needs: [deploy-staging]
    environment: production
    if: github.ref == 'refs/heads/main'

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up kubectl
      uses: azure/setup-kubectl@v3

    - name: Configure kubectl
      run: |
        echo "${{ secrets.KUBE_CONFIG_PRODUCTION }}" | base64 -d > kubeconfig
        export KUBECONFIG=kubeconfig

    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/backend \
          backend=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}-backend:main-${{ github.sha }} \
          -n leapmailr
        
        kubectl set image deployment/frontend \
          frontend=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}-frontend:main-${{ github.sha }} \
          -n leapmailr

    - name: Wait for rollout
      run: |
        kubectl rollout status deployment/backend -n leapmailr --timeout=10m
        kubectl rollout status deployment/frontend -n leapmailr --timeout=10m

    - name: Run smoke tests
      run: |
        curl -f https://api.leapmailr.com/health || exit 1
        curl -f https://app.leapmailr.com || exit 1

    - name: Notify deployment
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        text: 'Deployment to production completed!'
        webhook_url: ${{ secrets.SLACK_WEBHOOK }}
      if: always()
```

### GitLab CI/CD

Create `.gitlab-ci.yml`:

```yaml
stages:
  - test
  - build
  - deploy

variables:
  GO_VERSION: "1.23"
  DOCKER_DRIVER: overlay2
  DOCKER_TLS_CERTDIR: "/certs"

# ==========================================
# BACKEND TESTS
# ==========================================
backend:test:
  stage: test
  image: golang:${GO_VERSION}
  
  services:
    - postgres:14
    - redis:7
  
  variables:
    POSTGRES_DB: leapmailr_test
    POSTGRES_USER: test
    POSTGRES_PASSWORD: test
    DB_HOST: postgres
    REDIS_HOST: redis
  
  before_script:
    - cd leapmailr
    - go mod download
  
  script:
    - go test -v -race -coverprofile=coverage.out ./...
    - go tool cover -func=coverage.out
  
  coverage: '/total:.*?(\d+\.\d+)%/'
  
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml

# ==========================================
# FRONTEND TESTS
# ==========================================
frontend:test:
  stage: test
  image: node:18
  
  before_script:
    - cd leapmailr-ui
    - npm ci
  
  script:
    - npm run lint
    - npm run type-check
    - npm run build
  
  cache:
    paths:
      - leapmailr-ui/node_modules/

# ==========================================
# BUILD BACKEND
# ==========================================
backend:build:
  stage: build
  image: docker:latest
  
  services:
    - docker:dind
  
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  
  script:
    - cd leapmailr
    - docker build -t $CI_REGISTRY_IMAGE/backend:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE/backend:$CI_COMMIT_SHA
  
  only:
    - main
    - develop

# ==========================================
# BUILD FRONTEND
# ==========================================
frontend:build:
  stage: build
  image: docker:latest
  
  services:
    - docker:dind
  
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  
  script:
    - cd leapmailr-ui
    - docker build -t $CI_REGISTRY_IMAGE/frontend:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE/frontend:$CI_COMMIT_SHA
  
  only:
    - main
    - develop

# ==========================================
# DEPLOY STAGING
# ==========================================
deploy:staging:
  stage: deploy
  image: bitnami/kubectl:latest
  
  environment:
    name: staging
    url: https://staging.leapmailr.com
  
  before_script:
    - echo "$KUBE_CONFIG_STAGING" | base64 -d > /tmp/kubeconfig
    - export KUBECONFIG=/tmp/kubeconfig
  
  script:
    - kubectl set image deployment/backend backend=$CI_REGISTRY_IMAGE/backend:$CI_COMMIT_SHA -n leapmailr-staging
    - kubectl set image deployment/frontend frontend=$CI_REGISTRY_IMAGE/frontend:$CI_COMMIT_SHA -n leapmailr-staging
    - kubectl rollout status deployment/backend -n leapmailr-staging
    - kubectl rollout status deployment/frontend -n leapmailr-staging
  
  only:
    - develop

# ==========================================
# DEPLOY PRODUCTION
# ==========================================
deploy:production:
  stage: deploy
  image: bitnami/kubectl:latest
  
  environment:
    name: production
    url: https://app.leapmailr.com
  
  before_script:
    - echo "$KUBE_CONFIG_PRODUCTION" | base64 -d > /tmp/kubeconfig
    - export KUBECONFIG=/tmp/kubeconfig
  
  script:
    - kubectl set image deployment/backend backend=$CI_REGISTRY_IMAGE/backend:$CI_COMMIT_SHA -n leapmailr
    - kubectl set image deployment/frontend frontend=$CI_REGISTRY_IMAGE/frontend:$CI_COMMIT_SHA -n leapmailr
    - kubectl rollout status deployment/backend -n leapmailr
    - kubectl rollout status deployment/frontend -n leapmailr
  
  when: manual
  only:
    - main
```

### Required CI/CD Secrets

Configure these secrets in your CI/CD platform:

**GitHub Actions:**
- `GITHUB_TOKEN` (automatic)
- `KUBE_CONFIG_STAGING`
- `KUBE_CONFIG_PRODUCTION`
- `SLACK_WEBHOOK`
- `DOCKER_USERNAME`
- `DOCKER_PASSWORD`

**GitLab CI:**
- `CI_REGISTRY_USER`
- `CI_REGISTRY_PASSWORD`
- `KUBE_CONFIG_STAGING`
- `KUBE_CONFIG_PRODUCTION`

---

## Monitoring & Observability

### Prometheus Configuration

Create `monitoring/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'leapmailr-backend'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
```

### Grafana Dashboard

Import dashboard from `monitoring/grafana-dashboard.json` or access pre-built dashboards.

### Log Aggregation

```yaml
# Filebeat configuration for ELK stack
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /app/logs/*.log
    json.keys_under_root: true
    json.add_error_key: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "leapmailr-%{+yyyy.MM.dd}"
```

---

## Troubleshooting

### Common Issues

#### Database Connection Failed

```bash
# Check PostgreSQL is running
sudo systemctl status postgresql

# Test connection
psql -h localhost -U leapmailr -d leapmailr -c "SELECT 1"

# Check config.env
cat config.env | grep DB_

# View logs
tail -f logs/application.log | grep database
```

#### Redis Connection Failed

```bash
# Check Redis is running
sudo systemctl status redis

# Test connection
redis-cli ping

# Check if password is required
redis-cli -a your_password ping
```

#### Frontend Can't Connect to Backend

```bash
# Check CORS settings
curl -H "Origin: http://localhost:3000" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS http://localhost:8080/api/health

# Verify backend is running
curl http://localhost:8080/health

# Check frontend env
cat leapmailr-ui/.env.local
```

#### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080
# Or on Windows
netstat -ano | findstr :8080

# Kill process
kill -9 <PID>
# Or on Windows
taskkill /PID <PID> /F
```

#### Build Fails

```bash
# Clean and rebuild
go clean -cache
go mod tidy
go build .

# Check Go version
go version

# Install missing dependencies
go mod download
```

### Getting Help

- **Documentation:** Check `docs/` directory
- **Logs:** Review `logs/application.log`
- **Health Check:** `curl http://localhost:8080/health`
- **Metrics:** `curl http://localhost:8080/metrics`

---

## Quick Start Checklist

- [ ] Install prerequisites (Go, Node.js, PostgreSQL, Redis)
- [ ] Clone repositories
- [ ] Generate secrets
- [ ] Configure `config.env`
- [ ] Create database and user
- [ ] Start PostgreSQL and Redis
- [ ] Run backend: `go run .`
- [ ] Run frontend: `cd ../leapmailr-ui && npm run dev`
- [ ] Verify: Visit `http://localhost:3000`
- [ ] Run tests: `go test ./...`
- [ ] Build for production
- [ ] Set up CI/CD pipeline
- [ ] Deploy to staging
- [ ] Deploy to production

---

## Additional Resources

- [Backend Repository](https://github.com/dhawalhost/leapmailr)
- [Frontend Repository](https://github.com/dhawalhost/leapmailr-ui)
- [API Documentation](docs/API.md)
- [Architecture](docs/ARCHITECTURE.md)
- [SOC 2 Compliance](docs/SOC2_COMPLIANCE.md)
- [Secrets Management](docs/SECRETS_MANAGEMENT.md)
- [Disaster Recovery](docs/DISASTER_RECOVERY.md)

---

**Last Updated:** November 4, 2025  
**Maintainers:** Infrastructure Team  
**Support:** support@leapmailr.com
