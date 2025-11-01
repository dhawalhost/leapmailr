@echo off
REM LeapMailr Docker Verification Script (Windows)
REM Checks if your system is ready to run Docker

echo ======================================
echo LeapMailr Docker Environment Check
echo ======================================
echo.

set SUCCESS=0
set WARNINGS=0
set ERRORS=0

REM Check Docker
echo Checking Docker...
docker --version >nul 2>&1
if %errorlevel% equ 0 (
    docker --version
    echo [OK] Docker is installed
    set /a SUCCESS+=1
) else (
    echo [ERROR] Docker is not installed
    echo   Install from: https://www.docker.com/products/docker-desktop
    set /a ERRORS+=1
)

REM Check Docker Compose
echo.
echo Checking Docker Compose...
docker-compose --version >nul 2>&1
if %errorlevel% equ 0 (
    docker-compose --version
    echo [OK] Docker Compose is installed
    set /a SUCCESS+=1
) else (
    echo [ERROR] Docker Compose is not installed
    set /a ERRORS+=1
)

REM Check Docker daemon
echo.
echo Checking Docker daemon...
docker info >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] Docker daemon is running
    set /a SUCCESS+=1
) else (
    echo [ERROR] Docker daemon is not running
    echo   Please start Docker Desktop
    set /a ERRORS+=1
)

REM Check ports
echo.
echo Checking required ports...

REM Check port 3000
netstat -an | findstr ":3000" | findstr "LISTENING" >nul 2>&1
if %errorlevel% equ 0 (
    echo [WARNING] Port 3000 is in use
    set /a WARNINGS+=1
) else (
    echo [OK] Port 3000 is available
    set /a SUCCESS+=1
)

REM Check port 8080
netstat -an | findstr ":8080" | findstr "LISTENING" >nul 2>&1
if %errorlevel% equ 0 (
    echo [WARNING] Port 8080 is in use
    set /a WARNINGS+=1
) else (
    echo [OK] Port 8080 is available
    set /a SUCCESS+=1
)

REM Check port 5432
netstat -an | findstr ":5432" | findstr "LISTENING" >nul 2>&1
if %errorlevel% equ 0 (
    echo [WARNING] Port 5432 is in use
    set /a WARNINGS+=1
) else (
    echo [OK] Port 5432 is available
    set /a SUCCESS+=1
)

REM Check .env file
echo.
echo Checking .env file...
if exist .env (
    echo [OK] .env file exists
    set /a SUCCESS+=1
) else (
    echo [WARNING] .env file not found
    echo   Will be created from .env.example during setup
    set /a WARNINGS+=1
)

REM Summary
echo.
echo ======================================
echo Summary
echo ======================================
echo [OK] Success: %SUCCESS%
if %WARNINGS% gtr 0 (
    echo [WARNING] Warnings: %WARNINGS%
)
if %ERRORS% gtr 0 (
    echo [ERROR] Errors: %ERRORS%
)
echo.

if %ERRORS% equ 0 (
    echo [SUCCESS] Your system is ready for LeapMailr!
    echo.
    echo Next steps:
    echo   docker-setup.bat     # Run setup script
    exit /b 0
) else (
    echo [ERROR] Please fix the errors above before continuing
    exit /b 1
)
