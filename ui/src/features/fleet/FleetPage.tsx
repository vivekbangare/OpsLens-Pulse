import { useEffect, useState, useMemo } from "react"
import StatCard from "../../shared/components/ui/StatCard"
import { fetchHosts } from "../hosts/api"
import { useNavigate } from "react-router-dom"

type Host = {
  agent_id: string
  hostname: string
  alive: boolean
  cpu_percent?: number
  mem_used_mb?: number
  mem_total_mb?: number
}

export default function FleetPage() {
  const [hosts, setHosts] = useState<Host[]>([])
  const navigate = useNavigate()

  useEffect(() => {
    load()
  }, [])

  async function load() {
    try {
      const data = await fetchHosts()
      setHosts(Array.isArray(data) ? data : [])
    } catch (err) {
      console.error("Failed to load fleet:", err)
    }
  }

  const total = hosts.length
  const alive = hosts.filter(h => h.alive).length
  const dead = total - alive

  const fleetHealth = useMemo(() => {
    if (!total) return "—"
    return Math.round((alive / total) * 100) + "%"
  }, [alive, total])

  // Recently visited (stored locally)
  const recent = JSON.parse(localStorage.getItem("recentHosts") || "[]")

  return (
    <div>
      <h2>Fleet Overview</h2>
      <p className="subtitle">
        Real-time infrastructure health summary
      </p>

      {/* Stats */}
      <div className="grid">
        <StatCard label="Total Hosts" value={total} />
        <StatCard label="Alive" value={alive} />
        <StatCard label="Dead" value={dead} />
        <StatCard label="Fleet Health" value={fleetHealth} />
      </div>

      {/* Recently Visited */}
      <div className="cardPanel" style={{ marginTop: 20 }}>
        <h3>Recently Visited Hosts</h3>

        {recent.length === 0 ? (
          <div className="subText">No recent hosts</div>
        ) : (
          recent.slice(0, 5).map((h: any) => (
            <div
              key={h.agent_id}
              className="recentHostItem"
              onClick={() => navigate(`/hosts/${h.agent_id}`)}
            >
              {h.hostname}
            </div>
          ))
        )}
      </div>
    </div>
  )
}