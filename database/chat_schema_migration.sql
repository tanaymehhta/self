-- Chat System Migration - Add Chat Tables to Existing Schema
-- This adds conversational chat functionality on top of the existing QA pipeline

-- Chat conversations (chat sessions)
CREATE TABLE IF NOT EXISTS public.chat_conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    title TEXT, -- Auto-generated or user-set conversation title
    message_count INTEGER DEFAULT 0,
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Chat messages (user + AI messages)
CREATE TABLE IF NOT EXISTS public.chat_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES public.chat_conversations(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    sources JSONB DEFAULT '[]', -- Array of source chunks that contributed to answer
    confidence DECIMAL(3, 2), -- AI response confidence (0.00 to 1.00)
    metadata JSONB DEFAULT '{}', -- Additional message metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Link conversations to documents being discussed
CREATE TABLE IF NOT EXISTS public.conversation_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES public.chat_conversations(id) ON DELETE CASCADE,
    content_item_id UUID NOT NULL REFERENCES public.content_items(id) ON DELETE CASCADE,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(conversation_id, content_item_id)
);

-- Performance indexes for chat
CREATE INDEX IF NOT EXISTS idx_chat_conversations_user ON public.chat_conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_conversations_activity ON public.chat_conversations(last_activity DESC);
CREATE INDEX IF NOT EXISTS idx_chat_messages_conversation ON public.chat_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created ON public.chat_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_conversation_documents_conv ON public.conversation_documents(conversation_id);
CREATE INDEX IF NOT EXISTS idx_conversation_documents_content ON public.conversation_documents(content_item_id);

-- Row Level Security for chat tables
ALTER TABLE public.chat_conversations ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.chat_messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.conversation_documents ENABLE ROW LEVEL SECURITY;

-- RLS Policies for chat (single user - all data belongs to authenticated user)
-- Note: These policies assume Supabase auth.uid() function exists
-- For local development without Supabase auth, RLS will be disabled
CREATE POLICY "Users can access own conversations" ON public.chat_conversations
    FOR ALL USING (true); -- Allow all access for local development

CREATE POLICY "Users can access own messages" ON public.chat_messages
    FOR ALL USING (true); -- Allow all access for local development

CREATE POLICY "Users can access own conversation documents" ON public.conversation_documents
    FOR ALL USING (true); -- Allow all access for local development

-- Update trigger for chat_conversations.updated_at
CREATE OR REPLACE FUNCTION update_chat_conversation_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_chat_conversations_updated_at
    BEFORE UPDATE ON public.chat_conversations
    FOR EACH ROW
    EXECUTE FUNCTION update_chat_conversation_updated_at();

-- Update trigger for message count
CREATE OR REPLACE FUNCTION update_conversation_message_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE public.chat_conversations
        SET message_count = message_count + 1,
            last_activity = NOW()
        WHERE id = NEW.conversation_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE public.chat_conversations
        SET message_count = message_count - 1,
            last_activity = NOW()
        WHERE id = OLD.conversation_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_conversation_message_count_trigger
    AFTER INSERT OR DELETE ON public.chat_messages
    FOR EACH ROW
    EXECUTE FUNCTION update_conversation_message_count();