import { Navigate, Outlet } from "react-router-dom"
import { useAuth } from "../features/auth/AuthContext"

export default function ProtectedRoute() {
  const { isAuthenticated } = useAuth()

  // ✅ In dev mode always allow
  if (import.meta.env.DEV) {
    return <Outlet />
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <Outlet />
}