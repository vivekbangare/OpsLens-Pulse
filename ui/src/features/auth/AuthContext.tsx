import {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react"

export interface Tenant {
  id: string
  name: string
  slug: string
}

export interface User {
  id: string
  username: string
  email: string
  is_super_admin?: boolean
  currentTenantId?: string
  tenants?: Tenant[]
  permissions?: string[]
}

interface AuthContextType {
  token: string | null
  user: User | null
  login: (data: { token: string; user: User }) => void
  logout: () => void
  updateTenant: (tenantId: string) => void
  isAuthenticated: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(
  undefined
)

// -----------------------------
// SAFE STORAGE HELPERS
// -----------------------------
function safeParse<T>(value: string | null): T | null {
  if (!value || value === "undefined" || value === "null") {
    return null
  }

  try {
    return JSON.parse(value)
  } catch {
    return null
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(null)
  const [user, setUser] = useState<User | null>(null)

  // -----------------------------
  // LOAD FROM STORAGE (SAFE)
  // -----------------------------
  useEffect(() => {
    const storedToken = sessionStorage.getItem("token")
    const storedUser = safeParse<User>(
      sessionStorage.getItem("user")
    )

    if (storedToken && storedUser) {
      setToken(storedToken)
      setUser(storedUser)
      return
    }
    // ✅ DEV AUTO LOGIN
    if (import.meta.env.DEV) {
      const mockUser: User = {
        id: "1",
        username: "dev",
        email: "dev@opslens.local",
        is_super_admin: true,
        currentTenantId: "t1",
        tenants: [
          { id: "t1", name: "Dev Tenant", slug: "dev" },
        ],
      }

      sessionStorage.setItem("token", "dev-token")
      sessionStorage.setItem("user", JSON.stringify(mockUser))

      setToken("dev-token")
      setUser(mockUser)
    }

  }, [])

  // -----------------------------
  // LOGIN
  // -----------------------------
  function login(data: { token: string; user: User }) {
    console.log("Login User:", data.user)

    if (!data.user){
      console.error("Login failed: No user data received")
      return  
    }

    sessionStorage.setItem("token", data.token)
    sessionStorage.setItem("user", JSON.stringify(data.user))

    setToken(data.token)
    setUser(data.user)
  }

  // -----------------------------
  // LOGOUT
  // -----------------------------
  function logout() {
    sessionStorage.removeItem("token")
    sessionStorage.removeItem("user")

    setToken(null)
    setUser(null)
  }

  // -----------------------------
  // UPDATE TENANT
  // -----------------------------
  function updateTenant(tenantId: string) {
    if (!user) return

    const updatedUser = {
      ...user,
      currentTenantId: tenantId,
    }

    sessionStorage.setItem("user", JSON.stringify(updatedUser))
    setUser(updatedUser)
  }

  const value: AuthContextType = {
    token,
    user,
    login,
    logout,
    updateTenant,
    isAuthenticated: !!token,
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextType {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error("useAuth must be used inside AuthProvider")
  }
  return ctx
}
