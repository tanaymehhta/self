'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { authService, type User } from '@/lib/auth'

export default function DashboardPage() {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const router = useRouter()

  useEffect(() => {
    if (!authService.isAuthenticated()) {
      router.push('/auth/login')
      return
    }

    const currentUser = authService.getCurrentUser()
    setUser(currentUser)
    setLoading(false)
  }, [router])

  const handleSignOut = () => {
    authService.logout()
    router.push('/auth/login')
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Navigation Header */}
      <header className="border-b border-border bg-card">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-bold">Self Dashboard</h1>
            <div className="flex items-center gap-4">
              <button
                onClick={() => router.push('/chat')}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                Open Chat
              </button>
              <span className="text-sm text-muted-foreground">
                Welcome, {user?.full_name || user?.email}
              </span>
              <button
                onClick={handleSignOut}
                className="text-sm text-muted-foreground hover:text-foreground"
              >
                Sign out
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        <div className="grid gap-6">
          {/* Recording Status Card */}
          <div className="bg-card border border-border rounded-lg p-6">
            <div className="flex items-center gap-4">
              <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
              <div>
                <h2 className="font-semibold">Recording Status</h2>
                <p className="text-sm text-muted-foreground">
                  Ready to record - Click to start capturing audio
                </p>
              </div>
            </div>
          </div>

          {/* Quick Stats */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="bg-card border border-border rounded-lg p-6">
              <h3 className="font-semibold text-lg">0</h3>
              <p className="text-sm text-muted-foreground">Conversations</p>
            </div>
            <div className="bg-card border border-border rounded-lg p-6">
              <h3 className="font-semibold text-lg">0</h3>
              <p className="text-sm text-muted-foreground">Hours Transcribed</p>
            </div>
            <div className="bg-card border border-border rounded-lg p-6">
              <h3 className="font-semibold text-lg">0</h3>
              <p className="text-sm text-muted-foreground">Files Monitored</p>
            </div>
          </div>

          {/* Recent Activity */}
          <div className="bg-card border border-border rounded-lg p-6">
            <h2 className="font-semibold mb-4">Recent Activity</h2>
            <div className="text-center py-12">
              <p className="text-muted-foreground">
                No activity yet. Start by recording your first conversation or uploading an audio file.
              </p>
            </div>
          </div>

          {/* Getting Started */}
          <div className="bg-card border border-border rounded-lg p-6">
            <h2 className="font-semibold mb-4">Getting Started</h2>
            <div className="grid gap-4">
              <div className="flex items-start gap-3">
                <div className="w-2 h-2 bg-primary rounded-full mt-2"></div>
                <div>
                  <h3 className="font-medium">Download the Desktop App</h3>
                  <p className="text-sm text-muted-foreground">
                    Install the Self desktop application to start recording audio and monitoring files.
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <div className="w-2 h-2 bg-primary rounded-full mt-2"></div>
                <div>
                  <h3 className="font-medium">Record Your First Conversation</h3>
                  <p className="text-sm text-muted-foreground">
                    Use the desktop app to record a conversation and see it transcribed automatically.
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <div className="w-2 h-2 bg-primary rounded-full mt-2"></div>
                <div>
                  <h3 className="font-medium">Connect Your Services</h3>
                  <p className="text-sm text-muted-foreground">
                    Link your calendar, email, and other services to get intelligent insights.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}