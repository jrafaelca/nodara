import { createContext, useContext, useEffect, useMemo, useState } from 'react'

const AuthContext = createContext(null)

async function request(path, options = {}) {
  const response = await fetch(path, {
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  })
  const body = await response.json().catch(() => ({}))
  if (!response.ok) throw new Error(body.error || 'Request failed')
  return body
}

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)

  const refresh = async () => {
    try {
      const data = await request('/api/auth/me')
      setUser(data.user)
      return data.user
    } catch {
      setUser(null)
      return null
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { refresh() }, [])

  const value = useMemo(() => ({
    user,
    loading,
    login: async (identifier, password) => {
      const data = await request('/api/auth/login', { method: 'POST', body: JSON.stringify({ identifier, password }) })
      setUser(data.user)
      return data.user
    },
    logout: async () => {
      await request('/api/auth/logout', { method: 'POST', body: '{}' })
      setUser(null)
    },
    changePassword: async (password, passwordConfirmation) => {
      const data = await request('/api/auth/change-password', { method: 'POST', body: JSON.stringify({ password, password_confirmation: passwordConfirmation }) })
      setUser(data.user)
      return data.user
    },
    forgotPassword: (identifier) => request('/api/auth/forgot-password', { method: 'POST', body: JSON.stringify({ identifier }) }),
    resetPassword: (token, password, passwordConfirmation) => request('/api/auth/reset-password', { method: 'POST', body: JSON.stringify({ token, password, password_confirmation: passwordConfirmation }) }),
  }), [user, loading])

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth must be used within AuthProvider')
  return context
}
