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
if [ ! -f ".env" ]; then
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

# Start services with docker-compose
echo ""
echo "🚀 Starting services..."
docker-compose up -d

echo ""
echo "⏳ Waiting for services to be healthy..."
sleep 10

# Check health
echo ""
echo "🏥 Health checks..."

# Backend health
if curl -f http://localhost:8080/api/v1/health &> /dev/null; then
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
echo "  API Health:   http://localhost:8080/api/v1/health"
echo "  Database:     localhost:5432"
echo ""
echo "Useful commands:"
echo "  docker-compose logs -f           # View all logs"
echo "  docker-compose logs -f backend   # View backend logs"
echo "  docker-compose down              # Stop all services"
echo "  docker-compose restart backend   # Restart backend"
