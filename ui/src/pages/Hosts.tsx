import { useEffect, useState } from "react"
import HostTable from "../components/HostTable"
import HostDetails from "./HostDetails"
import { fetchHosts } from "../api/client"

interface Props {
  onLogout: () => void
}

export default function Hosts({ onLogout }: Props) {
  const [hosts, setHosts] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selected, setSelected] = useState<string | null>(null)

  useEffect(() => {
    loadHosts()

    const interval = setInterval(() => {
      loadHosts()
    }, 5000)

    return () => clearInterval(interval)
  }, [])

  async function loadHosts() {
    try {
      setLoading(true)
      setError(null)

      const data = await fetchHosts()
      setHosts(Array.isArray(data) ? data : [])
    } catch (err) {
      console.error("Failed to load hosts:", err)
      setError("Unable to fetch hosts. Check API key.")
      setHosts([])
    } finally {
      setLoading(false)
    }
  }

  // ---------------------------
  // Host Details View
  // ---------------------------
  if (selected) {
    return (
      <HostDetails
        agentId={selected}
        onBack={() => setSelected(null)}
      />
    )
  }

  // ---------------------------
  // Dashboard Layout (ALWAYS RENDERED)
  // ---------------------------
  return (
    <div style={{ padding: "20px" }}>
      {/* Header */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          marginBottom: "20px",
        }}
      >
        <h2>Fleet Overview</h2>
        <button onClick={onLogout}>Logout</button>
      </div>

      {/* Loading */}
      {loading && (
        <div style={{ opacity: 0.7 }}>
          Loading hosts...
        </div>
      )}

      {/* Error */}
      {!loading && error && (
        <div style={{ color: "#ef4444" }}>
          {error}
        </div>
      )}

      {/* Empty State */}
      {!loading && !error && hosts.length === 0 && (
        <div
          style={{
            padding: "40px",
            textAlign: "center",
            background: "#111827",
            borderRadius: "12px",
          }}
        >
          <h3>No Hosts Connected</h3>
          <p style={{ opacity: 0.7 }}>
            Deploy an OpsLens agent to begin monitoring your infrastructure.
          </p>
        </div>
      )}

      {/* Host Table */}
      {!loading && !error && hosts.length > 0 && (
        <HostTable hosts={hosts} onSelect={setSelected} />
      )}
    </div>
  )
}
