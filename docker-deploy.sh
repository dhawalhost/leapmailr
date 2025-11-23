#!/bin/bash
# Local Docker Build and Test Script

set -e

echo "🚀 LeapMailr Local Docker Build & Test"
echo "======================================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed"
    exit 1
fi

# Create network if it doesn't exist
if ! docker network inspect leapmailr-network &> /dev/null; then
    echo "📡 Creating Docker network..."
    docker network create leapmailr-network
    echo -e "${GREEN}✓ Network created${NC}"
fi

# Build backend
echo ""
echo "🔨 Building backend Docker image..."
docker build -t leapmailr-backend:local .
echo -e "${GREEN}✓ Backend image built${NC}"

# Check if .env exists
if [[ ! -f ".env" ]]; then
    echo -e "${YELLOW}⚠️  .env file not found, creating from example...${NC}"
    cat > .env <<EOF
DB_HOST=postgres
DB_PORT=5432
DB_USER=leapmailr
DB_PASSWORD=leapmailr123
DB_NAME=leapmailr
PORT=8080
GIN_MODE=release
JWT_SECRET=$(openssl rand -base64 32)
EOF
    echo -e "${GREEN}✓ .env file created${NC}"
fi

# Check if PostgreSQL container already exists and is running
echo ""
echo "🔍 Checking PostgreSQL status..."
POSTGRES_EXISTS=$(docker ps -a --format '{{.Names}}' | grep '^leapmailr-postgres$' || echo "")
POSTGRES_RUNNING=$(docker ps --format '{{.Names}}' | grep '^leapmailr-postgres$' || echo "")

if [[ -n "$POSTGRES_EXISTS" ]]; then
    if [[ -n "$POSTGRES_RUNNING" ]]; then
        echo -e "${GREEN}✓ PostgreSQL is already running${NC}"
    else
        echo -e "${YELLOW}⚠️  PostgreSQL container exists but is stopped, starting...${NC}"
        docker start leapmailr-postgres
        echo -e "${GREEN}✓ PostgreSQL started${NC}"
        sleep 3
    fi
else
    echo "📦 Creating new PostgreSQL container..."
    docker-compose up -d postgres
    echo -e "${GREEN}✓ PostgreSQL created${NC}"
    sleep 5
fi

# Remove existing backend container if it exists
BACKEND_EXISTS=$(docker ps -a --format '{{.Names}}' | grep '^leapmailr-backend$' || echo "")
if [[ -n "$BACKEND_EXISTS" ]]; then
    echo ""
    echo "🔄 Removing existing backend container..."
    docker rm -f leapmailr-backend
fi

# Start/update backend service
echo ""
echo "🚀 Starting backend service..."
docker-compose up -d --no-deps backend

echo ""
echo "⏳ Waiting for services to be healthy..."
sleep 10

# Check health
echo ""
echo "🏥 Health checks..."

# Backend health
if curl -f --max-time 2 http://localhost:8080/health &> /dev/null; then
    echo -e "${GREEN}✓ Backend is healthy${NC}"
else
    echo -e "${YELLOW}⚠️  Backend health check failed${NC}"
    echo "   Check logs: docker logs leapmailr-backend"
fi

# Show running containers
echo ""
echo "📊 Running containers:"
docker ps --filter "network=leapmailr-network"

echo ""
echo "✅ Deployment complete!"
echo ""
echo "Access your application:"
echo "  Backend API:  http://localhost:8080"
echo "  API Health:   http://localhost:8080/health"
echo "  Database:     localhost:5432"
echo ""
echo "Useful commands:"
echo "  docker-compose logs -f           # View all logs"
echo "  docker-compose logs -f backend   # View backend logs"
echo "  docker-compose down              # Stop all services"
echo "  docker-compose restart backend   # Restart backend"
