import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"

import {
  fetchLogs,
  fetchHostSummary,
  fetchHosts,
  fetchContainerMetrics,
} from "../api/client"

interface Props {
  agentId: string
}

type Tags = Record<string, string>

interface HostSummary {
  cpu_percent?: number
  mem_used_mb?: number
  mem_total_mb?: number
  disk_used_mb?: number
  network_in_mb?: number
  network_out_mb?: number
  uptime_sec?: number
  tags?: Tags
}

interface LogEntry {
  timestamp?: number
  level?: string
  message?: string
}

interface ContainerMetric {
  name: string
  status: string
}

export default function HostDetails({ agentId }: Props) {
  const navigate = useNavigate()

  const [logs, setLogs] = useState<LogEntry[]>([])
  const [summary, setSummary] = useState<HostSummary | null>(null)
  const [containers, setContainers] = useState<ContainerMetric[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadAll()

    const interval = setInterval(() => {
      loadAll()
    }, 5000)

    return () => clearInterval(interval)
  }, [agentId])

  async function loadAll() {
    setLoading(true)

    try {
      const [summaryData, logData, containerData, hostsData] =
        await Promise.all([
          fetchHostSummary(agentId),
          fetchLogs(agentId),
          fetchContainerMetrics(agentId).catch(() => []),
          fetchHosts(), // 👈 reuse this
        ])

      const currentHost = hostsData.find(
        (h: any) => h.agent_id === agentId
      )

      if (summaryData) {
        summaryData.tags = currentHost?.tags ?? null
      }

      setSummary(summaryData ?? null)
      setLogs(Array.isArray(logData) ? logData : [])
      setContainers(Array.isArray(containerData) ? containerData : [])

    } catch (err) {
      console.error("Unexpected error:", err)
    } finally {
      setLoading(false)
    }
  }

  function formatUptime(seconds?: number) {
    if (!seconds) return "—"

    const hours = Math.floor(seconds / 3600)
    const days = Math.floor(hours / 24)

    if (days > 0) return `${days}d ${hours % 24}h`
    return `${hours}h`
  }

  return (
    <div style={{ padding: "20px" }}>
      <button onClick={() => navigate(-1)}>⬅ Back</button>

      <h2 style={{ marginTop: "10px" }}>Host Details</h2>

      {loading && <p>Loading data...</p>}

      {/* ================= METRICS ================= */}
      <div
        style={{
          background: "#111827",
          padding: "20px",
          borderRadius: "12px",
          marginBottom: "20px",
        }}
      >
        <h3>Key Metrics (1h)</h3>

        <div
          style={{
            display: "flex",
            gap: "30px",
            flexWrap: "wrap",
            marginTop: "10px",
          }}
        >
          <div>
            <strong>CPU:</strong>{" "}
            {summary?.cpu_percent !== undefined
              ? summary.cpu_percent.toFixed(2)
              : "—"}{" "}
            %
          </div>

          <div>
            <strong>RAM:</strong>{" "}
            {summary?.mem_used_mb ?? "—"} /{" "}
            {summary?.mem_total_mb ?? "—"} MB
          </div>

          <div>
            <strong>Disk Used:</strong>{" "}
            {summary?.disk_used_mb ?? "—"} MB
          </div>

          <div>
            <strong>Network In:</strong>{" "}
            {summary?.network_in_mb ?? "—"} MB
          </div>

          <div>
            <strong>Network Out:</strong>{" "}
            {summary?.network_out_mb ?? "—"} MB
          </div>

          <div>
            <strong>Uptime:</strong>{" "}
            {formatUptime(summary?.uptime_sec)}
          </div>
        </div>

        {/* ================= TAGS ================= */}
        <div style={{ marginTop: "20px" }}>
          <strong>Tags:</strong>

          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fill, minmax(200px, 1fr))",
              gap: "8px",
              marginTop: "10px",
            }}
          >
            {summary?.tags && Object.keys(summary.tags).length > 0 ? (
              Object.entries(summary.tags).map(([key, value]) => (
                <div
                  key={key}
                  style={{
                    padding: "8px 12px",
                    background: "#1f2937",
                    borderRadius: "8px",
                    border: "1px solid #374151",
                    fontSize: "13px",
                  }}
                >
                  <strong style={{ opacity: 0.7 }}>{key}:</strong>{" "}
                  {value}
                </div>
              ))
            ) : (
              <span style={{ opacity: 0.6 }}>No tags</span>
            )}
          </div>
        </div>
      </div>

      {/* ================= LOGS ================= */}
      <div
        style={{
          background: "#0f172a",
          padding: "15px",
          borderRadius: "12px",
          marginBottom: "20px",
        }}
      >
        <h3>Logs</h3>

        {logs.length === 0 && <p>No logs available</p>}

        <div
          style={{
            maxHeight: "300px",
            overflowY: "auto",
            fontFamily: "monospace",
            fontSize: "13px",
            marginTop: "10px",
          }}
        >
          {logs.map((l, i) => (
            <div key={i} style={{ marginBottom: "4px" }}>
              {l.timestamp
                ? new Date(l.timestamp * 1000).toLocaleString()
                : "—"}{" "}
              [{l.level ?? "info"}] {l.message ?? ""}
            </div>
          ))}
        </div>
      </div>

      {/* ================= CONTAINERS ================= */}
      <div
        style={{
          background: "#0f172a",
          padding: "15px",
          borderRadius: "12px",
        }}
      >
        <h3>Containers</h3>

        {containers.length === 0 && <p>No containers detected</p>}

        {containers.map((c, i) => (
          <div key={i} style={{ marginBottom: "6px" }}>
            {c.name} — {c.status}
          </div>
        ))}
      </div>
    </div>
  )
}