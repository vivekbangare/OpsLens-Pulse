import { useEffect, useState, useMemo } from "react"

import {
  fetchLogs,
  fetchHostSummary,
  fetchContainerMetrics,
} from "../api"

import {
  LineChart,
  Line,
  ResponsiveContainer,
} from "recharts"

interface Props {
  host: any
  onClose: () => void
}

export default function Drawer({ host, onClose }: Props) {
  const [summary, setSummary] = useState<any | null>(null)
  const [logs, setLogs] = useState<any[]>([])
  const [containers, setContainers] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  const [cpuHistory, setCpuHistory] = useState<number[]>([])
  const [ramHistory, setRamHistory] = useState<number[]>([])

  const [search, setSearch] = useState("")
  const [levelFilter, setLevelFilter] = useState("all")

  // 🔥 ESC key support
  useEffect(() => {
    function handleEsc(e: KeyboardEvent) {
      if (e.key === "Escape") onClose()
    }

    window.addEventListener("keydown", handleEsc)
    return () => window.removeEventListener("keydown", handleEsc)
  }, [onClose])

  useEffect(() => {
    load()

    const interval = setInterval(load, 5000)
    return () => clearInterval(interval)
  }, [host.agent_id])

  async function load() {
    try {
      setLoading(true)

      const [summaryResponse, logData, containerData] =
        await Promise.all([
          fetchHostSummary(host.agent_id),
          fetchLogs(host.agent_id),
          fetchContainerMetrics(host.agent_id),
        ])

      const actualSummary =
        summaryResponse?.[host.agent_id] ?? summaryResponse ?? null

      setSummary(actualSummary)

      setLogs(Array.isArray(logData) ? logData : [])
      setContainers(Array.isArray(containerData) ? containerData : [])

      if (actualSummary?.cpu_percent !== undefined) {
        setCpuHistory((prev) => [
          ...prev.slice(-20),
          actualSummary.cpu_percent,
        ])
      }

      if (actualSummary?.mem_used_mb !== undefined) {
        setRamHistory((prev) => [
          ...prev.slice(-20),
          actualSummary.mem_used_mb,
        ])
      }

    } catch (err) {
      console.error("Drawer load failed:", err)
    } finally {
      setLoading(false)
    }
  }

  const filteredLogs = useMemo(() => {
    return logs.filter((l) => {
      const matchesSearch =
        l.message?.toLowerCase().includes(search.toLowerCase())

      const matchesLevel =
        levelFilter === "all" || l.level === levelFilter

      return matchesSearch && matchesLevel
    })
  }, [logs, search, levelFilter])

  return (
    <div
      className="drawerOverlay"
      onClick={onClose}
    >
      <div
        className="drawer"
        onClick={(e) => e.stopPropagation()}
      >

        {/* HEADER */}
        <div className="drawerHeader">
          <div>
            <h2>{host.hostname}</h2>
            <div style={{ fontSize: 12, opacity: 0.6 }}>
              {host.ip} • {host.os}
            </div>
          </div>
          <button onClick={onClose}>✕</button>
        </div>

        {loading && <div style={{ padding: 20 }}>Loading...</div>}

        {!loading && (
          <>
            {/* METRICS */}
            <div className="drawerSection">
              <h3>Key Metrics</h3>

              <div className="metricsGrid">
                <Metric label="CPU %" value={summary?.cpu_percent ?? "—"}>
                  <Sparkline data={cpuHistory} />
                </Metric>

                <Metric label="RAM (MB)" value={summary?.mem_used_mb ?? "—"}>
                  <Sparkline data={ramHistory} />
                </Metric>

                <Metric label="Uptime (sec)" value={summary?.uptime_sec ?? "—"} />
              </div>
            </div>

            {/* TAGS */}
            <div className="drawerSection">
              <h3>Tags</h3>

              {host?.tags && Object.keys(host.tags).length > 0 ? (
                <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
                  {Object.entries(host.tags).map(([k, v]) => (
                    <span
                      key={k}
                      style={{
                        padding: "4px 10px",
                        background: "#1f2937",
                        borderRadius: "8px",
                        fontSize: "12px",
                      }}
                    >
                      {k}: {String(v)}
                    </span>
                  ))}
                </div>
              ) : (
                <div style={{ opacity: 0.6 }}>
                  No tags available
                </div>
              )}
            </div>

            {/* LOGS */}
            <div className="drawerSection">
              <h3>Logs</h3>

              <div style={{ display: "flex", gap: 8, marginBottom: 10 }}>
                <input
                  placeholder="Search logs..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="logSearch"
                />

                <select
                  value={levelFilter}
                  onChange={(e) => setLevelFilter(e.target.value)}
                  className="logFilter"
                >
                  <option value="all">All</option>
                  <option value="info">Info</option>
                  <option value="warn">Warn</option>
                  <option value="error">Error</option>
                </select>
              </div>

              <div className="logBox">
                {filteredLogs.length === 0 && (
                  <div style={{ opacity: 0.6 }}>
                    No logs available
                  </div>
                )}

                {filteredLogs.map((l, i) => (
                  <div key={i} className="logLine">
                    {l.timestamp
                      ? new Date(l.timestamp * 1000).toLocaleString()
                      : "—"}{" "}
                    [{l.level ?? "info"}] {l.message ?? ""}
                  </div>
                ))}
              </div>
            </div>

            {/* CONTAINERS */}
            <div className="drawerSection">
              <h3>Containers</h3>

              {containers.length === 0 ? (
                <div style={{ opacity: 0.6 }}>
                  No containers detected
                </div>
              ) : (
                <table className="containerTable">
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Status</th>
                      <th>CPU</th>
                      <th>RAM</th>
                    </tr>
                  </thead>
                  <tbody>
                    {containers.map((c, i) => (
                      <tr key={i}>
                        <td>{c.name}</td>
                        <td>{c.status}</td>
                        <td>{c.cpu_percent ?? "—"}%</td>
                        <td>{c.mem_used_mb ?? "—"} MB</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  )
}

function Metric({
  label,
  value,
  children,
}: {
  label: string
  value: any
  children?: React.ReactNode
}) {
  return (
    <div className="metricCard">
      <div className="metricLabel">{label}</div>
      <div className="metricValue">{value}</div>
      {children}
    </div>
  )
}

function Sparkline({ data }: { data: number[] }) {
  const chartData = data.map((v, i) => ({
    name: i,
    value: v,
  }))

  if (chartData.length === 0) return null

  return (
    <div style={{ width: "100%", height: 60, minHeight: 60 }}>
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={chartData}>
          <Line
            type="monotone"
            dataKey="value"
            stroke="#3b82f6"
            strokeWidth={2}
            dot={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}

