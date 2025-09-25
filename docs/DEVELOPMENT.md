# Development Guide

## Prerequisites

- **Node.js** 20+ and npm
- **Go** 1.21+
- **Python** 3.11+
- **Rust** 1.70+ (for Tauri desktop app)
- **Docker** and Docker Compose
- **Git** for version control

## Quick Start

### 1. Clone and Install

```bash
git clone https://github.com/yourusername/self.git
cd self
npm install
```

### 2. Start Development Environment

```bash
# Start all infrastructure services
docker-compose up -d postgres redis minio nats qdrant

# Install dependencies for all services
npm run setup

# Start development servers
npm run dev
```

This will start:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- AI Services: http://localhost:8000
- MinIO Console: http://localhost:9001
- NATS Monitoring: http://localhost:8222

### 3. Desktop App Development

```bash
cd desktop-app
cargo tauri dev
```

## Project Structure

```
self/
├── frontend/          # Next.js web application
├── backend/          # Go API server
├── ai-services/      # Python AI processing
├── desktop-app/      # Tauri desktop application
├── integrations/     # External service connectors
├── docs/            # Documentation
└── .github/         # CI/CD workflows
```

## Development Workflow

### Branch Strategy

- `main` - Production-ready code
- `develop` - Integration branch for features
- `feature/*` - Individual feature branches

### Making Changes

```bash
# Start from develop
git checkout develop
git pull origin develop

# Create feature branch
git checkout -b feature/audio-processing

# Make changes and commit
git add .
git commit -m "feat: add audio recording functionality"

# Push and create PR
git push origin feature/audio-processing
```

### Code Style

- **Frontend**: ESLint + Prettier
- **Backend**: gofmt + golangci-lint
- **AI Services**: black + flake8 + isort
- **Desktop**: rustfmt + clippy

### Testing

```bash
# Run all tests
npm test

# Test specific service
npm run test:frontend
npm run test:backend
cd ai-services && pytest
cd desktop-app && cargo test
```

## Environment Variables

### Backend (.env)

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=self_dev
DB_USER=postgres
DB_PASSWORD=postgres

# Redis
REDIS_URL=localhost:6379

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123

# NATS
NATS_URL=localhost:4222

# Qdrant
QDRANT_URL=http://localhost:6333

# JWT
JWT_SECRET=your-secret-key-here

# External APIs
OPENAI_API_KEY=your-openai-key
ELEVENLABS_API_KEY=your-elevenlabs-key
```

### Frontend (.env.local)

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
NEXT_PUBLIC_MINIO_URL=http://localhost:9000
```

### AI Services (.env)

```bash
NATS_URL=localhost:4222
QDRANT_URL=http://localhost:6333
HUGGINGFACE_HUB_CACHE=/app/models
WHISPER_MODEL_PATH=/app/models/whisper
```

## Database Setup

### Run Migrations

```bash
cd backend
go run cmd/migrate/main.go up
```

### Reset Database

```bash
docker-compose down -v postgres
docker-compose up -d postgres
cd backend && go run cmd/migrate/main.go up
```

## Debugging

### Backend Debugging

```bash
cd backend
go run -race cmd/server/main.go
```

### Frontend Debugging

```bash
cd frontend
npm run dev -- --turbo
```

### AI Services Debugging

```bash
cd ai-services
python -m debugpy --listen 5678 main.py
```

## Performance Monitoring

### Local Monitoring Stack

```bash
# Start monitoring services
docker-compose -f docker-compose.monitoring.yml up -d

# Access dashboards
# Grafana: http://localhost:3001 (admin/admin)
# Prometheus: http://localhost:9090
# Jaeger: http://localhost:16686
```

### Profiling

```bash
# Go profiling
go tool pprof http://localhost:8080/debug/pprof/profile

# Node.js profiling
npm run dev -- --inspect

# Python profiling
py-spy top --pid $(pgrep -f "python main.py")
```

## Troubleshooting

### Common Issues

**Port conflicts:**
```bash
# Kill processes on specific ports
sudo lsof -ti:3000 | xargs kill -9
sudo lsof -ti:8080 | xargs kill -9
```

**Docker issues:**
```bash
# Reset Docker environment
docker-compose down -v
docker system prune -a
docker-compose up -d
```

**Go module issues:**
```bash
cd backend
go clean -modcache
go mod download
```

**Node.js issues:**
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### Log Locations

- Backend logs: `backend/logs/`
- Frontend logs: Browser DevTools Console
- AI Services logs: `ai-services/logs/`
- Docker logs: `docker-compose logs [service]`

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### PR Requirements

- [ ] All tests passing
- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] Security review completed
- [ ] Performance impact assessed

## Architecture Decisions

See [ADRs](./architecture-decisions/) for architectural decision records and design rationale.

## API Documentation

- Backend API: http://localhost:8080/docs (Swagger)
- AI Services API: http://localhost:8000/docs (FastAPI docs)

## Support

- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Discord**: [Development Discord](https://discord.gg/self-dev)