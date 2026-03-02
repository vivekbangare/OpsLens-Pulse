import { createContext, useContext, useState, ReactNode, useEffect } from "react"

export interface User {
  id: string
  username: string
  email?: string
  is_super_admin?: boolean
  tenants?: any[]
  currentTenantId?: string | null
}

interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
}

interface AuthContextType extends AuthState {
  login: (data: { token: string; user: User }) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(
    localStorage.getItem("token")
  )

  const [user, setUser] = useState<User | null>(
    localStorage.getItem("user")
      ? JSON.parse(localStorage.getItem("user")!)
      : null
  )

  const isAuthenticated = !!token

  function login(data: { token: string; user: User }) {
    setToken(data.token)
    setUser(data.user)

    localStorage.setItem("token", data.token)
    localStorage.setItem("user", JSON.stringify(data.user))
  }

  function logout() {
    setToken(null)
    setUser(null)

    localStorage.removeItem("token")
    localStorage.removeItem("user")
  }

  // 🔥 ADD THIS BLOCK (Global 401 Listener)
  useEffect(() => {
    function handleGlobalLogout() {
      logout()
    }

    window.addEventListener("auth:logout", handleGlobalLogout)

    return () => {
      window.removeEventListener("auth:logout", handleGlobalLogout)
    }
  }, [])

  return (
    <AuthContext.Provider
      value={{
        token,
        user,
        isAuthenticated,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider")
  }
  return ctx
}