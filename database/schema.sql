-- Self - Database Schema
-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "vector";

-- Users table (extends Supabase auth.users)
CREATE TABLE public.users (
    id UUID REFERENCES auth.users(id) PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    avatar_url TEXT,
    preferences JSONB DEFAULT '{}',
    storage_quota BIGINT DEFAULT 10737418240, -- 10GB in bytes
    storage_used BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Conversations table
CREATE TABLE public.conversations (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    title VARCHAR(500),
    summary TEXT,
    duration_seconds INTEGER DEFAULT 0,
    word_count INTEGER DEFAULT 0,
    speaker_count INTEGER DEFAULT 1,
    audio_file_url TEXT,
    audio_file_size BIGINT,
    audio_format VARCHAR(10) DEFAULT 'wav',
    status VARCHAR(20) DEFAULT 'processing', -- processing, completed, failed
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Transcriptions table
CREATE TABLE public.transcriptions (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES public.conversations(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    start_time DECIMAL(10, 3), -- seconds with millisecond precision
    end_time DECIMAL(10, 3),
    speaker_id VARCHAR(50),
    confidence DECIMAL(3, 2), -- 0.00 to 1.00
    language VARCHAR(5) DEFAULT 'en',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- File events table
CREATE TABLE public.file_events (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    file_name VARCHAR(500) NOT NULL,
    file_size BIGINT,
    file_type VARCHAR(100),
    event_type VARCHAR(20) NOT NULL, -- created, modified, deleted, accessed
    hash_sha256 VARCHAR(64),
    conversation_id UUID REFERENCES public.conversations(id),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Entities table (people, places, projects, etc.)
CREATE TABLE public.entities (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    name VARCHAR(500) NOT NULL,
    type VARCHAR(50) NOT NULL, -- person, organization, location, project, etc.
    description TEXT,
    metadata JSONB DEFAULT '{}',
    frequency INTEGER DEFAULT 1, -- how often mentioned
    first_mentioned TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_mentioned TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Entity mentions (links entities to conversations/transcriptions)
CREATE TABLE public.entity_mentions (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    entity_id UUID NOT NULL REFERENCES public.entities(id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES public.conversations(id) ON DELETE CASCADE,
    transcription_id UUID REFERENCES public.transcriptions(id) ON DELETE CASCADE,
    context TEXT, -- surrounding text where entity was mentioned
    confidence DECIMAL(3, 2) DEFAULT 1.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Connections table (relationships between entities, conversations, files)
CREATE TABLE public.connections (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    source_type VARCHAR(20) NOT NULL, -- conversation, file, entity
    source_id UUID NOT NULL,
    target_type VARCHAR(20) NOT NULL,
    target_id UUID NOT NULL,
    relationship_type VARCHAR(50), -- mentioned_in, related_to, references, etc.
    strength DECIMAL(3, 2) DEFAULT 0.5, -- relationship strength 0.00 to 1.00
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(source_type, source_id, target_type, target_id, relationship_type)
);

-- Embeddings table for vector search
CREATE TABLE public.embeddings (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    content_type VARCHAR(20) NOT NULL, -- conversation, transcription, file
    content_id UUID NOT NULL,
    content_text TEXT NOT NULL,
    embedding VECTOR(384), -- sentence-transformers all-MiniLM-L6-v2 dimension
    model_name VARCHAR(100) DEFAULT 'all-MiniLM-L6-v2',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insights table for proactive suggestions
CREATE TABLE public.insights (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- repeated_mention, action_item, pattern, etc.
    title VARCHAR(500) NOT NULL,
    description TEXT,
    data JSONB NOT NULL, -- structured data about the insight
    priority INTEGER DEFAULT 50, -- 1-100, higher = more important
    status VARCHAR(20) DEFAULT 'new', -- new, acknowledged, dismissed
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Integrations table for external services
CREATE TABLE public.integrations (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    service_name VARCHAR(50) NOT NULL, -- gmail, calendar, slack, etc.
    service_user_id VARCHAR(255), -- user ID in external service
    access_token TEXT, -- encrypted
    refresh_token TEXT, -- encrypted
    token_expires_at TIMESTAMP WITH TIME ZONE,
    scopes TEXT[], -- requested permissions
    settings JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    last_sync TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, service_name)
);

-- Create indexes for performance
CREATE INDEX idx_conversations_user_id ON public.conversations(user_id);
CREATE INDEX idx_conversations_created_at ON public.conversations(created_at DESC);
CREATE INDEX idx_transcriptions_conversation_id ON public.transcriptions(conversation_id);
CREATE INDEX idx_transcriptions_start_time ON public.transcriptions(start_time);
CREATE INDEX idx_file_events_user_id ON public.file_events(user_id);
CREATE INDEX idx_file_events_created_at ON public.file_events(created_at DESC);
CREATE INDEX idx_file_events_file_path ON public.file_events(file_path);
CREATE INDEX idx_entities_user_id ON public.entities(user_id);
CREATE INDEX idx_entities_type ON public.entities(type);
CREATE INDEX idx_entities_name ON public.entities(name);
CREATE INDEX idx_entity_mentions_entity_id ON public.entity_mentions(entity_id);
CREATE INDEX idx_entity_mentions_conversation_id ON public.entity_mentions(conversation_id);
CREATE INDEX idx_connections_user_id ON public.connections(user_id);
CREATE INDEX idx_connections_source ON public.connections(source_type, source_id);
CREATE INDEX idx_connections_target ON public.connections(target_type, target_id);
CREATE INDEX idx_embeddings_user_id ON public.embeddings(user_id);
CREATE INDEX idx_embeddings_content ON public.embeddings(content_type, content_id);
CREATE INDEX idx_insights_user_id ON public.insights(user_id);
CREATE INDEX idx_insights_status ON public.insights(status);
CREATE INDEX idx_insights_created_at ON public.insights(created_at DESC);
CREATE INDEX idx_integrations_user_id ON public.integrations(user_id);
CREATE INDEX idx_integrations_service ON public.integrations(service_name);

-- Vector similarity search index
CREATE INDEX ON public.embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Row Level Security (RLS) policies
ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.conversations ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.transcriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.file_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.entities ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.entity_mentions ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.connections ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.embeddings ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.insights ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.integrations ENABLE ROW LEVEL SECURITY;

-- RLS Policies
CREATE POLICY "Users can view own profile" ON public.users FOR SELECT USING (auth.uid() = id);
CREATE POLICY "Users can update own profile" ON public.users FOR UPDATE USING (auth.uid() = id);

CREATE POLICY "Users can view own conversations" ON public.conversations FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "Users can insert own conversations" ON public.conversations FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "Users can update own conversations" ON public.conversations FOR UPDATE USING (auth.uid() = user_id);
CREATE POLICY "Users can delete own conversations" ON public.conversations FOR DELETE USING (auth.uid() = user_id);

CREATE POLICY "Users can view own transcriptions" ON public.transcriptions FOR SELECT USING (
    auth.uid() IN (SELECT user_id FROM public.conversations WHERE id = conversation_id)
);
CREATE POLICY "Users can insert own transcriptions" ON public.transcriptions FOR INSERT WITH CHECK (
    auth.uid() IN (SELECT user_id FROM public.conversations WHERE id = conversation_id)
);
CREATE POLICY "Users can update own transcriptions" ON public.transcriptions FOR UPDATE USING (
    auth.uid() IN (SELECT user_id FROM public.conversations WHERE id = conversation_id)
);
CREATE POLICY "Users can delete own transcriptions" ON public.transcriptions FOR DELETE USING (
    auth.uid() IN (SELECT user_id FROM public.conversations WHERE id = conversation_id)
);

CREATE POLICY "Users can view own file events" ON public.file_events FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "Users can insert own file events" ON public.file_events FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "Users can update own file events" ON public.file_events FOR UPDATE USING (auth.uid() = user_id);
CREATE POLICY "Users can delete own file events" ON public.file_events FOR DELETE USING (auth.uid() = user_id);

CREATE POLICY "Users can view own entities" ON public.entities FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "Users can insert own entities" ON public.entities FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "Users can update own entities" ON public.entities FOR UPDATE USING (auth.uid() = user_id);
CREATE POLICY "Users can delete own entities" ON public.entities FOR DELETE USING (auth.uid() = user_id);

CREATE POLICY "Users can view own entity mentions" ON public.entity_mentions FOR SELECT USING (
    auth.uid() IN (SELECT user_id FROM public.entities WHERE id = entity_id)
);
CREATE POLICY "Users can insert own entity mentions" ON public.entity_mentions FOR INSERT WITH CHECK (
    auth.uid() IN (SELECT user_id FROM public.entities WHERE id = entity_id)
);

CREATE POLICY "Users can view own connections" ON public.connections FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "Users can insert own connections" ON public.connections FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "Users can update own connections" ON public.connections FOR UPDATE USING (auth.uid() = user_id);
CREATE POLICY "Users can delete own connections" ON public.connections FOR DELETE USING (auth.uid() = user_id);

CREATE POLICY "Users can view own embeddings" ON public.embeddings FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "Users can insert own embeddings" ON public.embeddings FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "Users can update own embeddings" ON public.embeddings FOR UPDATE USING (auth.uid() = user_id);
CREATE POLICY "Users can delete own embeddings" ON public.embeddings FOR DELETE USING (auth.uid() = user_id);

CREATE POLICY "Users can view own insights" ON public.insights FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "Users can insert own insights" ON public.insights FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "Users can update own insights" ON public.insights FOR UPDATE USING (auth.uid() = user_id);
CREATE POLICY "Users can delete own insights" ON public.insights FOR DELETE USING (auth.uid() = user_id);

CREATE POLICY "Users can view own integrations" ON public.integrations FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "Users can insert own integrations" ON public.integrations FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "Users can update own integrations" ON public.integrations FOR UPDATE USING (auth.uid() = user_id);
CREATE POLICY "Users can delete own integrations" ON public.integrations FOR DELETE USING (auth.uid() = user_id);

-- Functions for automatic timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add update triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON public.conversations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_entities_updated_at BEFORE UPDATE ON public.entities FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_insights_updated_at BEFORE UPDATE ON public.insights FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_integrations_updated_at BEFORE UPDATE ON public.integrations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to create user profile after signup
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO public.users (id, email, full_name, avatar_url)
    VALUES (
        NEW.id,
        NEW.email,
        COALESCE(NEW.raw_user_meta_data->>'full_name', NEW.raw_user_meta_data->>'name', split_part(NEW.email, '@', 1)),
        NEW.raw_user_meta_data->>'avatar_url'
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Trigger to create user profile on signup
CREATE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- Vector similarity search function
CREATE OR REPLACE FUNCTION match_embeddings(
    query_embedding VECTOR(384),
    match_threshold FLOAT DEFAULT 0.5,
    match_count INTEGER DEFAULT 10
)
RETURNS TABLE (
    id UUID,
    content_type TEXT,
    content_id UUID,
    content_text TEXT,
    similarity FLOAT
)
LANGUAGE SQL STABLE
AS $$
    SELECT
        embeddings.id,
        embeddings.content_type,
        embeddings.content_id,
        embeddings.content_text,
        1 - (embeddings.embedding <=> query_embedding) AS similarity
    FROM embeddings
    WHERE auth.uid() = embeddings.user_id
    AND 1 - (embeddings.embedding <=> query_embedding) > match_threshold
    ORDER BY embeddings.embedding <=> query_embedding
    LIMIT match_count;
$$;