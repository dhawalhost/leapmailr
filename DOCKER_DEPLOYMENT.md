# Docker Deployment Guide

## Overview
This guide covers deploying LeapMailr using Docker containers with automated CI/CD pipelines.

## Prerequisites
- Docker 20.10 or later
- Docker Compose 2.0 or later
- GitHub account with access to GitHub Container Registry
- (Optional) Docker Hub account

## Quick Start

### 1. Clone the Repositories
```bash
git clone https://github.com/yourusername/leapmailr.git
git clone https://github.com/yourusername/leapmailr-ui.git
```

### 2. Configure Environment Variables

**Backend (.env)**:
```bash
DB_USER=leapmailr
DB_PASSWORD=your_secure_password
DB_NAME=leapmailr
JWT_SECRET=your_jwt_secret_key_here
PORT=8080
GIN_MODE=release
```

**Frontend (.env.local)**:
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NODE_ENV=production
```

### 3. Build and Run with Docker Compose

**Backend only**:
```bash
cd leapmailr
docker-compose up -d
```

**Frontend only**:
```bash
cd leapmailr-ui
# First create the network if it doesn't exist
docker network create leapmailr-network
docker-compose up -d
```

### 4. Access the Application
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- API Health: http://localhost:8080/api/v1/health

## CI/CD Pipeline

### GitHub Actions Workflows

Both repositories include automated CI/CD pipelines that:

1. **Test** - Run linting and tests
2. **Build** - Build Docker images for multiple platforms (amd64, arm64)
3. **Push** - Push images to GitHub Container Registry and Docker Hub
4. **Deploy** - Automated deployment to staging/production

### Required GitHub Secrets

Add these secrets to your GitHub repository settings:

**Backend Repository**:
```
DOCKERHUB_USERNAME       - Docker Hub username
DOCKERHUB_TOKEN          - Docker Hub access token
STAGING_HOST             - Staging server hostname/IP
STAGING_USER             - SSH username for staging
STAGING_SSH_KEY          - SSH private key for staging
PROD_HOST                - Production server hostname/IP
PROD_USER                - SSH username for production
PROD_SSH_KEY             - SSH private key for production
```

**Frontend Repository**:
```
DOCKERHUB_USERNAME       - Docker Hub username
DOCKERHUB_TOKEN          - Docker Hub access token
NEXT_PUBLIC_API_URL      - Backend API URL
STAGING_FRONTEND_HOST    - Staging server hostname/IP
STAGING_FRONTEND_USER    - SSH username for staging
STAGING_SSH_KEY          - SSH private key for staging
PROD_FRONTEND_HOST       - Production server hostname/IP
PROD_FRONTEND_USER       - SSH username for production
PROD_SSH_KEY             - SSH private key for production
```

### Deployment Flow

**Development Branch (`develop`)**:
- Push to `develop` → Build → Push to registry → Deploy to staging

**Main Branch (`main`)**:
- Push to `main` → Build → Push to registry → Deploy to production

**Tags (`v*`)**:
- Create tag `v1.0.0` → Build → Push with version tags → Create GitHub release

## Manual Deployment

### Using Docker Images from Registry

**Backend**:
```bash
# Pull image
docker pull ghcr.io/yourusername/leapmailr:latest

# Run container
docker run -d \
  --name leapmailr-backend \
  --restart unless-stopped \
  -p 8080:8080 \
  --env-file .env \
  ghcr.io/yourusername/leapmailr:latest
```

**Frontend**:
```bash
# Pull image
docker pull ghcr.io/yourusername/leapmailr-ui:latest

# Run container
docker run -d \
  --name leapmailr-frontend \
  --restart unless-stopped \
  -p 3000:3000 \
  --env-file .env.local \
  ghcr.io/yourusername/leapmailr-ui:latest
```

### Building Images Locally

**Backend**:
```bash
cd leapmailr
docker build -t leapmailr-backend:local .
docker run -d -p 8080:8080 --env-file .env leapmailr-backend:local
```

**Frontend**:
```bash
cd leapmailr-ui
docker build --build-arg NEXT_PUBLIC_API_URL=http://localhost:8080 -t leapmailr-frontend:local .
docker run -d -p 3000:3000 leapmailr-frontend:local
```

## Server Setup

### 1. Install Docker on Server

```bash
# Update system
sudo apt-get update

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER

# Start Docker
sudo systemctl enable docker
sudo systemctl start docker
```

### 2. Create Network

```bash
docker network create leapmailr-network
```

### 3. Setup Directories

```bash
# Backend
sudo mkdir -p /opt/leapmailr
sudo chown $USER:$USER /opt/leapmailr
cd /opt/leapmailr

# Create .env file
cat > .env <<EOF
DB_HOST=postgres
DB_PORT=5432
DB_USER=leapmailr
DB_PASSWORD=your_secure_password
DB_NAME=leapmailr
PORT=8080
GIN_MODE=release
JWT_SECRET=your_jwt_secret
EOF

# Frontend
sudo mkdir -p /opt/leapmailr-ui
sudo chown $USER:$USER /opt/leapmailr-ui
cd /opt/leapmailr-ui

# Create .env.local file
cat > .env.local <<EOF
NEXT_PUBLIC_API_URL=http://backend:8080
NODE_ENV=production
EOF
```

### 4. Deploy with Docker Compose

Create `/opt/docker-compose.yml`:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: leapmailr-postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: leapmailr
      POSTGRES_PASSWORD: your_secure_password
      POSTGRES_DB: leapmailr
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - leapmailr-network

  backend:
    image: ghcr.io/yourusername/leapmailr:latest
    container_name: leapmailr-backend
    restart: unless-stopped
    ports:
      - "8080:8080"
    env_file:
      - /opt/leapmailr/.env
    depends_on:
      - postgres
    networks:
      - leapmailr-network

  frontend:
    image: ghcr.io/yourusername/leapmailr-ui:latest
    container_name: leapmailr-frontend
    restart: unless-stopped
    ports:
      - "3000:3000"
    env_file:
      - /opt/leapmailr-ui/.env.local
    depends_on:
      - backend
    networks:
      - leapmailr-network

volumes:
  postgres_data:

networks:
  leapmailr-network:
    driver: bridge
```

Start services:
```bash
cd /opt
docker-compose up -d
```

## Production Best Practices

### 1. Use Docker Secrets for Sensitive Data

```bash
# Create secrets
echo "your_db_password" | docker secret create db_password -
echo "your_jwt_secret" | docker secret create jwt_secret -

# Update docker-compose.yml to use secrets
```

### 2. Setup Nginx Reverse Proxy

Create `/etc/nginx/sites-available/leapmailr`:
```nginx
# Backend API
server {
    listen 80;
    server_name api.leapmailr.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Frontend
server {
    listen 80;
    server_name leapmailr.com www.leapmailr.com;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

Enable site and restart Nginx:
```bash
sudo ln -s /etc/nginx/sites-available/leapmailr /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 3. Setup SSL with Let's Encrypt

```bash
sudo apt-get install certbot python3-certbot-nginx
sudo certbot --nginx -d api.leapmailr.com -d leapmailr.com -d www.leapmailr.com
```

### 4. Setup Monitoring

**Docker container stats**:
```bash
docker stats
```

**View logs**:
```bash
# Backend
docker logs -f leapmailr-backend

# Frontend
docker logs -f leapmailr-frontend

# All services
docker-compose logs -f
```

### 5. Backup Database

```bash
# Backup
docker exec leapmailr-postgres pg_dump -U leapmailr leapmailr > backup.sql

# Restore
docker exec -i leapmailr-postgres psql -U leapmailr leapmailr < backup.sql
```

## Useful Commands

### Docker Management
```bash
# View running containers
docker ps

# Stop containers
docker stop leapmailr-backend leapmailr-frontend

# Remove containers
docker rm leapmailr-backend leapmailr-frontend

# Pull latest images
docker pull ghcr.io/yourusername/leapmailr:latest
docker pull ghcr.io/yourusername/leapmailr-ui:latest

# Clean up old images
docker image prune -f

# Clean up everything
docker system prune -a
```

### Docker Compose
```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Restart service
docker-compose restart backend

# Rebuild and restart
docker-compose up -d --build
```

## Troubleshooting

### Container won't start
```bash
# Check logs
docker logs leapmailr-backend

# Check if port is already in use
sudo lsof -i :8080

# Inspect container
docker inspect leapmailr-backend
```

### Database connection issues
```bash
# Check if postgres is running
docker ps | grep postgres

# Test database connection
docker exec -it leapmailr-postgres psql -U leapmailr -d leapmailr
```

### Image pull issues
```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Pull with authentication
docker pull ghcr.io/yourusername/leapmailr:latest
```

## Support

For issues or questions:
- GitHub Issues: https://github.com/yourusername/leapmailr/issues
- Documentation: https://docs.leapmailr.com
