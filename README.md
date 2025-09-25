# Self - Your Digital Memory Assistant

> An AI-powered personal knowledge management system that connects your conversations, files, calendar, and digital life into one intelligent workspace.

## Overview

Self is your external brain that records, understands, and connects everything you do. It transcribes your conversations, monitors your files, syncs with your calendar and email, then creates intelligent connections to help you never forget anything important.

## Key Features

### 🎤 Intelligent Audio Processing
- **Real-time transcription** of conversations and voice notes
- **Speaker identification** and conversation segmentation
- **Continuous background recording** with privacy controls
- **Voice commands** for hands-free interaction

### 🧠 Smart Knowledge Graph
- **Entity extraction** from conversations (people, projects, deadlines)
- **Relationship mapping** between topics across time
- **Context threading** - connects related discussions automatically
- **Timeline reconstruction** - see how ideas evolved

### 📁 File Intelligence
- **Activity monitoring** - tracks file creation, edits, and access
- **Content analysis** - understands what documents contain
- **Auto-tagging** based on conversation mentions
- **Smart organization** - suggests folder structures from your patterns

### 🔗 Universal Integration
- **Calendar sync** - connects meetings with conversations
- **Email integration** - links discussions to relevant threads
- **Cloud storage** - works with Google Drive, Dropbox, OneDrive
- **Communication tools** - Slack, Teams, Discord integration
- **Project management** - Notion, Trello, Asana connectivity

### 💡 Proactive Intelligence
- **Pattern recognition** - notices repeated mentions without action
- **Smart reminders** - "You mentioned calling Mike 3 times but haven't yet"
- **Context suggestions** - "Files changed in React project since last discussion"
- **Action tracking** - follows up on decisions and commitments

## Interface Components

### Main Dashboard
```
┌─────────────────────────────────────────────────────────┐
│  🎤 [Recording] Today: 4h 23m transcribed                │
├─────────────────────────────────────────────────────────┤
│  💬 Recent Conversations                                │
│  🔗 Smart Connections                                   │
│  📁 Recently Accessed Files                             │
│  📅 Today's Calendar Context                            │
│  💡 Proactive Insights                                  │
└─────────────────────────────────────────────────────────┘
```

### Chat Interface
Natural language queries like:
- "What did I decide about the budget?"
- "Show me everything about the mobile app project"
- "Create a summary of today's meetings"
- "Find that document I mentioned yesterday"

### Timeline View
Chronological view of:
- Conversations with timestamps
- File interactions
- Calendar events
- Email threads
- Cross-references and connections

## System Architecture

### Core System Components

#### 1. **Frontend Application** (Next.js/React)
- **Repository**: GitHub main repo `/frontend`
- **Hosting**: Vercel (auto-deploys from `main` branch)
- **Purpose**: User interface, dashboard, chat, timeline views
- **Tech Stack**: Next.js 14, TypeScript, Tailwind CSS, shadcn/ui

#### 2. **Backend API** (Node.js/Express)
- **Repository**: GitHub main repo `/backend`
- **Hosting**: Railway or Render (auto-deploys from `main`)
- **Purpose**: Authentication, data processing, integration orchestration
- **Tech Stack**: Node.js, Express, TypeScript, Prisma ORM

#### 3. **Database** (Supabase)
- **Service**: Supabase hosted PostgreSQL
- **Purpose**: User data, conversations, file metadata, relationships
- **Features**: Real-time subscriptions, RLS policies, vector embeddings

#### 4. **AI Processing Service** (Python/FastAPI)
- **Repository**: GitHub main repo `/ai-services`
- **Hosting**: Modal or Railway (GPU instances)
- **Purpose**: Speech-to-text, NLP, entity extraction, embeddings
- **Tech Stack**: Python, FastAPI, OpenAI Whisper, spaCy, sentence-transformers

#### 5. **Audio Storage** (AWS S3 or Supabase Storage)
- **Service**: S3 bucket with CloudFront CDN
- **Purpose**: Encrypted audio file storage with timestamps
- **Security**: Pre-signed URLs, client-side encryption

#### 6. **File Monitoring Service** (Electron/Tauri)
- **Repository**: GitHub main repo `/desktop-app`
- **Distribution**: GitHub Releases
- **Purpose**: Local file system monitoring, desktop integration
- **Tech Stack**: Tauri (Rust + Web frontend)

#### 7. **Background Processing** (Temporal.io or BullMQ)
- **Service**: Temporal Cloud or self-hosted Redis
- **Purpose**: Async transcription, integration syncs, scheduled tasks
- **Features**: Retry logic, workflow orchestration, cron jobs

#### 8. **Vector Database** (Pinecone or pgvector)
- **Service**: Pinecone hosted or Supabase pgvector extension
- **Purpose**: Semantic search, conversation similarity, content matching
- **Features**: High-dimensional embeddings, similarity queries

#### 9. **Integration Hub** (Microservices)
- **Repository**: GitHub main repo `/integrations`
- **Hosting**: Railway containers
- **Purpose**: OAuth handlers, webhook processors, data normalizers
- **Services**: Calendar sync, email processing, file watchers

#### 10. **Cache Layer** (Redis)
- **Service**: Upstash Redis or Railway Redis
- **Purpose**: Session storage, API rate limiting, temporary data
- **Features**: Real-time pub/sub, distributed caching

### Development Workflow

#### **Branch Strategy**
- **`main`** → Production deployments (auto-deploy to Vercel/Railway)
- **`develop`** → Your active development branch
- **`feature/*`** → Feature development branches

#### **Local Development**
```bash
git checkout develop
# Work on features
git checkout -b feature/audio-processing
# Development work
git merge feature/audio-processing → develop
# When ready for production
git merge develop → main (triggers deployments)
```

### System Connections & Data Flow

#### **Data Flow Architecture**

```
Desktop App (File Monitor) ←→ Backend API ←→ Supabase Database
                    ↓                ↓            ↓
Audio Files → S3 Storage    Background Jobs    Vector Store
                    ↓                ↓            ↓
            AI Processing ←→ Queue System ←→ Cache Layer
                    ↓                ↓            ↓
              Frontend App ←→ Integration Hub ←→ External APIs
```

#### **Connection Details**

**1. Desktop App ↔ Backend API**
- Desktop app monitors file changes, sends metadata to API
- WebSocket connection for real-time file activity
- Secure authentication with JWT tokens

**2. Frontend ↔ Backend API**
- Next.js API routes proxy to backend
- Real-time updates via Supabase subscriptions
- Authentication handled by Supabase Auth

**3. Backend API ↔ Supabase**
- Prisma ORM for database operations
- Real-time listeners for conversation updates
- Row Level Security for user data isolation

**4. Audio Processing Pipeline**
- Desktop app uploads audio to S3 with pre-signed URLs
- Background job triggers AI service for transcription
- Results stored in Supabase with vector embeddings in Pinecone

**5. Integration Hub ↔ External Services**
- OAuth flows handled by dedicated microservices
- Webhooks receive updates from calendar/email providers
- Data normalized and stored in Supabase

**6. Search & Intelligence**
- Vector search queries Pinecone for semantic similarity
- Cache layer (Redis) stores frequent queries
- AI service generates proactive insights asynchronously

#### **Real-time Data Synchronization**

**File Changes**: Desktop App → WebSocket → Backend → Supabase → Frontend (live updates)

**New Conversation**: Audio → S3 → Background Job → AI Processing → Supabase → Frontend (real-time transcription)

**Calendar Sync**: Google Calendar Webhook → Integration Hub → Supabase → Frontend (immediate calendar updates)

**Cross-References**: New entity detected → Vector search → Related content → Proactive insight → Frontend notification

#### **Security & Authentication Flow**

1. User authenticates via Supabase Auth (frontend)
2. JWT token validates API requests (backend)
3. Desktop app gets secure token for file uploads
4. All external integrations use OAuth 2.0
5. Data encrypted at rest (Supabase) and in transit (HTTPS)

#### **Deployment Pipeline**

1. **Code Push**: `develop` → `main` branch
2. **Triggers**: GitHub Actions workflows
3. **Parallel Deployments**:
   - Frontend → Vercel (automatic)
   - Backend API → Railway (automatic)
   - AI Services → Modal (automatic)
   - Desktop App → GitHub Releases (manual)
4. **Database Migrations**: Prisma migrate runs on deploy
5. **Environment Sync**: Production secrets managed via platform dashboards

---

## Open Source Architecture & Development Plan

### **Core Technology Stack**

#### **Backend Stack**
- **Language**: Go 1.21+ for high performance
- **Framework**: Fiber v2 (Express-like but blazing fast)
- **Database**: PostgreSQL 15+ with pgvector extension
- **Cache**: Redis 7+ for sessions and pub/sub
- **Queue**: NATS with JetStream for reliable messaging
- **Storage**: MinIO (S3-compatible, self-hosted)
- **Vector DB**: Qdrant for advanced similarity search

#### **AI/ML Stack**
- **Speech-to-Text**: Whisper.cpp (4-10x faster than Python)
- **Text-to-Speech**: EvenLabs API + Coqui TTS (open source)
- **LLM**: Ollama (local Llama 3.1, Mistral, CodeLlama)
- **Embeddings**: sentence-transformers (all-MiniLM-L6-v2)
- **NLP**: spaCy for entity extraction and parsing
- **Local API**: LocalAI for OpenAI-compatible endpoints

#### **Frontend Stack**
- **Web**: Next.js 14 with App Router + TypeScript
- **Desktop**: Tauri 2.0 + Rust (smaller, more secure)
- **Mobile**: React Native + Tauri Mobile
- **UI**: shadcn/ui + Tailwind CSS + Framer Motion
- **State**: Zustand + React Query for server state
- **Audio**: WebRTC for real-time streaming

### **Core System Components**

#### **1. Audio Processing Engine**
- **Local Transcription**: Whisper.cpp with GPU acceleration
- **Audio Streaming**: WebRTC for real-time, MinIO for storage
- **Format Support**: WAV, MP3, M4A, FLAC with conversion
- **Speaker Diarization**: pyannote.audio integration
- **Quality Control**: Confidence scoring and error detection

#### **2. Backend Core** (Go + Fiber)
- **API Server**: RESTful + GraphQL endpoints
- **WebSocket Handler**: Real-time transcription updates
- **Authentication**: JWT + refresh tokens, bcrypt hashing
- **Middleware**: CORS, rate limiting, logging, compression
- **File Processing**: Document analysis with Apache Tika

#### **3. Intelligence Engine**
- **Entity Extraction**: spaCy NLP pipeline
- **Semantic Search**: pgvector + Qdrant for complex queries
- **Pattern Recognition**: Custom algorithms + ML models
- **LLM Integration**: Ollama for local inference
- **Insight Generation**: Rule engine + prompt engineering

#### **4. Desktop Application** (Tauri + Rust)
- **Audio Capture**: cpal for cross-platform recording
- **File Monitoring**: fsnotify for real-time file watching
- **System Integration**: Tray icons, global hotkeys, notifications
- **Security**: Sandboxed execution with minimal permissions
- **Auto-updater**: Delta patches for efficient updates

#### **5. Integration Hub**
- **OAuth Provider**: Hydra for secure authentication
- **API Clients**: Custom Go clients for each service
- **Webhook Processing**: Reliable with retry mechanisms
- **Data Normalization**: Unified data models across services
- **Sync Engine**: Bidirectional sync with conflict resolution

#### **6. Storage & Files**
- **Object Storage**: MinIO with bucket policies
- **File Analysis**: Content extraction and tagging
- **Version Control**: Git integration with go-git
- **Cloud Sync**: rclone for multi-cloud support
- **Encryption**: AES-256 at rest, TLS 1.3 in transit

### **Development Phases**

#### **Phase 0: Foundation Architecture**
**Core Infrastructure Setup**

**Key Components:**
- Go backend with Fiber framework and middleware
- PostgreSQL with pgvector extension
- MinIO object storage with security policies
- Redis cache with pub/sub configuration
- Basic Tauri desktop app shell
- Next.js frontend with authentication flows
- Docker Compose development environment
- CI/CD pipelines with GitHub Actions

**Deliverables:**
- Database schema and migrations system
- JWT authentication with refresh tokens
- File upload/download with pre-signed URLs
- WebSocket connections for real-time updates
- Structured logging and basic monitoring

#### **Phase 1: Core Audio Pipeline**
**Record → Transcribe → Store → Search**

**Key Components:**
- Desktop audio recording with cpal
- Whisper.cpp integration for fast local transcription
- Audio processing pipeline with NATS messaging
- Speaker diarization for multi-person conversations
- File system monitoring with real-time events
- Basic conversation storage and threading

**Deliverables:**
- High-quality audio capture and format conversion
- Real-time transcription with confidence scoring
- Conversation segmentation by speaker and time gaps
- Simple text search across all transcriptions
- File change correlation with audio timeline
- WebSocket updates for live transcription display

#### **Phase 2: Intelligence Foundation**
**Entity Extraction & Basic Insights**

**Key Components:**
- spaCy NLP pipeline for entity recognition
- sentence-transformers for semantic embeddings
- pgvector integration for similarity search
- Basic relationship mapping between entities
- File content analysis and auto-tagging
- Simple insight generation rules

**Deliverables:**
- Named entity extraction (people, places, dates, projects)
- Semantic search across conversations and files
- Cross-reference detection between audio and documents
- Basic conversation summaries and key points
- File recommendations based on conversation content
- Simple pattern detection for repeated mentions

#### **Phase 3: Advanced Intelligence**
**Proactive Insights & Pattern Recognition**

**Key Components:**
- Ollama integration with local LLM models
- Advanced pattern recognition algorithms
- Action item detection and tracking
- Context window management for long conversations
- Smart notification system with configurable rules
- Temporal analysis for behavioral insights

**Deliverables:**
- Proactive reminders for unresolved mentions
- Intelligent conversation threading across time
- Custom insight generation with LLM prompting
- Action item tracking with follow-up suggestions
- Advanced search with natural language queries
- Behavioral pattern analysis and recommendations

#### **Phase 4: Integration Ecosystem**
**External Service Connections**

**Key Components:**
- OAuth2 server with Hydra for third-party auth
- Universal API client framework
- Real-time webhook processing system
- Multi-tenant data synchronization
- Cross-platform search and correlation
- Integration health monitoring

**Deliverables:**
- Google Calendar and Gmail integration
- Slack, Teams, and Discord connectors
- GitHub and GitLab repository analysis
- Cloud storage sync (Drive, Dropbox, OneDrive)
- Universal search across all connected data
- Real-time updates from external services

#### **Phase 5: Advanced Features & Scale**
**Polish, Performance & Team Collaboration**

**Key Components:**
- EvenLabs TTS for high-quality voice responses
- Advanced visualization with interactive timelines
- Team workspaces with shared intelligence
- Voice command processing with wake words
- Mobile companion app for on-the-go access
- Enterprise security and compliance features

**Deliverables:**
- Real-time collaborative features
- Advanced analytics dashboard
- Performance optimization and horizontal scaling
- Voice-controlled interface with natural commands
- Mobile app with core functionality
- Advanced privacy controls and data export

### **Technology Advantages**

#### **Performance Benefits**
- **Whisper.cpp**: 4-10x faster transcription than Python
- **Go Backend**: 10-50x better concurrency than Node.js
- **Local Processing**: No network latency for AI operations
- **pgvector**: Integrated vector search without API calls

#### **Cost Reduction**
- **~80% reduction** in AI processing costs (local vs cloud)
- **Self-hosted storage** eliminates expensive S3 bills
- **No SaaS subscriptions** for core AI functionality
- **Horizontal scaling** without vendor lock-in

#### **Privacy & Control**
- **Local-first processing** keeps sensitive data private
- **Full control** over AI models and data retention
- **Offline capability** after initial setup
- **Custom fine-tuning** for domain-specific needs

#### **Reliability & Scale**
- **No external API dependencies** for core features
- **Built-in redundancy** with message queues
- **Graceful degradation** when services are unavailable
- **Container-native** architecture for easy deployment

## Supported Integrations

### Communication
- **Email**: Gmail, Outlook, Apple Mail
- **Chat**: Slack, Teams, Discord, WhatsApp
- **Video**: Zoom, Google Meet, Teams meetings
- **Voice**: Phone calls, voice memos

### Productivity
- **Calendar**: Google Calendar, Outlook, Apple Calendar
- **Files**: Google Drive, Dropbox, OneDrive, iCloud
- **Notes**: Notion, Obsidian, Apple Notes, OneNote
- **Tasks**: Trello, Asana, Monday, Todoist

### Development
- **Code**: GitHub, GitLab, VSCode activity
- **Documentation**: Confluence, GitBook, wikis
- **Communication**: GitHub issues, code reviews

### Personal
- **Health**: Apple Health, fitness trackers
- **Finance**: Banking APIs, expense trackers
- **Social**: LinkedIn, Twitter (for work context)

## Getting Started

### Prerequisites
- Node.js 18+ or Python 3.9+
- 4GB RAM minimum (8GB recommended)
- Microphone access
- 10GB free storage

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/self.git
cd self

# Install dependencies
npm install
# or
pip install -r requirements.txt

# Configure integrations
npm run setup
# or
python setup.py configure

# Start the application
npm run dev
# or
python app.py
```

### Initial Setup

1. **Grant permissions** for microphone and file access
2. **Connect integrations** through OAuth flows
3. **Configure privacy settings** and recording preferences
4. **Train speaker recognition** with voice samples
5. **Import existing data** from connected services

## Usage Examples

### Voice Commands
```
"Show me what I said about the budget"
"Create a folder for the new mobile app project"
"Remind me to follow up with Sarah tomorrow"
"What meetings do I have related to the React project?"
```

### Chat Queries
```
"What were the key decisions from yesterday's standup?"
"Find all documents related to user authentication"
"Who did I promise to send the design files to?"
"Show me my conversation history with the development team"
```

### Proactive Insights
- Detects when you mention the same task multiple times without action
- Suggests file organization based on conversation topics
- Identifies scheduling conflicts between commitments and calendar
- Recommends following up on pending decisions

## Privacy & Security

### Data Handling
- **Local-first**: Core processing happens on your device
- **Encrypted storage**: All data encrypted at rest
- **Selective sync**: Choose what gets uploaded to cloud
- **Automatic deletion**: Configurable data retention policies

### Privacy Controls
- **Recording toggles**: Easy on/off for different contexts
- **Selective transcription**: Choose which conversations to process
- **Integration permissions**: Granular control over data access
- **Export/delete**: Full data portability and deletion rights

## Roadmap

### Phase 1 - Core System
- [x] Audio transcription pipeline
- [x] File monitoring system
- [x] Basic knowledge graph
- [ ] Calendar integration
- [ ] Email sync

### Phase 2 - Intelligence
- [ ] Advanced entity recognition
- [ ] Proactive insights engine
- [ ] Cross-platform search
- [ ] Voice command system

### Phase 3 - Integrations
- [ ] Communication platforms
- [ ] Project management tools
- [ ] Development workflows
- [ ] Mobile companion app

### Phase 4 - Advanced Features
- [ ] Team collaboration
- [ ] API for third-party integrations
- [ ] Advanced analytics
- [ ] Custom AI models

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Frontend (React/TypeScript)
cd frontend
npm install
npm run dev

# Backend (Node.js/Python)
cd backend
npm install
npm run dev:watch

# AI Services (Python)
cd ai-services
pip install -r requirements.txt
python main.py
```

## Support

- **Documentation**: [docs.self-app.com](https://docs.self-app.com)
- **Issues**: [GitHub Issues](https://github.com/yourusername/self/issues)
- **Discord**: [Community Server](https://discord.gg/self-app)
- **Email**: support@self-app.com

## License

MIT License - see [LICENSE](LICENSE) for details.

---

**Self** - Because your thoughts and work deserve perfect memory.