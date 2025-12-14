# CI/CD Pipeline Setup Complete ✅

## What's Been Created

### Backend (leapmailr)
- ✅ `.github/workflows/ci-cd.yml` - GitHub Actions workflow for Docker build & deploy
- ✅ `Dockerfile` - Multi-stage Docker build for Go application
- ✅ `.dockerignore` - Optimize Docker build context
- ✅ `docker-compose.yml` - Local development with PostgreSQL
- ✅ `DOCKER_DEPLOYMENT.md` - Comprehensive deployment documentation

### Frontend (leapmailr-ui)
- ✅ `.github/workflows/ci-cd.yml` - GitHub Actions workflow for Docker build & deploy
- ✅ `Dockerfile` - Multi-stage Docker build for Next.js
- ✅ `.dockerignore` - Optimize Docker build context
- ✅ `docker-compose.yml` - Local development setup
- ✅ `next.config.ts` - Updated with standalone output mode

## CI/CD Pipeline Features

### Automated Workflows
1. **Test** - Linting, type checking, and tests
2. **Build** - Multi-platform Docker images (amd64, arm64)
3. **Push** - Automatic push to GitHub Container Registry & Docker Hub
4. **Deploy** - Automated deployment to staging (develop) and production (main)
5. **Release** - Tagged releases with Docker images

### Deployment Environments
- **develop branch** → Staging environment
- **main branch** → Production environment
- **tags (v*)** → GitHub releases with Docker images

## Required GitHub Secrets

Add these to your repository settings (Settings → Secrets and variables → Actions):

### Backend Repository
```
DOCKERHUB_USERNAME       Docker Hub username
DOCKERHUB_TOKEN          Docker Hub access token
STAGING_HOST             Staging server IP/hostname
STAGING_USER             SSH username for staging
STAGING_SSH_KEY          SSH private key (contents)
PROD_HOST                Production server IP/hostname
PROD_USER                SSH username for production
PROD_SSH_KEY             SSH private key (contents)
```

### Frontend Repository
```
DOCKERHUB_USERNAME       Docker Hub username
DOCKERHUB_TOKEN          Docker Hub access token
NEXT_PUBLIC_API_URL      Backend API URL
STAGING_FRONTEND_HOST    Staging server IP/hostname
STAGING_FRONTEND_USER    SSH username
STAGING_SSH_KEY          SSH private key
PROD_FRONTEND_HOST       Production server IP/hostname
PROD_FRONTEND_USER       SSH username
PROD_SSH_KEY             SSH private key
```

## Quick Start

### Local Development
```bash
# Backend
cd leapmailr
docker-compose up -d

# Frontend
cd leapmailr-ui
docker network create leapmailr-network  # First time only
docker-compose up -d
```

### Build Locally
```bash
# Backend
cd leapmailr
docker build -t leapmailr .
docker run -p 8080:8080 --env-file .env leapmailr

# Frontend
cd leapmailr-ui
docker build --build-arg NEXT_PUBLIC_API_URL=http://localhost:8080 -t leapmailr-ui .
docker run -p 3000:3000 leapmailr-ui
```

### Trigger Deployment
```bash
# Deploy to staging
git checkout develop
git push origin develop

# Deploy to production
git checkout main
git push origin main

# Create release
git tag v1.0.0
git push origin v1.0.0
```

## Server Setup

### 1. Install Docker
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

### 2. Create Docker Network
```bash
docker network create leapmailr-network
```

### 3. Setup Environment Files
```bash
# Backend: /opt/leapmailr/.env
DB_HOST=postgres
DB_USER=leapmailr
DB_PASSWORD=your_password
DB_NAME=leapmailr
JWT_SECRET=your_secret

# Frontend: /opt/leapmailr-ui/.env.local
NEXT_PUBLIC_API_URL=http://backend:8080
```

### 4. SSH Access
Ensure the deployment user can:
- SSH into the server
- Run docker commands without sudo
- Access `/opt/leapmailr` and `/opt/leapmailr-ui` directories

## Docker Images

Images are automatically built and pushed to:
- GitHub Container Registry: `ghcr.io/yourusername/repository-name`
- Docker Hub: `yourusername/leapmailr` and `yourusername/leapmailr-ui`

### Available Tags
- `latest` - Latest from main branch
- `develop` - Latest from develop branch
- `main-{sha}` - Specific commit from main
- `develop-{sha}` - Specific commit from develop
- `v1.0.0` - Versioned releases

## Monitoring & Maintenance

### View Logs
```bash
docker logs -f leapmailr
docker logs -f leapmailr-ui
```

### Update Containers
```bash
# Pull latest images
docker pull ghcr.io/yourusername/leapmailr:latest
docker pull ghcr.io/yourusername/leapmailr-ui:latest

# Recreate containers
docker stop leapmailr leapmailr-ui
docker rm leapmailr leapmailr-ui
docker run -d [options] ghcr.io/yourusername/leapmailr:latest
docker run -d [options] ghcr.io/yourusername/leapmailr-ui:latest
```

### Backup Database
```bash
docker exec leapmailr-postgres pg_dump -U leapmailr leapmailr > backup.sql
```

## Next Steps

1. **Configure GitHub Secrets** - Add all required secrets to repositories
2. **Setup Servers** - Install Docker and configure deployment servers
3. **Test Pipeline** - Push to develop branch to test staging deployment
4. **Configure Nginx** - Setup reverse proxy for domains (optional)
5. **Setup SSL** - Configure Let's Encrypt for HTTPS (optional)
6. **Monitor** - Setup monitoring and alerting

## Resources

- Backend Workflow: `.github/workflows/ci-cd.yml`
- Frontend Workflow: `.github/workflows/ci-cd.yml`
- Deployment Guide: `DOCKER_DEPLOYMENT.md`
- Docker Compose: `docker-compose.yml`

## Support

For issues:
- Check workflow logs in GitHub Actions tab
- View container logs: `docker logs container-name`
- Test locally first: `docker-compose up`
