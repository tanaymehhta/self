# Database Schema - Supabase

This directory contains the database schema for the Self application using Supabase (PostgreSQL with extensions).

## Setup Instructions

### 1. Create Supabase Project

1. Go to [Supabase](https://supabase.com) and create a new project
2. Choose a project name: `self-app` (or similar)
3. Generate a strong database password
4. Select the region closest to you
5. Wait for the project to be created (~2 minutes)

### 2. Run Database Schema

1. In your Supabase dashboard, go to **SQL Editor**
2. Copy the contents of `schema.sql`
3. Paste and run the SQL to create all tables, indexes, and policies

### 3. Enable Required Extensions

The schema automatically enables these extensions:
- `uuid-ossp` - UUID generation
- `pgcrypto` - Encryption functions
- `vector` - Vector embeddings for semantic search

### 4. Get Connection Details

From your Supabase project dashboard:
- **API URL**: Project Settings → API → Project URL
- **Anon Key**: Project Settings → API → Project API keys → anon key
- **Service Role Key**: Project Settings → API → Project API keys → service_role key
- **Database URL**: Project Settings → Database → Connection string → URI

## Database Schema Overview

### Core Tables

#### Users (`public.users`)
- Extends Supabase's `auth.users` table
- Stores user preferences, storage quotas, and profile information
- Automatically created via trigger when user signs up

#### Conversations (`public.conversations`)
- Stores audio conversation metadata
- Links to audio files in Supabase Storage
- Tracks processing status and conversation statistics

#### Transcriptions (`public.transcriptions`)
- Individual speech segments with timestamps
- Speaker identification and confidence scores
- Links to parent conversation

#### File Events (`public.file_events`)
- File system activity monitoring
- File metadata and change tracking
- Optional correlation with conversations

#### Entities (`public.entities`)
- Named entities (people, places, projects, etc.)
- Extracted from conversations via NLP
- Tracks mention frequency and relationships

#### Entity Mentions (`public.entity_mentions`)
- Links entities to specific conversations/transcriptions
- Includes context and confidence scores
- Enables entity timeline tracking

#### Connections (`public.connections`)
- Relationships between different data types
- Weighted connections for importance
- Enables graph-based insights

#### Embeddings (`public.embeddings`)
- Vector embeddings for semantic search
- Uses pgvector extension for similarity search
- 384-dimensional vectors (sentence-transformers)

#### Insights (`public.insights`)
- Proactive suggestions and patterns
- Priority-based recommendation system
- User acknowledgment tracking

#### Integrations (`public.integrations`)
- External service connections
- Encrypted token storage
- Sync status and settings

### Security Features

#### Row Level Security (RLS)
- All tables have RLS enabled
- Users can only access their own data
- Policies enforce data isolation

#### Authentication Integration
- Automatic user profile creation
- JWT-based authentication via Supabase Auth
- Secure API access with user context

#### Vector Search
- Semantic similarity search function
- User-scoped vector queries
- Configurable similarity thresholds

## Environment Variables

Create these environment variables for your application:

```bash
# Supabase Configuration
NEXT_PUBLIC_SUPABASE_URL=your_supabase_url
NEXT_PUBLIC_SUPABASE_ANON_KEY=your_supabase_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_supabase_service_role_key
DATABASE_URL=your_supabase_database_url
```

## Data Flow

### User Registration
1. User signs up via Supabase Auth
2. Trigger creates profile in `public.users`
3. RLS policies activated for user data

### Conversation Processing
1. Audio uploaded to Supabase Storage
2. Conversation record created with `processing` status
3. AI services process audio and create transcriptions
4. Entities extracted and linked via mentions
5. Embeddings generated for semantic search
6. Connections created between related data

### Real-time Updates
- Supabase real-time subscriptions
- Live updates for transcription progress
- Instant insight notifications

## Backup and Maintenance

### Daily Backups
Supabase provides automated daily backups for paid plans.

### Manual Backup
```sql
-- Export user data
COPY (SELECT * FROM public.conversations WHERE user_id = 'user-uuid') TO '/tmp/conversations.csv' CSV HEADER;

-- Export embeddings
COPY (SELECT * FROM public.embeddings WHERE user_id = 'user-uuid') TO '/tmp/embeddings.csv' CSV HEADER;
```

### Database Maintenance
- Vacuum and analyze tables weekly
- Monitor query performance with `pg_stat_statements`
- Update vector index statistics monthly

## Scaling Considerations

### Performance Optimization
- Partitioning large tables by user_id or date
- Connection pooling via Supabase Pooler
- Read replicas for analytics queries

### Storage Management
- Archive old conversations to cold storage
- Compress embeddings for storage efficiency
- Implement data retention policies

### Vector Search Performance
- Tune `ivfflat` index parameters
- Consider `hnsw` index for higher accuracy
- Monitor vector query performance