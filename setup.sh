#!/bin/bash
# Quick Setup Script for LeapMailR (Local Development)
# Run this script to set up everything automatically

set -e

# Constants
readonly SEPARATOR_LINE='=========================================='

echo "$SEPARATOR_LINE"
echo "LeapMailR - Quick Setup Script"
echo "$SEPARATOR_LINE"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    local message="$1"
    echo -e "${GREEN}✓ $message${NC}"
    return 0
}

print_warning() {
    local message="$1"
    echo -e "${YELLOW}⚠ $message${NC}"
    return 0
}

print_error() {
    local message="$1"
    echo -e "${RED}✗ $message${NC}"
    exit 1
    return 1
}

# Check if command exists
command_exists() {
    local cmd="$1"
    command -v "$cmd" >/dev/null 2>&1
    return 0
}

# ==========================================
# 1. Check Prerequisites
# ==========================================
echo "Step 1: Checking prerequisites..."

if ! command_exists go; then
    print_error "Go is not installed. Please install Go 1.23 or higher."
fi
print_success "Go is installed ($(go version))"

if ! command_exists psql; then
    print_error "PostgreSQL is not installed. Please install PostgreSQL 14 or higher."
fi
print_success "PostgreSQL is installed"

if ! command_exists redis-cli; then
    print_warning "Redis is not installed. Installing Redis is recommended."
else
    print_success "Redis is installed"
fi

if ! command_exists node; then
    print_warning "Node.js is not installed. Frontend setup will be skipped."
else
    print_success "Node.js is installed ($(node --version))"
fi

echo ""

# ==========================================
# 2. Generate Secrets
# ==========================================
echo "Step 2: Generating secrets..."

# Create directories
mkdir -p secrets backups logs

# Generate encryption key
if ! grep -q "ENCRYPTION_KEY" config.env 2>/dev/null; then
    ENCRYPTION_KEY=$(openssl rand -base64 32)
    print_success "Generated ENCRYPTION_KEY"
else
    print_warning "ENCRYPTION_KEY already exists in config.env"
fi

# Generate JWT secret
if ! grep -q "JWT_SECRET" config.env 2>/dev/null; then
    JWT_SECRET=$(openssl rand -base64 64 | tr -d "=+/" | cut -c1-64)
    print_success "Generated JWT_SECRET"
else
    print_warning "JWT_SECRET already exists in config.env"
fi

# Generate session secret
if ! grep -q "SESSION_SECRET" config.env 2>/dev/null; then
    SESSION_SECRET=$(openssl rand -base64 32)
    print_success "Generated SESSION_SECRET"
else
    print_warning "SESSION_SECRET already exists in config.env"
fi

echo ""

# ==========================================
# 3. Create Configuration File
# ==========================================
echo "Step 3: Creating configuration file..."

if [[ ! -f config.env ]]; then
    cat > config.env << EOF
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
DB_PASSWORD=leapmailr_dev_123
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
JWT_SECRET=$JWT_SECRET
ENCRYPTION_KEY=$ENCRYPTION_KEY
SESSION_SECRET=$SESSION_SECRET

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
# EMAIL CONFIGURATION
# ============================================
SMTP_HOST=smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USER=your_mailtrap_user
SMTP_PASSWORD=your_mailtrap_password
FROM_EMAIL=noreply@leapmailr.local
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
EOF
    chmod 600 config.env
    print_success "Created config.env"
else
    print_warning "config.env already exists"
fi

echo ""

# ==========================================
# 4. Setup Database
# ==========================================
echo "Step 4: Setting up PostgreSQL database..."

# Check if PostgreSQL is running
if ! pg_isready -h localhost -p 5432 >/dev/null 2>&1; then
    print_error "PostgreSQL is not running. Please start PostgreSQL first."
fi

# Create database and user
DB_PASSWORD=$(grep "^DB_PASSWORD=" config.env | cut -d'=' -f2)

# Check if user exists
if psql -U postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='leapmailr'" | grep -q 1; then
    print_warning "Database user 'leapmailr' already exists"
else
    psql -U postgres -c "CREATE USER leapmailr WITH PASSWORD '$DB_PASSWORD';" 2>/dev/null || print_warning "Could not create user (may already exist)"
    print_success "Created database user 'leapmailr'"
fi

# Check if database exists
if psql -U postgres -lqt | cut -d \| -f 1 | grep -qw leapmailr; then
    print_warning "Database 'leapmailr' already exists"
else
    psql -U postgres -c "CREATE DATABASE leapmailr OWNER leapmailr;" 2>/dev/null || print_warning "Could not create database (may already exist)"
    print_success "Created database 'leapmailr'"
fi

psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE leapmailr TO leapmailr;" 2>/dev/null

# Test connection
if PGPASSWORD=$DB_PASSWORD psql -h localhost -U leapmailr -d leapmailr -c "SELECT 1" >/dev/null 2>&1; then
    print_success "Database connection verified"
else
    print_error "Could not connect to database. Please check PostgreSQL configuration."
fi

echo ""

# ==========================================
# 5. Setup Redis (if available)
# ==========================================
echo "Step 5: Setting up Redis..."

if command_exists redis-cli; then
    if redis-cli ping >/dev/null 2>&1; then
        print_success "Redis is running and accessible"
    else
        print_warning "Redis is not running. Starting Redis is recommended."
    fi
else
    print_warning "Redis not found. Rate limiting will use in-memory storage."
fi

echo ""

# ==========================================
# 6. Install Go Dependencies
# ==========================================
echo "Step 6: Installing Go dependencies..."

go mod download
go mod tidy
print_success "Go dependencies installed"

echo ""

# ==========================================
# 7. Build Application
# ==========================================
echo "Step 7: Building application..."

if go build -o leapmailr .; then
    print_success "Application built successfully"
else
    print_error "Build failed. Please check the error messages above."
fi

echo ""

# ==========================================
# 8. Setup Frontend (if Node.js available)
# ==========================================
if command_exists node && [[ -d "../leapmailr-ui" ]]; then
    echo "Step 8: Setting up frontend..."
    
    cd ../leapmailr-ui
    
    if [[ ! -f .env.local ]]; then
        cat > .env.local << EOF
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_ENVIRONMENT=development
NEXT_PUBLIC_ENABLE_MFA=true
NEXT_PUBLIC_ENABLE_ANALYTICS=true
EOF
        print_success "Created frontend .env.local"
    else
        print_warning "Frontend .env.local already exists"
    fi
    
    if [[ ! -d "node_modules" ]]; then
        npm install
        print_success "Frontend dependencies installed"
    else
        print_warning "Frontend dependencies already installed"
    fi
    
    cd - >/dev/null
    echo ""
fi

# ==========================================
# 9. File Permissions
# ==========================================
echo "Step 9: Setting file permissions..."

chmod 600 config.env
chmod 700 secrets/
chmod 700 backups/
chmod 700 logs/
chmod +x scripts/*.sh 2>/dev/null || true

print_success "File permissions set"

echo ""

# ==========================================
# Summary
# ==========================================
echo "=========================================="
echo "Setup Complete! 🎉"
echo "=========================================="
echo ""
echo "Next steps:"
echo ""
echo "1. Start the backend:"
echo "   $ ./leapmailr"
echo "   or"
echo "   $ go run ."
echo ""

if command_exists node && [[ -d "../leapmailr-ui" ]]; then
    echo "2. Start the frontend (in a new terminal):"
    echo "   $ cd ../leapmailr-ui"
    echo "   $ npm run dev"
    echo ""
fi

echo "3. Access the application:"
echo "   Backend API: http://localhost:8080"
echo "   Health Check: http://localhost:8080/health"
   Metrics: http://localhost:8080/metrics"

if command_exists node && [[ -d "../leapmailr-ui" ]]; then
    echo "   Frontend: http://localhost:3000"
fi

echo ""
echo "4. Default credentials will be created on first run."
echo ""
echo "For more information, see docs/SETUP_GUIDE.md"
echo ""
echo "$SEPARATOR_LINE"

# Create a start script
cat > start.sh << 'EOF'
#!/bin/bash
# Start LeapMailR Backend

# Load environment variables
if [[ -f config.env ]]; then
    export $(cat config.env | grep -v '^#' | xargs)
fi

# Start the application
./leapmailr
EOF
chmod +x start.sh

print_success "Created start.sh script"

echo ""
echo "You can now run './start.sh' to start the backend!"
