import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'

import {
  fetchMe,
  login as apiLogin,
  register as apiRegister,
  type AuthUser,
} from '@/lib/api'

const TOKEN_KEY = 'racecoach_token'

type AuthContextValue = {
  user: AuthUser | null
  token: string | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(TOKEN_KEY))
  const [user, setUser] = useState<AuthUser | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false

    async function hydrate() {
      if (!token) {
        setUser(null)
        setLoading(false)
        return
      }

      try {
        const me = await fetchMe(token)
        if (!cancelled) {
          setUser(me)
        }
      } catch {
        localStorage.removeItem(TOKEN_KEY)
        if (!cancelled) {
          setToken(null)
          setUser(null)
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void hydrate()
    return () => {
      cancelled = true
    }
  }, [token])

  const establishSession = useCallback(async (nextToken: string) => {
    localStorage.setItem(TOKEN_KEY, nextToken)
    setToken(nextToken)
    const me = await fetchMe(nextToken)
    setUser(me)
  }, [])

  const login = useCallback(
    async (email: string, password: string) => {
      const nextToken = await apiLogin(email, password)
      await establishSession(nextToken)
    },
    [establishSession],
  )

  const register = useCallback(
    async (email: string, password: string) => {
      const nextToken = await apiRegister(email, password)
      await establishSession(nextToken)
    },
    [establishSession],
  )

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY)
    setToken(null)
    setUser(null)
  }, [])

  const value = useMemo(
    () => ({ user, token, loading, login, register, logout }),
    [user, token, loading, login, register, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}
