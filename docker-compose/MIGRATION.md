# Docker Compose Migration

## What Changed

The full stack Docker Compose setup has been moved from the root directory to `leapmailr/docker-compose/` for better organization.

### Before (Old Location)
```
c:\Users\dhawa\go\src\
├── docker-compose.yml
├── .env.example
├── Makefile
├── docker-setup.sh
├── docker-setup.bat
├── verify-docker.sh
├── verify-docker.bat
├── DOCKER.md
├── DOCKER-QUICKSTART.md
├── leapmailr/
└── leapmailr-ui/
```

### After (New Location)
```
c:\Users\dhawa\go\src\
├── leapmailr/
│   ├── docker-compose/          ← New location
│   │   ├── docker-compose.yml
│   │   ├── .env.example
│   │   ├── .gitignore
│   │   ├── Makefile
│   │   ├── README.md
│   │   ├── DOCKER.md
│   │   ├── DOCKER-QUICKSTART.md
│   │   ├── docker-setup.sh
│   │   ├── docker-setup.bat
│   │   ├── verify-docker.sh
│   │   └── verify-docker.bat
│   ├── config/
│   ├── handlers/
│   └── ... (other backend files)
└── leapmailr-ui/
    ├── docker-compose.yml        ← Standalone UI compose
    ├── docker-compose.dev.yml
    ├── .env.example
    ├── DOCKER.md
    └── ... (other frontend files)
```

## Key Updates

### 1. Docker Compose Configuration
- **File**: `docker-compose/docker-compose.yml`
- **Changes**: Updated build contexts:
  - Backend: `context: ..` (parent directory)
  - Frontend: `context: ../../leapmailr-ui`

### 2. Makefile
- **File**: `docker-compose/Makefile`
- **Changes**: Updated paths for local development:
  - Backend: `cd .. && go run main.go`
  - Frontend: `cd ../../leapmailr-ui && npm run dev`

### 3. Setup Scripts
- **Files**: `docker-setup.sh`, `docker-setup.bat`
- **Changes**: Updated messaging and paths

### 4. Documentation
- **Files**: `DOCKER.md`, `DOCKER-QUICKSTART.md`, `README.md`
- **Changes**: Updated all references to use new directory structure

### 5. Main README
- **File**: `leapmailr/README.md`
- **Changes**: Updated Quick Start section to reference `docker-compose/` directory

## How to Use

### Full Stack (All Services)

```bash
# Navigate to docker-compose directory
cd leapmailr/docker-compose

# Run setup
./docker-setup.sh  # Linux/Mac
docker-setup.bat   # Windows

# Or manually
docker-compose up -d
```

### UI Only (Standalone)

```bash
# Navigate to UI directory
cd leapmailr-ui

# Production
docker-compose up -d

# Development (with hot reload)
docker-compose -f docker-compose.dev.yml up
```

### Backend Only

```bash
# Navigate to backend directory
cd leapmailr

# Build and run
docker build -t leapmailr-backend .
docker run -p 8080:8080 leapmailr-backend
```

## Benefits of New Structure

1. **Better Organization**: Full stack setup is now part of the backend repository
2. **Clearer Separation**: UI has its own standalone Docker setup
3. **Easier to Find**: Docker files are now in logical locations
4. **Independent Deployment**: UI and backend can be deployed separately
5. **Version Control**: Each repo manages its own Docker configuration

## Migration Guide

If you have an existing setup:

1. **Stop old containers**:
   ```bash
   cd c:\Users\dhawa\go\src
   docker-compose down
   ```

2. **Use new location**:
   ```bash
   cd leapmailr/docker-compose
   cp .env.example .env
   # Edit .env with your settings
   docker-compose up -d
   ```

3. **Data migration** (if needed):
   - Database volume name changed from `leapmailr_postgres_data` to `leapmailr_docker-compose_postgres_data`
   - To keep existing data, you can:
     ```bash
     # Backup from old
     docker-compose exec postgres pg_dump -U leapmailr leapmailr > backup.sql
     
     # Restore to new
     cd leapmailr/docker-compose
     docker-compose up -d postgres
     docker-compose exec -T postgres psql -U leapmailr leapmailr < ../../backup.sql
     ```

## Files You Can Remove

The following files in the root directory (`c:\Users\dhawa\go\src\`) are now redundant:

- `docker-compose.yml`
- `.env.example`
- `Makefile`
- `docker-setup.sh`
- `docker-setup.bat`
- `verify-docker.sh`
- `verify-docker.bat`
- `DOCKER.md`
- `DOCKER-QUICKSTART.md`

**Note**: Keep these if you want backward compatibility or have other projects using them.

## Rollback

If you need to revert:

1. Copy files back to root directory
2. Update paths in `docker-compose.yml`:
   - Backend: `context: ./leapmailr`
   - Frontend: `context: ./leapmailr-ui`
3. Run from root: `docker-compose up -d`

## Support

For issues:
- Full stack: See `leapmailr/docker-compose/DOCKER.md`
- UI only: See `leapmailr-ui/DOCKER.md`
- Backend: See `leapmailr/README.md`
