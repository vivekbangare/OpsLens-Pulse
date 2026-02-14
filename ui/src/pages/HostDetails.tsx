import { useEffect, useState } from "react"
import { fetchLogs } from "../api/client"

interface Props {
  agentId: string
  onBack: () => void
}

export default function HostDetails({ agentId, onBack }: Props) {
  const [logs, setLogs] = useState<any[]>([])

  useEffect(() => {
    loadLogs()
  }, [])

  async function loadLogs() {
    try {
      const data = await fetchLogs(agentId)
      if (Array.isArray(data)) {
        setLogs(data)
      }
    } catch (err) {
      console.error("Failed to fetch logs:", err)
    }
  }

  return (
    <div style={{ padding: "20px" }}>
      <button onClick={onBack}>⬅ Back</button>

      <h2>Logs - {agentId}</h2>

      <div style={{
        background: "#0f172a",
        padding: "15px",
        borderRadius: "12px",
        marginTop: "20px",
        maxHeight: "500px",
        overflowY: "auto",
        fontFamily: "monospace",
        fontSize: "13px"
      }}>
        {logs.map((l, i) => (
          <div key={i} style={{ marginBottom: "4px" }}>
            {new Date(l.timestamp * 1000).toLocaleString()} [{l.level}] {l.message}
          </div>
        ))}
      </div>
    </div>
  )
}
