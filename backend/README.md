# Self Backend

High-performance Go API server for the Self digital memory assistant.

## Features

- **Fast API**: Fiber framework for high-performance HTTP handling
- **Real-time**: WebSocket support for live updates
- **Database**: PostgreSQL with pgvector for vector operations
- **Caching**: Redis for sessions and temporary data
- **Message Queue**: NATS for reliable background processing
- **Object Storage**: MinIO integration for file storage
- **Authentication**: JWT with refresh token rotation

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Fiber v2
- **Database**: PostgreSQL 15+ with pgvector
- **Cache**: Redis 7+
- **Queue**: NATS with JetStream
- **Storage**: MinIO (S3-compatible)
- **Validation**: go-playground/validator
- **Logging**: structured logging with slog

## Getting Started

```bash
# Install dependencies
go mod tidy

# Run database migrations
go run cmd/migrate/main.go up

# Start development server
go run cmd/server/main.go

# Build for production
go build -o bin/server cmd/server/main.go
```

## Environment Variables

Create `.env`:

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
JWT_REFRESH_SECRET=your-refresh-secret-here

# Server
PORT=8080
ENVIRONMENT=development
```

## Project Structure

```
backend/
├── cmd/                   # Command line applications
│   ├── server/           # Main API server
│   └── migrate/          # Database migrations
├── internal/             # Private application code
│   ├── api/              # HTTP handlers
│   ├── auth/             # Authentication logic
│   ├── database/         # Database operations
│   ├── middleware/       # HTTP middleware
│   ├── models/           # Data models
│   ├── services/         # Business logic
│   └── websocket/        # WebSocket handlers
├── migrations/           # SQL migration files
├── pkg/                  # Public packages
│   ├── audio/            # Audio processing utilities
│   ├── config/           # Configuration management
│   └── logger/           # Logging utilities
└── tests/                # Test files
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/logout` - Logout user

### Audio
- `POST /api/v1/audio/upload` - Upload audio file
- `GET /api/v1/audio/:id` - Get audio file
- `POST /api/v1/audio/transcribe` - Start transcription
- `GET /api/v1/transcriptions/:id` - Get transcription

### Conversations
- `GET /api/v1/conversations` - List conversations
- `GET /api/v1/conversations/:id` - Get conversation
- `POST /api/v1/conversations` - Create conversation
- `PUT /api/v1/conversations/:id` - Update conversation

### Files
- `GET /api/v1/files` - List monitored files
- `GET /api/v1/files/:id` - Get file details
- `POST /api/v1/files/events` - Record file event

### Search
- `GET /api/v1/search` - Search conversations and files
- `POST /api/v1/search/semantic` - Semantic vector search

## Database Schema

### Core Tables
- `users` - User accounts and preferences
- `conversations` - Audio conversation records
- `transcriptions` - Speech-to-text results
- `file_events` - File system activity
- `entities` - Extracted entities (people, places, etc.)
- `connections` - Relationships between entities

### Vector Tables
- `conversation_embeddings` - Vector embeddings for semantic search
- `file_embeddings` - Document content embeddings

## WebSocket Events

### Client to Server
- `join_room` - Join conversation room
- `start_recording` - Begin audio recording
- `audio_chunk` - Send audio data
- `stop_recording` - End audio recording

### Server to Client
- `transcription_update` - Real-time transcription
- `conversation_update` - Conversation metadata update
- `file_event` - File system activity
- `insight_generated` - New proactive insight

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/services/...

# Run with race detection
go test -race ./...
```

### Database Operations

```bash
# Create new migration
migrate create -ext sql -dir migrations -seq add_conversations_table

# Run migrations
go run cmd/migrate/main.go up

# Rollback migrations
go run cmd/migrate/main.go down 1

# Reset database
go run cmd/migrate/main.go drop
go run cmd/migrate/main.go up
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Check for vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Performance

### Optimizations
- Connection pooling for database
- Redis caching for frequent queries
- Background processing with NATS
- Efficient JSON serialization
- HTTP/2 support with Fiber

### Monitoring
- Prometheus metrics at `/metrics`
- Health checks at `/health`
- Profiling endpoints at `/debug/pprof/`

## Deployment

### Docker

```bash
# Build image
docker build -t self-backend .

# Run container
docker run -p 8080:8080 --env-file .env self-backend
```

### Production Checklist

- [ ] Set strong JWT secrets
- [ ] Configure HTTPS/TLS
- [ ] Set up monitoring and alerting
- [ ] Configure log rotation
- [ ] Set resource limits
- [ ] Enable CORS properly
- [ ] Set up backup procedures