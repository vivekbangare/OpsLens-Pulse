import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { useAuth } from "../auth/AuthContext"
import { loginRequest } from "./api"

function parseJwt(token: string) {
  try {
    const base64Url = token.split(".")[1]
    const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/")
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split("")
        .map((c) =>
          "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2)
        )
        .join("")
    )
    return JSON.parse(jsonPayload)
  } catch {
    return null
  }
}

export default function Login() {
  const { login } = useAuth()
  const navigate = useNavigate()

  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()

    setLoading(true)
    setError("")

    try {
      // const res = await fetch("/api/login", {
      //   method: "POST",
      //   headers: { "Content-Type": "application/json" },
      //   body: JSON.stringify({ username, password }),
      // })

      // if (!res.ok) {
      //   throw new Error("Invalid credentials")
      // }

      // const data = await res.json()

      const data = await loginRequest(username, password)

      if (!data.token) {
        throw new Error("Token missing in response")
      }

      // 🔥 Decode token
      const payload = parseJwt(data.token)
      console.log("JWT PAYLOAD:", payload)

      if (!payload) {
        throw new Error("Invalid token payload")
      }

      login({
        token: data.token,
        user: {
          id: payload.sub || "1",
          username: payload.username || payload.sub || "admin",
          email: payload.email || "",
          is_super_admin: payload.is_super_admin || false,
          tenants: payload.tenants || [],
          currentTenantId:
            payload.currentTenantId ||
            payload.current_tenant_id ||
            payload.tenants?.[0]?.id ||
            null,
        },
      })

      navigate("/fleet")
    } catch (err: any) {
      console.error("LOGIN ERROR:", err)
      setError(err.message || "Login failed")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-page-wrapper">
      <div className="login-container">
        <h1>OpsLens</h1>
        <p style={{ opacity: 0.7, marginBottom: "20px" }}>
          Sign in to continue
        </p>

        <form onSubmit={handleSubmit}>
          <input
            type="text"
            placeholder="Username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />

          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />

          <button type="submit" disabled={loading}>
            {loading ? "Signing in..." : "Login"}
          </button>
        </form>

        {error && (
          <div style={{ marginTop: "14px", color: "#FF4D6D" }}>
            {error}
          </div>
        )}
      </div>
    </div>
  )
}