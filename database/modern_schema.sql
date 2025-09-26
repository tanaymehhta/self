-- Self - Modern Database Schema (Based on Best Practices)
-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "vector";

-- Users table (simplified single-user focus)
CREATE TABLE public.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    avatar_url TEXT,
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Content items (any type of content)
CREATE TABLE public.content_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    content_type TEXT NOT NULL, -- 'document', 'audio', 'image', 'web'
    title TEXT,
    file_path TEXT,
    file_size BIGINT,
    checksum TEXT,
    language TEXT DEFAULT 'en',
    source_metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Chunked content (retrievable units)
CREATE TABLE public.chunks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_item_id UUID NOT NULL REFERENCES public.content_items(id) ON DELETE CASCADE,
    chunk_text TEXT NOT NULL,
    chunk_index INTEGER NOT NULL,
    token_count INTEGER,
    chunk_span JSONB DEFAULT '{}', -- {page: 3, start_time: 45.2, etc}
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Modern embeddings table (model-agnostic)
CREATE TABLE public.embeddings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chunk_id UUID NOT NULL REFERENCES public.chunks(id) ON DELETE CASCADE,
    embedding_model TEXT NOT NULL, -- 'text-embedding-3-small', 'text-embedding-3-large'
    embedding_dim INTEGER NOT NULL, -- 1536, 3072, etc
    embedding VECTOR(3072), -- Use max dim you plan to support
    embedding_version INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Speakers table (for audio content)
CREATE TABLE public.speakers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    voice_signature JSONB DEFAULT '{}', -- Audio characteristics
    confidence_threshold FLOAT DEFAULT 0.8,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audio segments (who said what when)
CREATE TABLE public.audio_segments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_item_id UUID NOT NULL REFERENCES public.content_items(id) ON DELETE CASCADE,
    speaker_id UUID REFERENCES public.speakers(id),
    text TEXT NOT NULL,
    start_time DECIMAL(10, 3), -- seconds with millisecond precision
    end_time DECIMAL(10, 3),
    confidence DECIMAL(3, 2), -- 0.00 to 1.00
    needs_review BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Modern indexes (HNSW for better performance)
CREATE INDEX ON public.embeddings USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Full-text search index
CREATE INDEX ON public.chunks USING gin (to_tsvector('english', chunk_text));

-- Performance indexes
CREATE INDEX idx_content_items_type ON public.content_items(content_type);
CREATE INDEX idx_chunks_content_item ON public.chunks(content_item_id);
CREATE INDEX idx_embeddings_model ON public.embeddings(embedding_model);
CREATE INDEX idx_embeddings_chunk ON public.embeddings(chunk_id);
CREATE INDEX idx_audio_segments_speaker ON public.audio_segments(speaker_id);
CREATE INDEX idx_audio_segments_time ON public.audio_segments(start_time, end_time);

-- Row Level Security (RLS) policies
ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.content_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.chunks ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.embeddings ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.speakers ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.audio_segments ENABLE ROW LEVEL SECURITY;

-- RLS Policies (single user - all data belongs to authenticated user)
CREATE POLICY "Users can access own data" ON public.content_items
    FOR ALL USING (user_id = auth.uid());

CREATE POLICY "Users can access own chunks" ON public.chunks
    FOR ALL USING (content_item_id IN (
        SELECT id FROM public.content_items WHERE user_id = auth.uid()
    ));

CREATE POLICY "Users can access own embeddings" ON public.embeddings
    FOR ALL USING (chunk_id IN (
        SELECT c.id FROM public.chunks c
        JOIN public.content_items ci ON c.content_item_id = ci.id
        WHERE ci.user_id = auth.uid()
    ));

CREATE POLICY "Users can access own speakers" ON public.speakers
    FOR ALL USING (user_id = auth.uid());

CREATE POLICY "Users can access own audio segments" ON public.audio_segments
    FOR ALL USING (content_item_id IN (
        SELECT id FROM public.content_items WHERE user_id = auth.uid()
    ));

-- Insert test user (for local development)
INSERT INTO public.users (id, email, password_hash, full_name) VALUES
(
    '2dec7597-f49f-42ad-ae54-d422ef0bd143',
    'test@example.com',
    '$2a$10$Nh8NQjKZzTlj9Hd6xjJGxu8z8d6KvWl9J5f2fF5cC1b0wQ4xR4xR4x',
    'Test User'
)
ON CONFLICT (email) DO NOTHING;