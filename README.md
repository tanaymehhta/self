# Self - Your Personal Digital Memory Assistant

> A sophisticated AI-powered personal knowledge management system that records, understands, and connects everything in your digital life into one intelligent workspace.

## Overview

Self is your personal external brain - a single-user AI knowledge management system that records, understands, and connects everything you do. It transcribes your conversations with speaker identification, processes your documents, analyzes your images, and creates intelligent cross-modal connections to help you never forget anything important.

### Personal & Private
Self is designed as a **personal system for individual use**. All conversations, documents, and data belong to you alone. The system learns your communication patterns, recognizes voices in your conversations, and builds a comprehensive knowledge graph of your digital life with complete privacy.

## Core Capabilities

### 🎤 Intelligent Audio Processing
- **Real-time transcription** with Whisper.cpp integration
- **Speaker identification and voice signatures** - automatically identifies who is speaking
- **Emotional tone analysis** - understands confidence, concern, excitement in speech
- **Conversation threading** - connects related discussions across time
- **Manual speaker tagging** - easy interface to tag unknown voices

### 🧠 Multi-Modal Content Understanding
- **Document processing** - PDFs, EPUBs, Word docs with structure preservation
- **Image analysis** - OCR text extraction + AI visual understanding
- **Video processing** - combines audio transcription with visual analysis
- **Web content** - saves and processes articles, research, bookmarks
- **Cross-modal search** - "budget discussions" finds audio, documents, and whiteboard photos

### 💡 Personal AI Assistant
- **Natural language queries** - "What did Sarah say about the marketing budget?"
- **Context-rich answers** - includes exact timestamps, emotional tone, related documents
- **Pattern recognition** - notices recurring topics and unresolved discussions
- **Smart connections** - automatically links conversations to related documents and images
- **Timeline reconstruction** - see how ideas and projects evolved over time

## System Architecture

### The 5-Component Architecture

#### 1. **Input Processing Layer** (Multi-Pipeline Ingestion)
Different content types require specialized processing pipelines:

**Audio Pipeline**
```
Audio File → Whisper Transcription → Speaker Identification →
Emotion Analysis → Audio+Text Embeddings → Storage
```

**Document Pipeline**
```
PDF/EPUB → Structure-Aware Text Extraction → Section Analysis →
Document+Context Embeddings → Storage
```

**Image Pipeline**
```
Image → OCR Text Extraction → AI Visual Description →
Image+Text Embeddings → Storage
```

**Video Pipeline**
```
Video → Audio Track + Visual Frames → Combined Processing →
Audio+Visual Embeddings → Storage
```

#### 2. **Hybrid Embedding Strategy**
Self uses a sophisticated multi-layered approach to preserve context while enabling powerful search:

**Universal Text Embeddings** - For broad, cross-modal search
- All content converted to text embeddings in single space
- Enables queries like "tell me about budgets" across all content types
- Fast similarity search using vector operations

**Source-Specific Embeddings** - For contextual search
- **Audio embeddings**: Include speaker characteristics, emotional tone, timing context
- **Document embeddings**: Preserve document structure, page context, section relationships
- **Visual embeddings**: Combine image descriptions with OCR text understanding
- **Video embeddings**: Merge audio transcription with visual scene analysis

#### 3. **Storage Architecture** (PostgreSQL + pgvector)
**Why PostgreSQL over Pinecone:**
- **Cost-effective**: No $70/month vector database fees
- **Privacy**: Your data stays local or in your cloud
- **Simplicity**: Single database for all data types
- **Integration**: Native SQL queries with vector search
- **Performance**: Fast enough for personal-scale data

```sql
-- Core tables
content_items: {id, type, file_path, created_at}
speakers: {id, name, voice_signature, confidence_threshold}
conversation_segments: {id, conversation_id, speaker_id, text, start_time, end_time}

-- Multi-layered embeddings
universal_embeddings: {content_id, embedding, text_content}
audio_embeddings: {content_id, speaker_id, audio_features, emotion_vector}
document_embeddings: {content_id, page_number, document_structure}
visual_embeddings: {content_id, image_description, ocr_text}
```

#### 4. **Search Orchestration** (Intelligent Query Router)
Smart system that analyzes your query and routes to appropriate search strategies:

- **Semantic Search**: "What did we discuss about budgets?" → Vector similarity
- **Speaker-Specific**: "What did John say about X?" → Audio embeddings + speaker filter
- **Temporal Search**: "Last week's discussions" → Time-based filtering
- **Source-Specific**: "Show me budget documents" → Document type filtering
- **Cross-Modal**: "Everything about Project Alpha" → Searches all content types

#### 5. **Response Assembly** (Context-Rich Results)
Search results include full source attribution and context:

```json
{
  "query": "What did Sarah say about being overwhelmed?",
  "results": [
    {
      "source_type": "audio",
      "text": "I'm feeling a bit overwhelmed with the Q4 deadlines",
      "speaker": "Sarah",
      "emotional_tone": "concerned",
      "conversation_title": "Team Standup",
      "timestamp": "2:34.5 - 2:41.2",
      "date": "2024-03-15T10:30:00Z"
    },
    {
      "source_type": "document",
      "text": "Team workload analysis shows high stress indicators",
      "document_title": "Team_Health_Report.pdf",
      "page": 3,
      "section": "Workload Analysis"
    }
  ]
}
```

## Tech Stack

### Backend (Go + Fiber)
- **Go**: High-performance, concurrent processing
- **Fiber**: Fast HTTP framework for API endpoints
- **GORM**: Database ORM with PostgreSQL support
- **JWT**: Secure authentication (single-user focused)
- **Whisper.cpp**: Local speech-to-text processing

### Database (PostgreSQL + pgvector)
- **PostgreSQL 16**: Primary database with JSONB support
- **pgvector extension**: Vector similarity search capabilities
- **Local deployment**: Docker container for development
- **Supabase ready**: Easy migration to hosted PostgreSQL

### Frontend (Next.js)
- **Next.js 14**: React framework with App Router
- **TypeScript**: Type-safe development
- **Tailwind CSS**: Utility-first styling
- **Local auth service**: Custom authentication system

### AI/ML Stack
- **Whisper.cpp**: Open source speech recognition
- **OpenAI embeddings**: text-embedding-ada-002 or local alternatives
- **Sentence Transformers**: Local embedding models option
- **spaCy/NLTK**: Natural language processing

## Current Status

### ✅ **Completed: Foundation Architecture**
- ✅ **Local PostgreSQL database** with pgvector extension running
- ✅ **Go backend server** with authentication and API endpoints
- ✅ **Database schema** with speakers, conversations, and embedding tables
- ✅ **JWT authentication system** working end-to-end
- ✅ **Next.js frontend** with login/register pages
- ✅ **Docker infrastructure** with Redis, MinIO, NATS, Qdrant services
- ✅ **Local development environment** fully operational

### 🚧 **Next Immediate Steps** (Current Development Phase)

#### **Step 1: Audio Pipeline Implementation** (2-3 weeks)
```
Priority: HIGH - Core functionality
```
1. **Whisper.cpp Integration**
   - Set up Whisper.cpp Go bindings
   - Implement audio file upload and processing
   - Create transcription with timestamps

2. **Speaker Identification System**
   - Build voice signature extraction
   - Create speaker detection algorithms
   - Design manual speaker tagging UI

3. **Audio Embeddings**
   - Generate embeddings for audio segments
   - Store with speaker and timing context
   - Test basic audio search functionality

#### **Step 2: Search Foundation** (1-2 weeks)
```
Priority: HIGH - Essential for user experience
```
1. **Vector Search Setup**
   - Configure pgvector indexes
   - Implement similarity search queries
   - Create embedding generation service

2. **Basic Query Interface**
   - Build simple search API endpoints
   - Create frontend search interface
   - Test with audio transcriptions

#### **Step 3: Document Processing** (2-3 weeks)
```
Priority: MEDIUM - Expands functionality
```
1. **PDF/Document Parser**
   - Implement text extraction with structure
   - Create document embeddings
   - Build document upload interface

2. **Cross-Modal Search**
   - Connect audio and document searches
   - Implement unified search results
   - Add timeline view for related content

### 🔮 **Future Development Phases**

#### **Phase 2: Advanced Features** (4-6 weeks)
- Image/OCR processing pipeline
- Video content analysis
- Smart notifications and insights
- Conversation pattern analysis

#### **Phase 3: Desktop Integration** (6-8 weeks)
- Desktop app for continuous recording
- File system monitoring
- Calendar and email integration
- Cross-device synchronization

#### **Phase 4: AI Intelligence** (Ongoing)
- Advanced conversation insights
- Proactive suggestions and reminders
- Personal communication pattern analysis
- Smart content recommendations

## Development Workflow

### Getting Started
```bash
# Start infrastructure services
docker-compose up -d

# Start backend (Terminal 1)
cd backend
go run cmd/server/main.go

# Start frontend (Terminal 2)
cd frontend
npm run dev

# Access application
# Frontend: http://localhost:3001
# Backend API: http://localhost:8080
# Database: localhost:5432
```

### Testing the System
```bash
# Test authentication
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'

# Check health
curl http://localhost:8080/health
```

---

## **Latest Architecture Plan: QA-Based Multi-Modal Search**

### **Revolutionary Approach: Answer Extraction over Chunk Comparison**

After deep analysis of the multi-modal relevance problem (audio vs text content having different information densities), we've pivoted to a **Question-Answering based retrieval system** that solves the fundamental comparison issue by extracting answers instead of comparing chunks.

#### **The Problem We Solved**
- Audio transcripts are verbose and conversational
- Text documents are concise and information-dense
- Traditional relevance scoring unfairly favors verbose audio content
- Users want **answers**, not chunks

#### **The Solution: Two-Stage Pipeline**

**Stage 1: Enhanced Retrieval** (builds on existing hybrid search)
- Vector + fulltext search finds top 10-20 candidate chunks
- Mixed audio transcript chunks and text document chunks

**Stage 2: Answer Extraction** (NEW - the game changer)
- LLM processes each chunk with the specific user query
- Extracts precise answer + confidence score from each chunk
- Returns "No relevant answer" for irrelevant chunks
- Ranks extracted **answers** by quality, not source verbosity

#### **Implementation Architecture**

```go
type AnswerResult struct {
    Answer       string  `json:"answer"`          // Extracted answer text
    Confidence   float64 `json:"confidence"`      // LLM confidence score
    SourceChunk  string  `json:"source_chunk"`    // Original chunk for context
    ChunkID      string  `json:"chunk_id"`        // Database reference
    SourceTitle  string  `json:"source_title"`    // Document/audio title
    ContentType  string  `json:"content_type"`    // "audio", "document", etc.
    HasAnswer    bool    `json:"has_answer"`      // True if chunk contains answer
    // Audio-specific metadata
    StartTime    *float64 `json:"start_time,omitempty"`    // Audio timestamp
    EndTime      *float64 `json:"end_time,omitempty"`      // Audio timestamp
    Speaker      *string  `json:"speaker,omitempty"`       // Speaker identification
}
```

#### **Phase Implementation Plan**

**Phase 1: Text QA Foundation** (1-2 weeks)
1. ✅ Enhanced chunk retrieval (existing hybrid search)
2. 🚧 AnswerExtractionService - LLM integration for answer extraction
3. 🚧 Answer confidence scoring and ranking system
4. 🚧 Text-based QA testing with documents and EPUBs

**Phase 2: Audio Extension** (1-2 weeks)
1. 🚧 Audio transcription service (Whisper integration)
2. 🚧 Transcript chunking with timestamp metadata
3. 🚧 Audio QA pipeline (transcript chunks → same answer extraction)
4. 🚧 Enhanced attribution with timing and speaker info

**Phase 3: Unified Answer System** (1 week)
1. 🚧 Cross-modal answer deduplication
2. 🚧 Rich source attribution (timestamps for audio, pages for documents)
3. 🚧 Answer quality thresholding and filtering
4. 🚧 Frontend answer presentation with expandable source context

#### **Why This Architecture Wins**
- **Solves core problem**: No more unfair audio vs text comparisons
- **Answer quality focus**: Best answer wins, regardless of source verbosity
- **Natural normalization**: Verbose audio and concise text produce equivalent answers
- **Preserves context**: Full chunks available when users need details
- **Query-specific**: Each answer tailored to the exact question asked
- **Computationally efficient**: Only processes top retrieved chunks

**Example Query Flow:**
```
Query: "What is the capital of France?"

Audio chunk: [500 words of conversation mentioning Paris casually]
→ Answer: "Paris is the capital of France" (confidence: 0.92)

Text chunk: [One precise sentence about Paris]
→ Answer: "Paris is the capital of France" (confidence: 0.98)

Result: Both produce equivalent answers, ranked by confidence
```

## **Immediate Next Steps Summary**

**Current Priority: Implement QA-Based Search Pipeline**

1. **Design AnswerExtractionService** - LLM integration for chunk → answer processing
2. **Enhance SearchService** - Integrate answer extraction into existing search flow
3. **Test Text QA Pipeline** - Validate approach with document content
4. **Implement Audio Transcription** - Whisper service integration
5. **Extend QA to Audio** - Apply answer extraction to transcript chunks
6. **Unified Answer Ranking** - Cross-modal answer quality comparison
7. **Rich Source Attribution** - Timestamps, speakers, page numbers
8. **Frontend Answer Interface** - Clean answer presentation with source context

**The Breakthrough:** We're not building a better search engine - we're building a personal answer engine that happens to use search for retrieval.

The foundation is solid - now we build the intelligent answer extraction system that makes Self truly understand and respond to your questions across all your content!