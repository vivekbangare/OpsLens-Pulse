import { useEffect, useState, useMemo } from "react"
import { useNavigate } from "react-router-dom"
import HostTable from "./components/HostTable"
import { fetchHosts } from "./api"

type Host = {
  agent_id: string
  alive: boolean
  hostname: string
  public_ip: string
  ip: string
  os?: string
  first_seen?: string
  last_seen?: string
  tags?: Record<string, string>
  cpu_percent?: number
  mem_used_mb?: number
  mem_total_mb?: number
}

export default function Hosts() {
  const navigate = useNavigate()

  const [hosts, setHosts] = useState<Host[]>([])
  const [loading, setLoading] = useState(true)

  const [query, setQuery] = useState("")
  const [statusFilter, setStatusFilter] =
    useState<"all" | "active" | "down">("all")

  useEffect(() => {
    load()
    const interval = setInterval(load, 30000)
    return () => clearInterval(interval)
  }, [])

  async function load() {
    try {
      const data = await fetchHosts()
      setHosts(Array.isArray(data) ? data : [])
    } catch (err) {
      console.error("Failed to load hosts:", err)
    } finally {
      setLoading(false)
    }
  }

  const filteredHosts = useMemo(() => {
    let result = [...hosts]

    if (query) {
      const q = query.toLowerCase()
      result = result.filter(
        (h) =>
          h.hostname.toLowerCase().includes(q) ||
          h.ip.toLowerCase().includes(q)
      )
    }

    if (statusFilter !== "all") {
      result = result.filter((h) =>
        statusFilter === "active" ? h.alive : !h.alive
      )
    }

    return result
  }, [hosts, query, statusFilter])

  return (
    <div className="pageContainer">

      {/* ================= HEADER ================= */}
      <div className="pageHeader">
        <div className="titleRow">
          <h2>Hosts</h2>

          <span className="hostCountBadge">
            {filteredHosts.length}{" "}
            {filteredHosts.length === 1 ? "Host" : "Hosts"}
          </span>
        </div>

        <p className="pageSubtitle">
          Manage infrastructure nodes
        </p>
      </div>

      {/* ================= CONTROLS ================= */}
      <div className="hostsControls">

        <div className="searchBox">
          <input
            className="input"
            placeholder="Search hostname or IP..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>

        <div className="statusFilters">
          {["all", "active", "down"].map((status) => (
            <button
              key={status}
              className={`filterBtn ${
                statusFilter === status ? "activeFilter" : ""
              }`}
              onClick={() =>
                setStatusFilter(status as any)
              }
            >
              {status.toUpperCase()}
            </button>
          ))}
        </div>

      </div>

      {/* ================= TABLE ================= */}
      {loading ? (
        <div className="cardPanel" style={{ padding: 20 }}>
          Loading hosts...
        </div>
      ) : (
        <div className="tableWrap">
          <HostTable
            hosts={filteredHosts}
            onSelect={(id) => navigate(`/hosts/${id}`)}
          />
        </div>
      )}

    </div>
  )
}