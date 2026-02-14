import { useState } from "react"

interface Props {
  onLogin: (token: string) => void
}

export default function Login({ onLogin }: Props) {
  const [apiKey, setApiKey] = useState("")
  const [error, setError] = useState("")

  const handleLogin = () => {
    if (!apiKey) {
      setError("API key is required")
      return
    }

    onLogin(apiKey)
  }

  return (
    <div className="login-page-wrapper">
      <div className="login-container">
        <h1>OpsLens Pulse</h1>
        <input
          type="password"
          placeholder="Paste API Key"
          value={apiKey}
          onChange={(e) => setApiKey(e.target.value)}
        />
        <button onClick={handleLogin}>Login</button>
        {error && <div className="error">{error}</div>}
      </div>
    </div>
  )
}
