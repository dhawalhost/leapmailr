# LeapMailr - Full Stack Docker Compose

This directory contains the complete Docker Compose setup for running the entire LeapMailr stack (PostgreSQL, Backend API, and Frontend UI).

## 🚀 Quick Start

### Option 1: Automated Setup (Recommended)

**Windows:**
```cmd
cd docker-compose
verify-docker.bat
docker-setup.bat
```

**Linux/Mac:**
```bash
cd docker-compose
chmod +x verify-docker.sh docker-setup.sh
./verify-docker.sh
./docker-setup.sh
```

### Option 2: Manual Setup

```bash
# 1. Navigate to this directory
cd docker-compose

# 2. Copy environment file
cp .env.example .env

# 3. Edit .env with your settings (especially email provider)
nano .env  # or your preferred editor

# 4. Start all services
docker-compose up -d

# 5. Check status
docker-compose ps
```

## 📦 What's Included

```
docker-compose/
├── docker-compose.yml       # Main compose configuration
├── .env.example             # Environment variables template
├── Makefile                 # Convenient command shortcuts
├── docker-setup.sh/.bat     # Automated setup scripts
├── verify-docker.sh/.bat    # Pre-flight checks
├── DOCKER.md                # Comprehensive documentation
└── DOCKER-QUICKSTART.md     # 5-minute quick start guide
```

## 🎯 Services

### 1. PostgreSQL Database
- **Container**: `leapmailr-postgres`
- **Port**: 5432
- **Persistent**: Yes (volume: `postgres_data`)

### 2. Backend API (Go)
- **Container**: `leapmailr-backend`
- **Port**: 8080
- **Build from**: `../Dockerfile`
- **Health check**: `/api/v1/health`

### 3. Frontend UI (Next.js)
- **Container**: `leapmailr-frontend`
- **Port**: 3000
- **Build from**: `../../leapmailr-ui/Dockerfile`

## 🔧 Configuration

### Required Configuration

Edit `.env` file before starting:

1. **JWT Secret** (Production only)
   ```env
   JWT_SECRET=your-super-secret-jwt-key-min-32-chars
   ```

2. **Database Password** (Production only)
   ```env
   DB_PASSWORD=your-secure-password
   ```

### Email Service Configuration

Email providers (SMTP, SendGrid, Mailgun, etc.) are now configured through the Email Service API after authentication. 

**No environment variables are required for email configuration.**

After starting the application and registering/logging in, use the `/api/v1/email-services` API endpoints to configure your email services.

### Optional Configuration

- **CORS Origins**: Update for production domain
- **Rate Limiting**: Adjust requests per duration

## 📊 Common Commands

### Using Docker Compose

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Restart a service
docker-compose restart backend

# Rebuild and restart
docker-compose up -d --build

# Check status
docker-compose ps
```

### Using Makefile (Optional)

```bash
# Show all commands
make help

# Setup and start
make setup
make up

# View logs
make logs
make logs-backend
make logs-frontend

# Database operations
make db-shell
make db-backup
make db-restore

# Development
make dev-backend
make dev-frontend

# Health check
make health
```

## 🌐 Access Points

Once running, access:

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **API Health**: http://localhost:8080/api/v1/health
- **Database**: `localhost:5432` (user: `leapmailr`, db: `leapmailr`)

## 📚 Documentation

- **[DOCKER-QUICKSTART.md](DOCKER-QUICKSTART.md)** - Get started in 5 minutes
- **[DOCKER.md](DOCKER.md)** - Comprehensive guide with troubleshooting
- **[../README.md](../README.md)** - Backend API documentation
- **[../../leapmailr-ui/README.md](../../leapmailr-ui/README.md)** - Frontend UI documentation

## 🛠️ Development Workflow

### Option 1: Database in Docker, Services Local

Best for active development with hot reload:

```bash
# Start only database
docker-compose up -d postgres

# Terminal 1: Run backend locally
cd ..
go run main.go

# Terminal 2: Run frontend locally
cd ../../leapmailr-ui
npm run dev
```

### Option 2: All in Docker

Good for testing production-like environment:

```bash
# Start everything
docker-compose up -d

# Watch logs
docker-compose logs -f

# Rebuild after changes
docker-compose up -d --build backend
```

## 🔍 Troubleshooting

### Port Already in Use

```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/Mac
lsof -ti:8080 | xargs kill -9
```

### Services Not Starting

```bash
# Check logs
docker-compose logs backend

# Check service health
docker-compose ps

# Restart service
docker-compose restart backend
```

### Database Connection Issues

```bash
# Check database health
docker-compose exec postgres pg_isready -U leapmailr

# Access database
docker-compose exec postgres psql -U leapmailr -d leapmailr

# Reset database
docker-compose down -v
docker-compose up -d
```

### Build Failures

```bash
# Clear cache and rebuild
docker-compose build --no-cache
docker-compose up -d
```

## 🔒 Production Checklist

Before deploying to production:

- [ ] Change `JWT_SECRET` to strong random value (32+ characters)
- [ ] Update `DB_PASSWORD` to secure password
- [ ] Configure production email provider
- [ ] Update `CORS_ORIGINS` to production domain
- [ ] Enable HTTPS/TLS
- [ ] Set up monitoring and logging
- [ ] Configure automated backups
- [ ] Review resource limits
- [ ] Implement log rotation
- [ ] Set up health check monitoring

## 📦 Volumes

- **postgres_data**: Persistent database storage
  - Location: Docker managed volume
  - Backup: `make db-backup`
  - Remove: `docker-compose down -v` (⚠️ deletes all data)

## 🌐 Networks

- **leapmailr-network**: Bridge network for service communication
  - Isolated network for all LeapMailr services
  - Services can communicate by container name

## 🤝 Contributing

1. Make changes to backend (`../`) or frontend (`../../leapmailr-ui/`)
2. Test locally with `docker-compose up -d --build`
3. Check logs: `docker-compose logs -f`
4. Verify health: `curl http://localhost:8080/api/v1/health`

## 📄 License

See [../LICENSE](../LICENSE) for details.

## 🆘 Support

- **Issues**: Check [DOCKER.md](DOCKER.md) troubleshooting section
- **Logs**: `docker-compose logs -f [service]`
- **Health**: `make health` or `docker-compose ps`

---

For more detailed information, see the comprehensive [DOCKER.md](DOCKER.md) guide.
