# Quick Start Guide - GitHub Container Registry Only

## ✅ What's Already Done
- Docker images build successfully locally
- Workflows configured to use GitHub Container Registry (ghcr.io)
- No Docker Hub account needed!

## 🎯 Next Steps

### Step 1: Enable GitHub Container Registry (Already enabled by default!)
GitHub Container Registry is automatically available for your repositories. No setup needed!

### Step 2: Test the CI/CD Pipeline

#### 2a. Create develop branch (if it doesn't exist)
```bash
# Backend
cd /c/Users/dhawa/go/src/leapmailr
git checkout -b develop
git push -u origin develop

# Frontend
cd /c/Users/dhawa/go/src/leapmailr-ui
git checkout -b develop
git push -u origin develop
```

#### 2b. Commit and push your Docker changes
```bash
# Backend
cd /c/Users/dhawa/go/src/leapmailr
git add .
git commit -m "feat: add Docker support and CI/CD pipeline"
git push origin main

# Frontend  
cd /c/Users/dhawa/go/src/leapmailr-ui
git add .
git commit -m "feat: add Docker support and CI/CD pipeline"
git push origin main
```

#### 2c. Watch the GitHub Actions
1. Go to https://github.com/dhawalhost/leapmailr/actions
2. Go to https://github.com/dhawalhost/leapmailr-ui/actions
3. You should see the workflows running!

### Step 3: Access Your Docker Images

After successful build, your images will be at:
- **Backend**: `ghcr.io/dhawalhost/leapmailr:latest`
- **Frontend**: `ghcr.io/dhawalhost/leapmailr-ui:latest`

### Step 4: Pull and Run Your Images

```bash
# Login to GitHub Container Registry (one time)
echo $GITHUB_TOKEN | docker login ghcr.io -u dhawalhost --password-stdin

# Pull images
docker pull ghcr.io/dhawalhost/leapmailr:latest
docker pull ghcr.io/dhawalhost/leapmailr-ui:latest

# Run backend
docker run -d \
  --name leapmailr \
  -p 8080:8080 \
  -e DB_HOST=postgres \
  -e DB_USER=leapmailr \
  -e DB_PASSWORD=your_password \
  -e DB_NAME=leapmailr \
  -e JWT_SECRET=your_secret \
  ghcr.io/dhawalhost/leapmailr:latest

# Run frontend
docker run -d \
  --name leapmailr-ui \
  -p 3000:3000 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080 \
  ghcr.io/dhawalhost/leapmailr-ui:latest
```

## 🚀 Optional: Server Deployment Setup

Only needed if you want automated deployment to your own servers.

### Required GitHub Secrets (per repository)

**For Backend Repository** (Settings → Secrets → Actions):
```
STAGING_HOST             - Your staging server IP/hostname
STAGING_USER             - SSH username
STAGING_SSH_KEY          - SSH private key (paste the entire key)
PROD_HOST                - Your production server IP/hostname  
PROD_USER                - SSH username
PROD_SSH_KEY             - SSH private key
```

**For Frontend Repository**:
```
NEXT_PUBLIC_API_URL      - Backend API URL (e.g., https://api.leapmailr.com)
STAGING_FRONTEND_HOST    - Staging server IP/hostname
STAGING_FRONTEND_USER    - SSH username
STAGING_SSH_KEY          - SSH private key
PROD_FRONTEND_HOST       - Production server IP/hostname
PROD_FRONTEND_USER       - SSH username
PROD_SSH_KEY             - SSH private key
```

### Server Setup
```bash
# Install Docker on your server
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Create network
docker network create leapmailr-network

# Create directories
sudo mkdir -p /opt/leapmailr /opt/leapmailr-ui
sudo chown $USER:$USER /opt/leapmailr /opt/leapmailr-ui
```

## 📊 Workflow Triggers

- **Push to `main`** → Build & Deploy to Production
- **Push to `develop`** → Build & Deploy to Staging  
- **Pull Request** → Build & Test only (no deployment)
- **Tag `v*`** → Build & Create GitHub Release

## 🎉 You're Done!

The pipeline will automatically:
1. ✅ Run tests and linting
2. ✅ Build Docker images for multiple platforms
3. ✅ Push to GitHub Container Registry
4. ✅ Deploy to servers (if configured)
5. ✅ Run health checks

No Docker Hub account needed - everything runs through GitHub!
