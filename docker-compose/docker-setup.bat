@echo off
REM LeapMailr Docker Setup Script (Windows)
REM This script helps you set up the Docker environment

echo ======================================
echo LeapMailr Docker Setup
echo ======================================
echo.

REM Check if Docker is installed
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker is not installed
    echo Please install Docker Desktop from: https://www.docker.com/products/docker-desktop
    exit /b 1
)

REM Check if Docker Compose is installed
docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker Compose is not installed
    echo Please install Docker Compose
    exit /b 1
)

echo [OK] Docker is installed
echo [OK] Docker Compose is installed
echo.

REM Check if .env file exists
if not exist .env (
    echo Creating .env file from .env.example...
    copy .env.example .env
    echo [OK] Created .env file
    echo.
    echo [WARNING] IMPORTANT: Please edit .env file with your configuration
    echo    - Update JWT_SECRET with a strong random value
    echo    - Configure your email provider (SMTP, SendGrid, etc.)
    echo    - Update DB_PASSWORD for production
    echo.
    pause
) else (
    echo [OK] .env file exists
)

echo.
echo Starting LeapMailr services...
echo.

REM Pull latest images
echo Pulling latest base images...
docker-compose pull

REM Build images
echo.
echo Building images (this may take a few minutes)...
docker-compose build

REM Start services
echo.
echo Starting services...
docker-compose up -d

echo.
echo Waiting for services to be healthy...
timeout /t 10 /nobreak >nul

REM Check service status
docker-compose ps

echo.
echo ======================================
echo [SUCCESS] Setup Complete!
echo ======================================
echo.
echo Your LeapMailr instance is running:
echo.
echo   Frontend:  http://localhost:3000
echo   Backend:   http://localhost:8080
echo   Health:    http://localhost:8080/api/v1/health
echo.
echo Useful commands:
echo   docker-compose logs -f       # View all logs
echo   docker-compose down          # Stop all services
echo   docker-compose restart       # Restart services
echo.
echo For detailed documentation, see:
echo   - DOCKER-QUICKSTART.md (quick start guide)
echo   - DOCKER.md (comprehensive documentation)
echo.
pause
