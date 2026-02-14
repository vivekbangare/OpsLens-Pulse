import { Routes, Route, Navigate } from "react-router-dom"
import { useState, useEffect } from "react"
import Login from "./pages/Login"
import Hosts from "./pages/Hosts"
import AppLayout from "./layout/AppLayout"
import FleetOverview from "./pages/FleetOverview"


export default function App() {
  const [token, setToken] = useState<string | null>(null)

  useEffect(() => {
    setToken(sessionStorage.getItem("apiKey"))
  }, [])

  const handleLogin = (newToken: string) => {
    sessionStorage.setItem("apiKey", newToken)
    setToken(newToken)
  }

  const handleLogout = () => {
    sessionStorage.removeItem("apiKey")
    setToken(null)
  }

  return (
    <Routes>
      <Route
        path="/"
        element={
          token ? (
            <Navigate to="/hosts" />
          ) : (
            <Login onLogin={handleLogin} />
          )
        }
      />
      <Route
        path="/hosts"
        element={
          token ? (
            <AppLayout>
              <FleetOverview />
            </AppLayout>
          ) : (
            <Navigate to="/" />
          )
        }
      />
      <Route path="*" element={<Navigate to="/" />} />
    </Routes>
  )
}
