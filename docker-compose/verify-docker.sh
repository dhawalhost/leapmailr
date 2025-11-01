#!/bin/bash

# LeapMailr Docker Verification Script
# Checks if your system is ready to run Docker

echo "======================================"
echo "LeapMailr Docker Environment Check"
echo "======================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SUCCESS=0
WARNINGS=0
ERRORS=0

# Check Docker
echo -n "Checking Docker... "
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version | cut -d ' ' -f3 | cut -d ',' -f1)
    echo -e "${GREEN}✓ Installed (version $DOCKER_VERSION)${NC}"
    ((SUCCESS++))
else
    echo -e "${RED}✗ Not installed${NC}"
    echo "  Install from: https://www.docker.com/products/docker-desktop"
    ((ERRORS++))
fi

# Check Docker Compose
echo -n "Checking Docker Compose... "
if command -v docker-compose &> /dev/null; then
    COMPOSE_VERSION=$(docker-compose --version | cut -d ' ' -f4 | cut -d ',' -f1)
    echo -e "${GREEN}✓ Installed (version $COMPOSE_VERSION)${NC}"
    ((SUCCESS++))
else
    echo -e "${RED}✗ Not installed${NC}"
    ((ERRORS++))
fi

# Check Docker daemon
echo -n "Checking Docker daemon... "
if docker info &> /dev/null; then
    echo -e "${GREEN}✓ Running${NC}"
    ((SUCCESS++))
else
    echo -e "${RED}✗ Not running${NC}"
    echo "  Please start Docker Desktop"
    ((ERRORS++))
fi

# Check available disk space
echo -n "Checking disk space... "
if command -v df &> /dev/null; then
    AVAILABLE=$(df -h . 2>/dev/null | awk 'NR==2 {print $4}' | sed 's/G//' 2>/dev/null || echo "0")
    if [ ! -z "$AVAILABLE" ] && [ "$AVAILABLE" -gt 10 ] 2>/dev/null; then
        echo -e "${GREEN}✓ Sufficient ($AVAILABLE GB available)${NC}"
        ((SUCCESS++))
    elif [ ! -z "$AVAILABLE" ] && [ "$AVAILABLE" -gt 0 ] 2>/dev/null; then
        echo -e "${YELLOW}⚠ Low disk space ($AVAILABLE GB available)${NC}"
        echo "  Recommended: 10GB+ free space"
        ((WARNINGS++))
    else
        echo -e "${YELLOW}⚠ Could not determine${NC}"
        ((WARNINGS++))
    fi
else
    echo -e "${YELLOW}⚠ Could not check (df not available)${NC}"
    ((WARNINGS++))
fi

# Check ports
echo "Checking required ports..."

check_port() {
    PORT=$1
    if command -v lsof &> /dev/null; then
        if lsof -Pi :$PORT -sTCP:LISTEN -t &> /dev/null 2>&1; then
            echo -e "  Port $PORT: ${YELLOW}⚠ In use${NC}"
            ((WARNINGS++))
            return 1
        else
            echo -e "  Port $PORT: ${GREEN}✓ Available${NC}"
            ((SUCCESS++))
            return 0
        fi
    elif command -v netstat &> /dev/null; then
        if netstat -an 2>/dev/null | grep ":$PORT " | grep -i "LISTEN" &> /dev/null; then
            echo -e "  Port $PORT: ${YELLOW}⚠ In use${NC}"
            ((WARNINGS++))
            return 1
        else
            echo -e "  Port $PORT: ${GREEN}✓ Available${NC}"
            ((SUCCESS++))
            return 0
        fi
    else
        echo -e "  Port $PORT: ${YELLOW}⚠ Could not check${NC}"
        ((WARNINGS++))
        return 0
    fi
}

check_port 3000
check_port 8080
check_port 5432

# Check if .env exists
echo -n "Checking .env file... "
if [ -f .env ]; then
    echo -e "${GREEN}✓ Exists${NC}"
    ((SUCCESS++))
else
    echo -e "${YELLOW}⚠ Not found${NC}"
    echo "  Will be created from .env.example during setup"
    ((WARNINGS++))
fi

# Summary
echo ""
echo "======================================"
echo "Summary"
echo "======================================"
echo -e "${GREEN}✓ Success: $SUCCESS${NC}"
if [ $WARNINGS -gt 0 ]; then
    echo -e "${YELLOW}⚠ Warnings: $WARNINGS${NC}"
fi
if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}✗ Errors: $ERRORS${NC}"
fi
echo ""

if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✓ Your system is ready for LeapMailr!${NC}"
    echo ""
    echo "Next steps:"
    echo "  ./docker-setup.sh    # Run setup script"
    echo "  make setup           # Or use Makefile"
    exit 0
else
    echo -e "${RED}✗ Please fix the errors above before continuing${NC}"
    exit 1
fi
