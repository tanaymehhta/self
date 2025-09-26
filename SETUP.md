# Self - Development Setup Guide

This guide will help you set up the Self application for local development after completing Step 2 (Foundation Architecture).

## Prerequisites

- **Docker & Docker Compose** - For infrastructure services
- **Node.js 20+** - For frontend development
- **Go 1.21+** - For backend development
- **Supabase Account** - For managed PostgreSQL database

## Step 1: Infrastructure Services

Start the required infrastructure services using Docker Compose:

```bash
# Start PostgreSQL, Redis, MinIO, NATS, and Qdrant
docker-compose up -d postgres redis minio nats qdrant

# Verify services are running
docker-compose ps
```

Services will be available at:
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379
- **MinIO**: localhost:9000 (Console: localhost:9001)
- **NATS**: localhost:4222 (Monitoring: localhost:8222)
- **Qdrant**: localhost:6333

## Step 2: Database Setup

### Option A: Use Supabase (Recommended)

1. **Create Supabase Project**:
   - Go to [supabase.com](https://supabase.com)
   - Create new project: "self-app"
   - Note down your project URL and API keys

2. **Run Database Schema**:
   - Open Supabase SQL Editor
   - Copy contents of `database/schema.sql`
   - Execute the SQL to create tables and policies

3. **Configure Environment**:
   ```bash
   # Copy environment template
   cp backend/.env.example backend/.env
   cp frontend/.env.local.example frontend/.env.local

   # Edit backend/.env with your Supabase credentials
   SUPABASE_URL=https://your-project.supabase.co
   SUPABASE_ANON_KEY=your-anon-key
   SUPABASE_SERVICE_KEY=your-service-key
   DATABASE_URL=postgresql://postgres:your-password@db.your-project.supabase.co:5432/postgres

   # Edit frontend/.env.local with your Supabase credentials
   NEXT_PUBLIC_SUPABASE_URL=https://your-project.supabase.co
   NEXT_PUBLIC_SUPABASE_ANON_KEY=your-anon-key
   ```

### Option B: Use Local PostgreSQL (Alternative)

If you prefer local PostgreSQL:

1. **Apply Schema to Local DB**:
   ```bash
   # Connect to local PostgreSQL
   psql -h localhost -p 5432 -U postgres -d self_dev -f database/schema.sql

   # Use local DATABASE_URL in backend/.env
   DATABASE_URL=postgresql://postgres:postgres@localhost:5432/self_dev
   ```

## Step 3: Backend Setup

```bash
# Navigate to backend directory
cd backend

# Install Go dependencies
go mod tidy

# Run the server
go run cmd/server/main.go
```

The Go backend will start at **http://localhost:8080**

### Verify Backend is Running

```bash
# Test health endpoint
curl http://localhost:8080/health

# Should return:
# {"status":"healthy","database":"connected","version":"1.0.0"}
```

## Step 4: Frontend Setup

```bash
# Navigate to frontend directory (in new terminal)
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

The Next.js frontend will start at **http://localhost:3000**

## Step 5: Test the Application

1. **Open Browser**: Go to http://localhost:3000

2. **Authentication Flow**:
   - You'll be redirected to `/auth/login`
   - Create account using Supabase Auth
   - After login, you'll see the dashboard

3. **API Integration**:
   - Dashboard should load user data
   - Backend API endpoints should be accessible

## Development Workflow

### Starting Development Session

```bash
# Terminal 1: Start infrastructure
docker-compose up -d postgres redis minio nats qdrant

# Terminal 2: Start backend
cd backend && go run cmd/server/main.go

# Terminal 3: Start frontend
cd frontend && npm run dev
```

### Stopping Development Session

```bash
# Stop applications (Ctrl+C in terminals)
# Stop infrastructure
docker-compose down
```

### Database Migrations

```bash
# Add new migration
cd backend
# Edit database/schema.sql with changes
# Apply to Supabase via SQL Editor

# Or for local PostgreSQL
psql -h localhost -U postgres -d self_dev -f database/schema.sql
```

## Troubleshooting

### Backend Issues

**"Failed to connect to database"**:
- Check Supabase credentials in `backend/.env`
- Verify DATABASE_URL format
- Ensure Supabase project is active

**"Port 8080 already in use"**:
```bash
# Kill process on port 8080
lsof -ti:8080 | xargs kill -9
```

### Frontend Issues

**"Module not found" errors**:
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

**Authentication redirects not working**:
- Check Supabase URL and keys in `frontend/.env.local`
- Verify middleware.ts is configured correctly

### Infrastructure Issues

**Docker services not starting**:
```bash
# Reset Docker environment
docker-compose down -v
docker system prune -f
docker-compose up -d
```

**Port conflicts**:
```bash
# Check which ports are in use
netstat -tulpn | grep :5432
# Kill conflicting processes or change ports in docker-compose.yml
```

## Next Steps

Once the foundation is running:

1. **Test all endpoints** using API client (Postman/Insomnia)
2. **Verify database connections** and user creation
3. **Start Phase 1**: Core Audio Pipeline development
4. **Add desktop application** for file monitoring

## Available Scripts

### Backend
```bash
cd backend
go run cmd/server/main.go      # Start server
go test ./...                  # Run tests
go build cmd/server/main.go    # Build binary
```

### Frontend
```bash
cd frontend
npm run dev        # Start development server
npm run build      # Build for production
npm run start      # Start production server
npm run lint       # Run ESLint
npm run type-check # Run TypeScript checks
```

### Infrastructure
```bash
docker-compose up -d          # Start all services
docker-compose down           # Stop all services
docker-compose logs [service] # View service logs
docker-compose ps             # List running services
```

## Architecture Overview

```
Frontend (Next.js)     Backend (Go + Fiber)     Database (Supabase)
     :3000          →        :8080           →      PostgreSQL
        ↓                      ↓                        ↓
  Authentication         JWT Validation          Row Level Security
        ↓                      ↓                        ↓
  Dashboard UI          API Endpoints           User Data Storage
        ↓                      ↓                        ↓
  Real-time Updates     WebSocket Handler       Real-time Subscriptions
```

The foundation is now ready for Phase 1 development: Core Audio Pipeline! 🚀