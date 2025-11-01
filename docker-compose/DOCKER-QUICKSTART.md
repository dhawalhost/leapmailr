# 🐳 Quick Start with Docker

Get LeapMailr running in 5 minutes with Docker!

## 1️⃣ Prerequisites

- **Docker Desktop** installed ([Download here](https://www.docker.com/products/docker-desktop))
- **4GB RAM** minimum
- **Ports free**: 3000, 8080, 5432

## 2️⃣ Quick Start (Automated)

### Windows
```cmd
cd docker-compose
verify-docker.bat
docker-setup.bat
```

### Linux/Mac
```bash
cd docker-compose
chmod +x verify-docker.sh docker-setup.sh
./verify-docker.sh
./docker-setup.sh
```

## 3️⃣ Manual Start

```bash
# 1. Navigate to docker-compose directory
cd docker-compose

# 2. Copy environment file
cp .env.example .env

# 3. Start all services
docker-compose up -d

# 4. Wait for services to be healthy (~30 seconds)
docker-compose ps

# 5. Access the application
# Frontend: http://localhost:3000
# Backend:  http://localhost:8080
```

## 4️⃣ Verify Everything Works

```bash
# Check backend health
curl http://localhost:8080/api/v1/health

# Should return: {"status":"healthy"}

# Check frontend
# Open browser to http://localhost:3000
```

## 5️⃣ Create Your First Account

1. Open http://localhost:3000
2. Click "Get Started" or "Register"
3. Fill in your details:
   - Email: your@email.com
   - Password: (minimum 8 characters)
   - First Name & Last Name
4. Click "Create Account"
5. Login with your credentials

## 6️⃣ Send Your First Email

1. Go to Dashboard > Templates
2. Click "New Template"
3. Create a simple template:
   ```html
   <h1>Hello {{name}}!</h1>
   <p>Welcome to LeapMailr</p>
   ```
4. Go to Dashboard > Send Email
5. Select your template
6. Add recipient email
7. Fill in variables
8. Click "Send Now"

## 📊 Common Commands

```bash
# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Restart services
docker-compose restart

# Rebuild after code changes
docker-compose up -d --build

# Check status
docker-compose ps
```

## 🔧 Configuration

### Email Service Setup

Email providers (SMTP, SendGrid, Mailgun, etc.) are configured through the Email Service API after authentication.

**No environment file changes needed for email configuration.**

After starting the application:
1. Register/login through the UI at http://localhost:3000
2. Navigate to Email Services settings
3. Add your email service configuration through the API
4. Set a default email service

For API details, see the main documentation or use the API endpoints at http://localhost:8080/api/v1/email-services

## ❌ Stop Everything

```bash
# Stop services (keep data)
docker-compose down

# Stop and remove all data
docker-compose down -v
```

## 🐛 Troubleshooting

### Port Already in Use
```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/Mac
sudo lsof -ti:8080 | xargs kill -9
```

### Services Not Starting
```bash
# Check logs
docker-compose logs backend

# Restart specific service
docker-compose restart backend
```

### Database Issues
```bash
# Reset database
docker-compose down -v
docker-compose up -d
```

## 📚 Next Steps

- Read [DOCKER.md](DOCKER.md) for detailed documentation
- Check [../README.md](../README.md) for feature overview
- See [../../leapmailr-ui/README.md](../../leapmailr-ui/README.md) for frontend details

## 🎯 Architecture

```
┌─────────────┐
│  Browser    │ :3000
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Frontend   │ Next.js Container
│  (Node 20)  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Backend    │ :8080 Go/Gin Container
│  (Alpine)   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Postgres   │ :5432 Database Container
│  (Alpine)   │
└─────────────┘
```

## 🚀 Using Makefile (Optional)

If you have `make` installed:

```bash
# Show all commands
make help

# Quick setup
make setup
make up

# View logs
make logs

# Database backup
make db-backup

# Health check
make health
```
