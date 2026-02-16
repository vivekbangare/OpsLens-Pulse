import { useEffect, useState } from "react"
import { fetchHosts } from "../api/client"

export default function FleetOverview() {
  const [hosts, setHosts] = useState<any[]>([])

  useEffect(() => {
    load()
    const interval = setInterval(load, 5000)
    return () => clearInterval(interval)
  }, [])

  async function load() {
    const data = await fetchHosts()
    setHosts(Array.isArray(data) ? data : [])
  }

  // ---------------- KPI Calculations ----------------
  const safeHosts = Array.isArray(hosts) ? hosts : []

  const total = safeHosts.length
  const alive = safeHosts.filter(h => h?.alive).length
  const dead = total - alive

  const fleetHealth =
    total > 0 ? Math.round((alive / total) * 100) : 0

  const environments = [
    ...new Set(hosts.map((h) => h.environment).filter(Boolean)),
  ]

  return (
    <div style={{ padding: "20px" }}>
      <h2>Fleet Overview</h2>
      <p style={{ opacity: 0.6 }}>
        Real-time infrastructure health summary
      </p>

      {/* ---------------- KPI GRID ---------------- */}
      <div className="kpis" style={{ marginTop: "20px" }}>
        <div className="kpi">
          <div className="label">Total Hosts</div>
          <div className="value">{total}</div>
        </div>

        <div className="kpi">
          <div className="label">Alive</div>
          <div className="value">{alive}</div>
        </div>

        <div className="kpi">
          <div className="label">Dead</div>
          <div className="value">{dead}</div>
        </div>

        <div className="kpi">
          <div className="label">Fleet Health</div>
          <div className="value">{fleetHealth}%</div>
        </div>

        <div className="kpi">
          <div className="label">Environments</div>
          <div className="value">{environments.length}</div>
        </div>
      </div>

      {/* ---------------- RECENT HOSTS ---------------- */}
      <div
        className="card"
        style={{ marginTop: "30px", padding: "20px" }}
      >
        <h3>Recently Seen Hosts</h3>

        {hosts.length === 0 && (
          <div style={{ opacity: 0.6, marginTop: "10px" }}>
            No hosts connected
          </div>
        )}

        {hosts.slice(0, 5).map((h) => (
          <div
            key={h.agent_id}
            style={{
              display: "flex",
              justifyContent: "space-between",
              padding: "8px 0",
              borderBottom: "1px solid #1f2937",
            }}
          >
            <div>
              <strong>{h.hostname}</strong>
              <div style={{ fontSize: 12, opacity: 0.6 }}>
                {h.ip}
              </div>
            </div>

            <div
              style={{
                color: h.alive ? "#22c55e" : "#ef4444",
                fontWeight: 600,
              }}
            >
              {h.alive ? "Active" : "Down"}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
