#!/bin/bash
# Backend Deployment Script
# This script sets up the LeapMailr backend on a fresh server

set -e

echo "🚀 LeapMailr Backend Deployment Script"
echo "======================================="

# Configuration
APP_USER="leapmailr"
APP_DIR="/opt/leapmailr"
SERVICE_NAME="leapmailr"
BINARY_NAME="leapmailr"

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Please run as root (use sudo)"
    exit 1
fi

echo "📦 Step 1: Creating application user and directories..."
# Create user if doesn't exist
if ! id "$APP_USER" &>/dev/null; then
    useradd -r -s /bin/false -d $APP_DIR $APP_USER
    echo "✓ User $APP_USER created"
else
    echo "✓ User $APP_USER already exists"
fi

# Create directories
mkdir -p $APP_DIR
mkdir -p $APP_DIR/logs
chown -R $APP_USER:$APP_USER $APP_DIR
echo "✓ Directories created"

echo ""
echo "📦 Step 2: Installing dependencies..."
# Update package list
apt-get update -qq

# Install PostgreSQL if not installed
if ! command -v psql &> /dev/null; then
    apt-get install -y postgresql postgresql-contrib
    systemctl enable postgresql
    systemctl start postgresql
    echo "✓ PostgreSQL installed"
else
    echo "✓ PostgreSQL already installed"
fi

echo ""
echo "📦 Step 3: Setting up database..."
# Setup database
read -p "Enter database name [leapmailr]: " DB_NAME
DB_NAME=${DB_NAME:-leapmailr}

read -p "Enter database user [leapmailr]: " DB_USER
DB_USER=${DB_USER:-leapmailr}

read -sp "Enter database password: " DB_PASSWORD
echo ""

# Create database and user
sudo -u postgres psql <<EOF
CREATE DATABASE $DB_NAME;
CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
\c $DB_NAME
GRANT ALL ON SCHEMA public TO $DB_USER;
EOF

echo "✓ Database created"

echo ""
echo "📦 Step 4: Creating environment file..."
# Create .env file
cat > $APP_DIR/.env <<EOF
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
DB_NAME=$DB_NAME

# Server Configuration
PORT=8080
GIN_MODE=release

# JWT Secret (change this!)
JWT_SECRET=$(openssl rand -base64 32)

# Add your other environment variables here
# SMTP_HOST=smtp.example.com
# SMTP_PORT=587
# etc.
EOF

chown $APP_USER:$APP_USER $APP_DIR/.env
chmod 600 $APP_DIR/.env
echo "✓ Environment file created"

echo ""
echo "📦 Step 5: Installing systemd service..."
# Copy systemd service file
if [ -f "./leapmailr.service" ]; then
    cp ./leapmailr.service /etc/systemd/system/
    systemctl daemon-reload
    echo "✓ Service file installed"
else
    echo "⚠️  Service file not found. Please copy manually."
fi

echo ""
echo "📦 Step 6: Installing binary..."
if [ -f "./$BINARY_NAME" ]; then
    cp ./$BINARY_NAME $APP_DIR/
    chmod +x $APP_DIR/$BINARY_NAME
    chown $APP_USER:$APP_USER $APP_DIR/$BINARY_NAME
    echo "✓ Binary installed"
else
    echo "⚠️  Binary not found. Please place the binary in $APP_DIR/"
fi

echo ""
echo "📦 Step 7: Setting up firewall (optional)..."
if command -v ufw &> /dev/null; then
    read -p "Configure UFW firewall? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        ufw allow 8080/tcp
        ufw allow 22/tcp
        ufw --force enable
        echo "✓ Firewall configured"
    fi
fi

echo ""
echo "📦 Step 8: Setting up Nginx reverse proxy (optional)..."
read -p "Install and configure Nginx? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    apt-get install -y nginx
    
    read -p "Enter domain name (e.g., api.leapmailr.com): " DOMAIN
    
    cat > /etc/nginx/sites-available/leapmailr <<EOF
server {
    listen 80;
    server_name $DOMAIN;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
    }
}
EOF

    ln -sf /etc/nginx/sites-available/leapmailr /etc/nginx/sites-enabled/
    nginx -t && systemctl restart nginx
    echo "✓ Nginx configured"
    
    read -p "Install SSL certificate with Certbot? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        apt-get install -y certbot python3-certbot-nginx
        certbot --nginx -d $DOMAIN
        echo "✓ SSL certificate installed"
    fi
fi

echo ""
echo "📦 Step 9: Starting service..."
systemctl enable $SERVICE_NAME
systemctl start $SERVICE_NAME
systemctl status $SERVICE_NAME --no-pager

echo ""
echo "✅ Deployment complete!"
echo ""
echo "Useful commands:"
echo "  systemctl status $SERVICE_NAME    # Check service status"
echo "  systemctl restart $SERVICE_NAME   # Restart service"
echo "  journalctl -u $SERVICE_NAME -f    # View logs"
echo "  sudo -u $APP_USER $APP_DIR/$BINARY_NAME  # Run manually for testing"
echo ""
echo "Next steps:"
echo "  1. Edit $APP_DIR/.env and configure your settings"
echo "  2. Restart the service: systemctl restart $SERVICE_NAME"
echo "  3. Test the API: curl http://localhost:8080/api/v1/health"
