'use client'

import { useState, useEffect, useRef } from 'react'
import { authService } from '@/lib/auth'
import { Button } from '@/components/ui/button'
import { Send, FileText, Upload, Paperclip } from 'lucide-react'

interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  sources?: AnswerResult[]
  confidence?: number
  created_at: string
}

interface AnswerResult {
  answer: string
  confidence: number
  source_title: string
  content_type: string
  chunk_id: string
}

interface ChatConversation {
  id: string
  title?: string
  message_count: number
  last_activity: string
  created_at: string
}

interface ChatResponse {
  conversation_id: string
  message_id: string
  response: string
  sources: AnswerResult[]
  confidence?: number
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export default function ChatPage() {
  const [conversations, setConversations] = useState<ChatConversation[]>([])
  const [currentConversation, setCurrentConversation] = useState<string | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [inputMessage, setInputMessage] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isLoadingConversations, setIsLoadingConversations] = useState(true)
  const [isUploading, setIsUploading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Auto-scroll to bottom when new messages arrive
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  // Load conversations on component mount
  useEffect(() => {
    loadConversations()
  }, [])

  // Load messages when conversation changes
  useEffect(() => {
    if (currentConversation) {
      loadMessages(currentConversation)
    }
  }, [currentConversation])

  const loadConversations = async () => {
    try {
      setIsLoadingConversations(true)
      const response = await authService.fetchWithAuth(`${API_BASE_URL}/api/v1/chat/conversations`)

      if (response.ok) {
        const data = await response.json()
        setConversations(data.conversations || [])

        // Auto-select first conversation if exists
        if (data.conversations?.length > 0 && !currentConversation) {
          setCurrentConversation(data.conversations[0].id)
        }
      }
    } catch (error) {
      console.error('Failed to load conversations:', error)
    } finally {
      setIsLoadingConversations(false)
    }
  }

  const loadMessages = async (conversationId: string) => {
    try {
      const response = await authService.fetchWithAuth(
        `${API_BASE_URL}/api/v1/chat/conversations/${conversationId}/messages`
      )

      if (response.ok) {
        const data = await response.json()
        setMessages(data.messages || [])
      }
    } catch (error) {
      console.error('Failed to load messages:', error)
    }
  }

  const createNewConversation = async () => {
    try {
      const response = await authService.fetchWithAuth(
        `${API_BASE_URL}/api/v1/chat/conversations`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
        }
      )

      if (response.ok) {
        const data = await response.json()
        const newConversation = data.conversation
        setConversations(prev => [newConversation, ...prev])
        setCurrentConversation(newConversation.id)
        setMessages([])
      }
    } catch (error) {
      console.error('Failed to create conversation:', error)
    }
  }

  const sendMessage = async () => {
    if (!inputMessage.trim()) return

    const messageText = inputMessage.trim()
    setInputMessage('')
    setIsLoading(true)

    try {
      // If no current conversation, create one
      let conversationId = currentConversation
      if (!conversationId) {
        await createNewConversation()
        conversationId = currentConversation
      }

      const response = await authService.fetchWithAuth(
        `${API_BASE_URL}/api/v1/chat/conversations/${conversationId}/message`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            conversation_id: conversationId,
            message: messageText,
          }),
        }
      )

      if (response.ok) {
        const chatResponse: ChatResponse = await response.json()

        // Add both user and assistant messages
        const userMessage: ChatMessage = {
          id: `user-${Date.now()}`,
          role: 'user',
          content: messageText,
          created_at: new Date().toISOString(),
        }

        const assistantMessage: ChatMessage = {
          id: chatResponse.message_id,
          role: 'assistant',
          content: chatResponse.response,
          sources: chatResponse.sources,
          confidence: chatResponse.confidence,
          created_at: new Date().toISOString(),
        }

        setMessages(prev => [...prev, userMessage, assistantMessage])

        // Refresh conversations to update last activity
        loadConversations()
      } else {
        const error = await response.json()
        console.error('Failed to send message:', error)
      }
    } catch (error) {
      console.error('Failed to send message:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString([], {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const uploadDocument = async (file: File) => {
    setIsUploading(true)
    try {
      const formData = new FormData()
      formData.append('document', file)

      const response = await authService.fetchWithAuth(
        `${API_BASE_URL}/api/v1/text/upload`,
        {
          method: 'POST',
          body: formData,
        }
      )

      if (response.ok) {
        const result = await response.json()
        alert(`Document "${file.name}" uploaded successfully! You can now ask questions about it.`)
      } else {
        const error = await response.json()
        alert(`Upload failed: ${error.message || 'Unknown error'}`)
      }
    } catch (error) {
      console.error('Upload failed:', error)
      alert('Upload failed. Please try again.')
    } finally {
      setIsUploading(false)
    }
  }

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      uploadDocument(file)
    }
    // Reset input so same file can be selected again
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex">
      {/* Sidebar - Conversations */}
      <div className="w-80 bg-white border-r border-gray-200 flex flex-col">
        <div className="p-4 border-b border-gray-200">
          <div className="flex items-center justify-between mb-3">
            <h1 className="text-xl font-semibold text-gray-800">Chat</h1>
            <Button
              onClick={createNewConversation}
              size="sm"
              className="bg-blue-600 hover:bg-blue-700"
            >
              New Chat
            </Button>
          </div>
          <div className="flex items-center space-x-2">
            <Button
              onClick={() => fileInputRef.current?.click()}
              variant="outline"
              size="sm"
              disabled={isUploading}
              className="flex-1"
            >
              <Paperclip size={16} className="mr-2" />
              {isUploading ? 'Uploading...' : 'Upload Document'}
            </Button>
            <input
              ref={fileInputRef}
              type="file"
              accept=".pdf,.doc,.docx,.txt,.epub,.html"
              onChange={handleFileSelect}
              className="hidden"
            />
          </div>
        </div>

        <div className="flex-1 overflow-y-auto">
          {isLoadingConversations ? (
            <div className="p-4 text-center text-gray-500">Loading conversations...</div>
          ) : conversations.length === 0 ? (
            <div className="p-4 text-center text-gray-500">
              <div className="mb-2">No conversations yet</div>
              <Button onClick={createNewConversation} variant="outline" size="sm">
                Start your first chat
              </Button>
            </div>
          ) : (
            conversations.map((conv) => (
              <div
                key={conv.id}
                onClick={() => setCurrentConversation(conv.id)}
                className={`p-4 border-b border-gray-100 cursor-pointer hover:bg-gray-50 ${
                  currentConversation === conv.id ? 'bg-blue-50 border-blue-200' : ''
                }`}
              >
                <div className="font-medium text-gray-800 truncate">
                  {conv.title || `Chat ${conv.message_count} messages`}
                </div>
                <div className="text-sm text-gray-500 mt-1">
                  {formatDate(conv.last_activity)}
                </div>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Main Chat Area */}
      <div className="flex-1 flex flex-col">
        {currentConversation ? (
          <>
            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
              {messages.map((message) => (
                <div
                  key={message.id}
                  className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}
                >
                  <div
                    className={`max-w-2xl rounded-lg px-4 py-2 ${
                      message.role === 'user'
                        ? 'bg-blue-600 text-white'
                        : 'bg-white border border-gray-200 text-gray-800'
                    }`}
                  >
                    <div className="whitespace-pre-wrap">{message.content}</div>

                    {/* Sources for assistant messages */}
                    {message.role === 'assistant' && message.sources && message.sources.length > 0 && (
                      <div className="mt-3 pt-3 border-t border-gray-200">
                        <div className="text-xs text-gray-500 mb-2">Sources:</div>
                        <div className="space-y-1">
                          {message.sources.map((source, idx) => (
                            <div key={idx} className="text-xs bg-gray-50 rounded px-2 py-1">
                              <div className="font-medium">{source.source_title}</div>
                              <div className="text-gray-600">
                                Confidence: {Math.round(source.confidence * 100)}%
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    <div className="text-xs opacity-70 mt-2">
                      {formatTime(message.created_at)}
                    </div>
                  </div>
                </div>
              ))}

              {isLoading && (
                <div className="flex justify-start">
                  <div className="bg-white border border-gray-200 rounded-lg px-4 py-2">
                    <div className="flex items-center space-x-2">
                      <div className="animate-spin w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full"></div>
                      <span className="text-gray-600">Thinking...</span>
                    </div>
                  </div>
                </div>
              )}

              <div ref={messagesEndRef} />
            </div>

            {/* Input Area */}
            <div className="border-t border-gray-200 p-4 bg-white">
              <div className="flex items-end space-x-2">
                <div className="flex-1">
                  <textarea
                    value={inputMessage}
                    onChange={(e) => setInputMessage(e.target.value)}
                    onKeyPress={handleKeyPress}
                    placeholder="Ask a question about your documents..."
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    rows={1}
                    disabled={isLoading}
                  />
                </div>
                <Button
                  onClick={sendMessage}
                  disabled={!inputMessage.trim() || isLoading}
                  className="bg-blue-600 hover:bg-blue-700 disabled:opacity-50"
                >
                  <Send size={16} />
                </Button>
              </div>
            </div>
          </>
        ) : (
          /* Welcome Screen */
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center max-w-md">
              <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <FileText size={32} className="text-blue-600" />
              </div>
              <h2 className="text-xl font-semibold text-gray-800 mb-2">
                Welcome to Self Chat
              </h2>
              <p className="text-gray-600 mb-6">
                Ask questions about your documents and get AI-powered answers with source citations.
              </p>
              <Button onClick={createNewConversation} className="bg-blue-600 hover:bg-blue-700">
                Start New Conversation
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}