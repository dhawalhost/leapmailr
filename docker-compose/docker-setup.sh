#!/bin/bash

# LeapMailr Docker Setup Script
# This script helps you set up the Docker environment

echo "======================================"
echo "LeapMailr Docker Setup"
echo "======================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    echo "Please install Docker Desktop from: https://www.docker.com/products/docker-desktop"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    echo "Please install Docker Compose"
    exit 1
fi

echo -e "${GREEN}✓ Docker is installed${NC}"
echo -e "${GREEN}✓ Docker Compose is installed${NC}"
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}Creating .env file from .env.example...${NC}"
    cp .env.example .env
    echo -e "${GREEN}✓ Created .env file${NC}"
    echo ""
    echo -e "${YELLOW}⚠️  IMPORTANT: Please edit .env file with your configuration${NC}"
    echo "   - Update JWT_SECRET with a strong random value"
    echo "   - Configure your email provider (SMTP, SendGrid, etc.)"
    echo "   - Update DB_PASSWORD for production"
    echo ""
    read -p "Press Enter to continue or Ctrl+C to edit .env first..."
else
    echo -e "${GREEN}✓ .env file exists${NC}"
fi

echo ""
echo "Starting LeapMailr services..."
echo ""

# Pull latest images
echo "Pulling latest base images..."
docker-compose pull || echo "Warning: Could not pull all images, will use existing/build locally"

# Build images
echo ""
echo "Building images (this may take a few minutes)..."
docker-compose build || {
    echo "Error: Build failed. Check the output above."
    exit 1
}

# Start services
echo ""
echo "Starting services..."
docker-compose up -d || {
    echo "Error: Failed to start services. Check the output above."
    exit 1
}

echo ""
echo "Waiting for services to be healthy..."
sleep 10

# Check service status
docker-compose ps

echo ""
echo "======================================"
echo -e "${GREEN}✓ Setup Complete!${NC}"
echo "======================================"
echo ""
echo "Your LeapMailr instance is running:"
echo ""
echo "  Frontend:  http://localhost:3000"
echo "  Backend:   http://localhost:8080"
echo "  Health:    http://localhost:8080/api/v1/health"
echo ""
echo "Useful commands:"
echo "  docker-compose logs -f       # View all logs"
echo "  docker-compose down          # Stop all services"
echo "  docker-compose restart       # Restart services"
echo "  make help                    # See all make commands"
echo ""
echo "For detailed documentation, see:"
echo "  - DOCKER-QUICKSTART.md (quick start guide)"
echo "  - DOCKER.md (comprehensive documentation)"
echo ""
