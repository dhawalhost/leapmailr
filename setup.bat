@echo off
REM Quick Setup Script for LeapMailR (Windows)
REM Run this script to set up everything automatically

echo ==========================================
echo LeapMailR - Quick Setup Script (Windows)
echo ==========================================
echo.

REM ==========================================
REM 1. Check Prerequisites
REM ==========================================
echo Step 1: Checking prerequisites...

where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed. Please install Go 1.23 or higher.
    exit /b 1
)
echo [OK] Go is installed

where psql >nul 2>&1
if %errorlevel% neq 0 (
    echo [WARNING] PostgreSQL is not installed. Please install PostgreSQL 14 or higher.
) else (
    echo [OK] PostgreSQL is installed
)

where redis-cli >nul 2>&1
if %errorlevel% neq 0 (
    echo [WARNING] Redis is not installed. Installing Redis is recommended.
) else (
    echo [OK] Redis is installed
)

where node >nul 2>&1
if %errorlevel% neq 0 (
    echo [WARNING] Node.js is not installed. Frontend setup will be skipped.
) else (
    echo [OK] Node.js is installed
)

echo.

REM ==========================================
REM 2. Create Directories
REM ==========================================
echo Step 2: Creating directories...

if not exist secrets mkdir secrets
if not exist backups mkdir backups
if not exist logs mkdir logs

echo [OK] Directories created

echo.

REM ==========================================
REM 3. Generate Secrets (using PowerShell)
REM ==========================================
echo Step 3: Generating secrets...

if not exist config.env (
    echo Generating new secrets...
    
    REM Generate secrets using PowerShell
    for /f "delims=" %%i in ('powershell -Command "[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))"') do set ENCRYPTION_KEY=%%i
    for /f "delims=" %%i in ('powershell -Command "[Convert]::ToBase64String((1..64 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))"') do set JWT_SECRET_RAW=%%i
    for /f "delims=" %%i in ('powershell -Command "[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))"') do set SESSION_SECRET=%%i
    
    REM Clean JWT secret (remove = + /)
    set JWT_SECRET=%JWT_SECRET_RAW:~0,64%
    
    echo [OK] Secrets generated
) else (
    echo [WARNING] config.env already exists
)

echo.

REM ==========================================
REM 4. Create Configuration File
REM ==========================================
echo Step 4: Creating configuration file...

if not exist config.env (
    (
        echo # ============================================
        echo # SERVER CONFIGURATION
        echo # ============================================
        echo PORT=8080
        echo ENVIRONMENT=development
        echo FRONTEND_URL=http://localhost:3000
        echo.
        echo # ============================================
        echo # DATABASE CONFIGURATION
        echo # ============================================
        echo DB_HOST=localhost
        echo DB_PORT=5432
        echo DB_USER=leapmailr
        echo DB_PASSWORD=leapmailr_dev_123
        echo DB_NAME=leapmailr
        echo DB_SSLMODE=disable
        echo.
        echo # ============================================
        echo # REDIS CONFIGURATION
        echo # ============================================
        echo REDIS_HOST=localhost
        echo REDIS_PORT=6379
        echo REDIS_PASSWORD=
        echo REDIS_DB=0
        echo.
        echo # ============================================
        echo # SECURITY SECRETS
        echo # ============================================
        echo JWT_SECRET=%JWT_SECRET%
        echo ENCRYPTION_KEY=%ENCRYPTION_KEY%
        echo SESSION_SECRET=%SESSION_SECRET%
        echo.
        echo # ============================================
        echo # CORS CONFIGURATION
        echo # ============================================
        echo ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000
        echo.
        echo # ============================================
        echo # RATE LIMITING
        echo # ============================================
        echo RATE_LIMIT_ENABLED=true
        echo RATE_LIMIT_GLOBAL=100
        echo RATE_LIMIT_AUTH=10
        echo RATE_LIMIT_API=50
        echo.
        echo # ============================================
        echo # EMAIL CONFIGURATION
        echo # ============================================
        echo SMTP_HOST=smtp.mailtrap.io
        echo SMTP_PORT=2525
        echo SMTP_USER=your_mailtrap_user
        echo SMTP_PASSWORD=your_mailtrap_password
        echo FROM_EMAIL=noreply@leapmailr.local
        echo FROM_NAME=LeapMailR
        echo.
        echo # ============================================
        echo # LOGGING
        echo # ============================================
        echo LOG_LEVEL=info
        echo LOG_FORMAT=json
        echo.
        echo # ============================================
        echo # SECURITY HEADERS
        echo # ============================================
        echo HSTS_MAX_AGE=31536000
        echo FORCE_HTTPS=false
        echo.
        echo # ============================================
        echo # MFA CONFIGURATION
        echo # ============================================
        echo MFA_ISSUER=LeapMailR
        echo MFA_BACKUP_CODES_COUNT=10
        echo.
        echo # ============================================
        echo # SECRETS MANAGEMENT
        echo # ============================================
        echo SECRETS_PROVIDER=local
        echo SECRETS_DIR=./secrets
        echo.
        echo # ============================================
        echo # BACKUP CONFIGURATION
        echo # ============================================
        echo BACKUP_DIR=./backups
        echo BACKUP_RETENTION_DAYS=30
        echo.
        echo # ============================================
        echo # MONITORING
        echo # ============================================
        echo METRICS_ENABLED=true
        echo METRICS_PORT=9090
    ) > config.env
    
    echo [OK] Created config.env
) else (
    echo [WARNING] config.env already exists
)

echo.

REM ==========================================
REM 5. Setup Database
REM ==========================================
echo Step 5: Setting up PostgreSQL database...

where psql >nul 2>&1
if %errorlevel% neq 0 (
    echo [WARNING] PostgreSQL not found. Skipping database setup.
    echo Please install PostgreSQL and run the following commands:
    echo   psql -U postgres -c "CREATE USER leapmailr WITH PASSWORD 'leapmailr_dev_123';"
    echo   psql -U postgres -c "CREATE DATABASE leapmailr OWNER leapmailr;"
) else (
    echo Creating database user and database...
    psql -U postgres -c "CREATE USER leapmailr WITH PASSWORD 'leapmailr_dev_123';" 2>nul
    psql -U postgres -c "CREATE DATABASE leapmailr OWNER leapmailr;" 2>nul
    psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE leapmailr TO leapmailr;" 2>nul
    echo [OK] Database setup attempted
)

echo.

REM ==========================================
REM 6. Install Go Dependencies
REM ==========================================
echo Step 6: Installing Go dependencies...

go mod download
go mod tidy

echo [OK] Go dependencies installed

echo.

REM ==========================================
REM 7. Build Application
REM ==========================================
echo Step 7: Building application...

go build -o leapmailr.exe .

if %errorlevel% equ 0 (
    echo [OK] Application built successfully
) else (
    echo [ERROR] Build failed
    exit /b 1
)

echo.

REM ==========================================
REM 8. Setup Frontend
REM ==========================================
where node >nul 2>&1
if %errorlevel% equ 0 (
    if exist ..\leapmailr-ui (
        echo Step 8: Setting up frontend...
        
        cd ..\leapmailr-ui
        
        if not exist .env.local (
            (
                echo NEXT_PUBLIC_API_URL=http://localhost:8080
                echo NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
                echo NEXT_PUBLIC_ENVIRONMENT=development
                echo NEXT_PUBLIC_ENABLE_MFA=true
                echo NEXT_PUBLIC_ENABLE_ANALYTICS=true
            ) > .env.local
            echo [OK] Created frontend .env.local
        )
        
        if not exist node_modules (
            npm install
            echo [OK] Frontend dependencies installed
        )
        
        cd ..\leapmailr
        echo.
    )
)

REM ==========================================
REM Create Start Scripts
REM ==========================================
echo Creating start scripts...

(
    echo @echo off
    echo REM Start LeapMailR Backend
    echo.
    echo REM Load environment variables
    echo for /f "tokens=*" %%%%a in ^(config.env^) do ^(
    echo     set %%%%a
    echo ^)
    echo.
    echo REM Start the application
    echo leapmailr.exe
) > start.bat

echo [OK] Created start.bat

echo.

REM ==========================================
REM Summary
REM ==========================================
echo ==========================================
echo Setup Complete!
echo ==========================================
echo.
echo Next steps:
echo.
echo 1. Start the backend:
echo    start.bat
echo    or
echo    go run .
echo.

where node >nul 2>&1
if %errorlevel% equ 0 (
    if exist ..\leapmailr-ui (
        echo 2. Start the frontend ^(in a new terminal^):
        echo    cd ..\leapmailr-ui
        echo    npm run dev
        echo.
    )
)

echo 3. Access the application:
echo    Backend API: http://localhost:8080
echo    Health Check: http://localhost:8080/health
echo    Metrics: http://localhost:8080/metrics

where node >nul 2>&1
if %errorlevel% equ 0 (
    if exist ..\leapmailr-ui (
        echo    Frontend: http://localhost:3000
    )
)

echo.
echo For more information, see docs\SETUP_GUIDE.md
echo.
echo ==========================================

pause
