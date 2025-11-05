@echo off
REM
REM Automated PostgreSQL Database Backup Script for Windows with S3 Upload
REM GAP-AV-001: Automated database backups for disaster recovery
REM
REM Prerequisites:
REM   - PostgreSQL client tools (pg_dump.exe)
REM   - AWS CLI for Windows
REM   - Configured AWS credentials
REM   - S3 bucket created for backups
REM
REM Setup:
REM   1. Copy this script to C:\leapmailr\scripts\backup-db.bat
REM   2. Update configuration variables below
REM   3. Test manually: backup-db.bat
REM   4. Schedule with Task Scheduler:
REM      - Open Task Scheduler
REM      - Create Basic Task
REM      - Name: "Leapmailr Database Backup"
REM      - Trigger: Daily at 2:00 AM
REM      - Action: Start a program
REM      - Program: C:\leapmailr\scripts\backup-db.bat
REM      - Finish and test
REM

setlocal enabledelayedexpansion

REM ========== CONFIGURATION ==========
set "DB_NAME=leapmailr"
set "DB_USER=postgres"
set "DB_HOST=localhost"
set "DB_PORT=5432"
set "PGPASSWORD=your_password_here"

set "BACKUP_DIR=C:\backups\leapmailr"
set "S3_BUCKET=s3://your-bucket-name/leapmailr/backups"
set "RETENTION_DAYS=30"
set "MAX_LOCAL_BACKUPS=7"

REM PostgreSQL bin path (update if different)
set "PG_BIN=C:\Program Files\PostgreSQL\15\bin"

REM ========== SCRIPT STARTS HERE ==========

REM Generate timestamp
for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /value') do set datetime=%%I
set TIMESTAMP=%datetime:~0,8%_%datetime:~8,6%
set DATE_ONLY=%datetime:~0,8%

set "BACKUP_FILE=%BACKUP_DIR%\leapmailr_%TIMESTAMP%.sql"
set "BACKUP_FILE_GZ=%BACKUP_FILE%.gz"
set "BACKUP_FILENAME=leapmailr_%TIMESTAMP%.sql.gz"
set "LOG_FILE=C:\logs\leapmailr-backup_%DATE_ONLY%.log"

REM Create directories if they don't exist
if not exist "%BACKUP_DIR%" mkdir "%BACKUP_DIR%"
if not exist "C:\logs" mkdir "C:\logs"

echo ========================================= >> "%LOG_FILE%"
echo [%date% %time%] Starting Leapmailr Database Backup >> "%LOG_FILE%"
echo ========================================= >> "%LOG_FILE%"
echo Database: %DB_NAME% >> "%LOG_FILE%"
echo Backup file: %BACKUP_FILE_GZ% >> "%LOG_FILE%"

REM Check if pg_dump exists
if not exist "%PG_BIN%\pg_dump.exe" (
    echo [%date% %time%] ERROR: pg_dump.exe not found at %PG_BIN% >> "%LOG_FILE%"
    echo ERROR: PostgreSQL not found!
    exit /b 1
)

REM Check if AWS CLI exists
where aws >nul 2>&1
if %errorlevel% neq 0 (
    echo [%date% %time%] ERROR: AWS CLI not found >> "%LOG_FILE%"
    echo ERROR: AWS CLI not installed!
    exit /b 1
)

REM Step 1: Create database backup
echo [%date% %time%] Creating database backup... >> "%LOG_FILE%"
echo Creating database backup...

"%PG_BIN%\pg_dump.exe" -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% --no-owner --no-acl -f "%BACKUP_FILE%"

if %errorlevel% neq 0 (
    echo [%date% %time%] ERROR: Database backup failed >> "%LOG_FILE%"
    echo ERROR: Database backup failed!
    exit /b 1
)

REM Step 2: Compress backup using PowerShell
echo [%date% %time%] Compressing backup... >> "%LOG_FILE%"
echo Compressing backup...

powershell -Command "& {$input = '%BACKUP_FILE%'; $output = '%BACKUP_FILE_GZ%'; $fileInput = [System.IO.File]::OpenRead($input); $fileOutput = [System.IO.File]::Create($output); $gzipStream = New-Object System.IO.Compression.GZipStream($fileOutput, [System.IO.Compression.CompressionMode]::Compress); $fileInput.CopyTo($gzipStream); $gzipStream.Close(); $fileOutput.Close(); $fileInput.Close()}"

if %errorlevel% neq 0 (
    echo [%date% %time%] ERROR: Compression failed >> "%LOG_FILE%"
    del "%BACKUP_FILE%"
    exit /b 1
)

REM Delete uncompressed file
del "%BACKUP_FILE%"

REM Get file size
for %%A in ("%BACKUP_FILE_GZ%") do set BACKUP_SIZE=%%~zA
set /a BACKUP_SIZE_MB=%BACKUP_SIZE% / 1048576
echo [%date% %time%] Backup created: %BACKUP_SIZE_MB% MB >> "%LOG_FILE%"
echo Backup created: %BACKUP_SIZE_MB% MB

REM Step 3: Upload to S3
echo [%date% %time%] Uploading to S3... >> "%LOG_FILE%"
echo Uploading to S3...

aws s3 cp "%BACKUP_FILE_GZ%" "%S3_BUCKET%/" --storage-class STANDARD_IA --only-show-errors

if %errorlevel% neq 0 (
    echo [%date% %time%] ERROR: S3 upload failed >> "%LOG_FILE%"
    echo ERROR: S3 upload failed!
    exit /b 1
)

echo [%date% %time%] Upload successful >> "%LOG_FILE%"
echo Upload successful

REM Step 4: Verify S3 upload
echo [%date% %time%] Verifying S3 upload... >> "%LOG_FILE%"

aws s3 ls "%S3_BUCKET%/%BACKUP_FILENAME%" >nul 2>&1

if %errorlevel% neq 0 (
    echo [%date% %time%] ERROR: S3 verification failed >> "%LOG_FILE%"
    echo WARNING: Could not verify S3 upload
) else (
    echo [%date% %time%] S3 verification passed >> "%LOG_FILE%"
    echo Verification passed
)

REM Step 5: Clean up old local backups
echo [%date% %time%] Cleaning up old local backups... >> "%LOG_FILE%"

set COUNT=0
for /f %%F in ('dir /b /o-d "%BACKUP_DIR%\leapmailr_*.sql.gz" 2^>nul') do (
    set /a COUNT+=1
    if !COUNT! gtr %MAX_LOCAL_BACKUPS% (
        del "%BACKUP_DIR%\%%F"
        echo [%date% %time%] Deleted old backup: %%F >> "%LOG_FILE%"
    )
)

REM Summary
echo ========================================= >> "%LOG_FILE%"
echo [%date% %time%] BACKUP COMPLETED SUCCESSFULLY >> "%LOG_FILE%"
echo ========================================= >> "%LOG_FILE%"
echo Summary: >> "%LOG_FILE%"
echo   - Database: %DB_NAME% >> "%LOG_FILE%"
echo   - Backup file: %BACKUP_FILENAME% >> "%LOG_FILE%"
echo   - Size: %BACKUP_SIZE_MB% MB >> "%LOG_FILE%"
echo   - S3 location: %S3_BUCKET%/%BACKUP_FILENAME% >> "%LOG_FILE%"
echo ========================================= >> "%LOG_FILE%"

echo.
echo ========================================
echo BACKUP COMPLETED SUCCESSFULLY
echo ========================================
echo Backup file: %BACKUP_FILENAME%
echo Size: %BACKUP_SIZE_MB% MB
echo S3 location: %S3_BUCKET%/%BACKUP_FILENAME%
echo ========================================

exit /b 0
