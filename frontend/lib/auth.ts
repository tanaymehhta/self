// Local authentication service (no Supabase)
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export interface User {
  id: string
  email: string
  full_name?: string
  avatar_url?: string
  created_at: string
}

export interface AuthTokens {
  access_token: string
  refresh_token: string
  expires_at: string
  token_type: string
}

export interface AuthResponse {
  user: User
  tokens: AuthTokens
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterCredentials {
  email: string
  password: string
  full_name?: string
}

class AuthService {
  private getStoredTokens(): AuthTokens | null {
    if (typeof window === 'undefined') return null

    const stored = localStorage.getItem('auth_tokens')
    if (!stored) return null

    try {
      return JSON.parse(stored)
    } catch {
      return null
    }
  }

  private setStoredTokens(tokens: AuthTokens | null) {
    if (typeof window === 'undefined') return

    if (tokens) {
      localStorage.setItem('auth_tokens', JSON.stringify(tokens))
    } else {
      localStorage.removeItem('auth_tokens')
    }
  }

  private getStoredUser(): User | null {
    if (typeof window === 'undefined') return null

    const stored = localStorage.getItem('auth_user')
    if (!stored) return null

    try {
      return JSON.parse(stored)
    } catch {
      return null
    }
  }

  private setStoredUser(user: User | null) {
    if (typeof window === 'undefined') return

    if (user) {
      localStorage.setItem('auth_user', JSON.stringify(user))
    } else {
      localStorage.removeItem('auth_user')
    }
  }

  async login(credentials: LoginCredentials): Promise<AuthResponse> {
    const response = await fetch(`${API_BASE_URL}/api/v1/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(credentials),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.message || 'Login failed')
    }

    const authResponse: AuthResponse = await response.json()

    // Store tokens and user
    this.setStoredTokens(authResponse.tokens)
    this.setStoredUser(authResponse.user)

    return authResponse
  }

  async register(credentials: RegisterCredentials): Promise<AuthResponse> {
    const response = await fetch(`${API_BASE_URL}/api/v1/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(credentials),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.message || 'Registration failed')
    }

    const authResponse: AuthResponse = await response.json()

    // Store tokens and user
    this.setStoredTokens(authResponse.tokens)
    this.setStoredUser(authResponse.user)

    return authResponse
  }

  async refreshTokens(): Promise<AuthTokens | null> {
    const currentTokens = this.getStoredTokens()
    if (!currentTokens?.refresh_token) return null

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          refresh_token: currentTokens.refresh_token,
        }),
      })

      if (!response.ok) {
        this.logout()
        return null
      }

      const newTokens: AuthTokens = await response.json()
      this.setStoredTokens(newTokens)

      return newTokens
    } catch {
      this.logout()
      return null
    }
  }

  logout() {
    this.setStoredTokens(null)
    this.setStoredUser(null)
  }

  getAccessToken(): string | null {
    const tokens = this.getStoredTokens()
    if (!tokens) return null

    // Check if token is expired
    const expiresAt = new Date(tokens.expires_at)
    const now = new Date()

    if (now >= expiresAt) {
      // Token expired, try to refresh
      this.refreshTokens()
      return null
    }

    return tokens.access_token
  }

  getCurrentUser(): User | null {
    return this.getStoredUser()
  }

  isAuthenticated(): boolean {
    return this.getAccessToken() !== null && this.getCurrentUser() !== null
  }

  async fetchWithAuth(url: string, options: RequestInit = {}): Promise<Response> {
    const accessToken = this.getAccessToken()

    if (!accessToken) {
      throw new Error('No access token available')
    }

    const headers = {
      ...options.headers,
      'Authorization': `Bearer ${accessToken}`,
    }

    const response = await fetch(url, { ...options, headers })

    // If unauthorized, try to refresh token
    if (response.status === 401) {
      const newTokens = await this.refreshTokens()
      if (newTokens) {
        // Retry with new token
        const newHeaders = {
          ...options.headers,
          'Authorization': `Bearer ${newTokens.access_token}`,
        }
        return fetch(url, { ...options, headers: newHeaders })
      } else {
        throw new Error('Authentication failed')
      }
    }

    return response
  }
}

export const authService = new AuthService()
export default authService