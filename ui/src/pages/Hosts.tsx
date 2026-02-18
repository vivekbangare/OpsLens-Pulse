import { useEffect, useState, useMemo } from "react"
import { useNavigate } from "react-router-dom"
import { fetchHosts } from "../api/client"

type Host = {
  agent_id: string
  alive: boolean
  hostname: string
  public_ip: string
  ip: string
  os?: string
  last_seen?: string
  tags?: Record<string, string>
}

type SortKey = "hostname" | "ip" | "os" | "last_seen" | "public_ip"

export default function Hosts() {
  const navigate = useNavigate()

  const [hosts, setHosts] = useState<Host[]>([])
  const [loading, setLoading] = useState(true)

  // Search
  const [query, setQuery] = useState("")
  const [debouncedQuery, setDebouncedQuery] = useState("")

  // Status filter
  const [statusFilter, setStatusFilter] =
    useState<"all" | "active" | "down">("all")

  // Sorting
  const [sortKey, setSortKey] = useState<SortKey>("hostname")
  const [sortDir, setSortDir] = useState<"asc" | "desc">("asc")

  // Pagination
  const [page, setPage] = useState(1)
  const pageSize = 10

  // =============================
  // Debounce Search (300ms)
  // =============================
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(query.toLowerCase())
      setPage(1)
    }, 300)

    return () => clearTimeout(timer)
  }, [query])

  // =============================
  // Load Hosts
  // =============================
  useEffect(() => {
    load()
    const interval = setInterval(load, 5000)
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

  // =============================
  // Filter + Sort
  // =============================
  const processedHosts = useMemo(() => {
    let result = [...hosts]

    // 🔎 Search
    if (debouncedQuery) {
      result = result.filter((h) => {
        const tagString = Object.values(h.tags || {})
          .join(" ")
          .toLowerCase()

        return (
          h.hostname.toLowerCase().includes(debouncedQuery) ||
          h.ip.toLowerCase().includes(debouncedQuery) ||
          h.public_ip?.toLowerCase().includes(debouncedQuery) ||
          (h.os?.toLowerCase().includes(debouncedQuery) ?? false) ||
          tagString.includes(debouncedQuery)
        )
      })
    }

    // 🟢 Status filter
    if (statusFilter !== "all") {
      result = result.filter((h) =>
        statusFilter === "active" ? h.alive : !h.alive
      )
    }

    // 🔀 Sorting
    result.sort((a, b) => {
      const aVal = (a[sortKey] || "") as string
      const bVal = (b[sortKey] || "") as string

      if (sortDir === "asc") {
        return aVal.localeCompare(bVal)
      }
      return bVal.localeCompare(aVal)
    })

    return result
  }, [hosts, debouncedQuery, statusFilter, sortKey, sortDir])

  // =============================
  // Pagination
  // =============================
  const totalPages = Math.ceil(processedHosts.length / pageSize)

  const paginatedHosts = processedHosts.slice(
    (page - 1) * pageSize,
    page * pageSize
  )

  // =============================
  // Highlight Match
  // =============================
  function highlight(text?: string) {
    if (!text || !debouncedQuery) return text

    const regex = new RegExp(`(${debouncedQuery})`, "gi")

    return text.split(regex).map((part, i) =>
      part.toLowerCase() === debouncedQuery ? (
        <mark key={i} className="highlight">
          {part}
        </mark>
      ) : (
        part
      )
    )
  }

  function toggleSort(key: SortKey) {
    if (sortKey === key) {
      setSortDir(sortDir === "asc" ? "desc" : "asc")
    } else {
      setSortKey(key)
      setSortDir("asc")
    }
  }

  return (
    <div className="pageContainer">

      {/* ================= HEADER ================= */}
      <div className="pageHeader">
        <div>
          <h2>Hosts</h2>
          <p className="pageSubtitle">
            Manage infrastructure nodes
          </p>
        </div>

        <div className="pageSearch">
          <input
            placeholder="Search hostname, IP, OS, tags[value]..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>
      </div>

      {/* ================= STATUS FILTER ================= */}
      <div className="filterChips">
        {["all", "active", "down"].map((status) => (
          <button
            key={status}
            className={`chip ${
              statusFilter === status ? "activeChip" : ""
            }`}
            onClick={() =>
              setStatusFilter(status as any)
            }
          >
            {status.toUpperCase()}
          </button>
        ))}
      </div>

      {loading && <p>Loading...</p>}

      {/* ================= TABLE ================= */}
      <div className="tableWrap">
        <table>
          <thead>
            <tr>
              <th className="col-center">Status</th>

              <th
                className="col-center sortable"
                onClick={() => toggleSort("hostname")}
              >
                Host {sortKey === "hostname" && (sortDir === "asc" ? "↑" : "↓")}
              </th>

              <th
                className="col-center sortable"
                onClick={() => toggleSort("public_ip")}
              >
                Public IP {sortKey === "public_ip" && (sortDir === "asc" ? "↑" : "↓")}
              </th>

              {/* <th
                className="col-left sortable"
                onClick={() => toggleSort("ip")}
              >
                Internal IP {sortKey === "ip" && (sortDir === "asc" ? "↑" : "↓")}
              </th> */}

              <th
                className="col-center sortable"
                onClick={() => toggleSort("os")}
              >
                OS {sortKey === "os" && (sortDir === "asc" ? "↑" : "↓")}
              </th>

              <th className="col-center">Tags</th>

              <th
                className="col-center sortable"
                onClick={() => toggleSort("last_seen")}
              >
                Last Seen {sortKey === "last_seen" && (sortDir === "asc" ? "↑" : "↓")}
              </th>
            </tr>
          </thead>


          <tbody>
            {paginatedHosts.map((h) => {
              const tagsArray = Object.entries(h.tags || {})

              return (
                <tr
                  key={h.agent_id}
                  className="row"
                  onClick={() => navigate(`/hosts/${h.agent_id}`)}
                >
                  {/* Status */}
                  <td className="col-center">
                    <span className="status">
                      <span
                        className={`dot ${h.alive ? "good" : "bad"}`}
                      ></span>
                      {h.alive ? "Active" : "Dead"}
                    </span>
                  </td>

                  {/* Host */}
                  <td className="col-center">
                    <b>{highlight(h.hostname)}</b>
                    <div style={{ fontSize: 12, opacity: 0.6 }}>
                      {highlight(h.ip)}
                    </div>
                  </td>

                  {/* Public IP */}
                  <td className="col-center">
                    {highlight(h.public_ip) || "—"}
                  </td>

                  {/* OS */}
                  <td className="col-center">{highlight(h.os)}</td>

                  {/* Tags */}
                  <td className="col-center" style={{ position: "relative" }}>
                    {tagsArray.length > 0 ? (
                      <>
                        {tagsArray.slice(0, 3).map(([key, value]) => (
                          <span
                            key={key}
                            style={{
                              marginRight: "6px",
                              padding: "4px 8px",
                              background: "#1f2937",
                              borderRadius: "6px",
                              fontSize: "12px",
                            }}
                          >
                            {key}: {value}
                          </span>
                        ))}

                        {tagsArray.length > 3 && (
                          <span
                            className="more-tags"
                            onClick={(e) => e.stopPropagation()}
                            style={{
                              padding: "4px 8px",
                              background: "#374151",
                              borderRadius: "6px",
                              fontSize: "12px",
                              cursor: "pointer",
                            }}
                          >
                            +{tagsArray.length - 3} more

                            <div className="tooltip">
                              {tagsArray.slice(3).map(([key, value]) => (
                                <div key={key}>
                                  {key}: {value}
                                </div>
                              ))}
                            </div>
                          </span>
                        )}
                      </>
                    ) : (
                      <span style={{ opacity: 0.6 }}>No tags</span>
                    )}
                  </td>

                  {/* Last Seen */}
                  <td className="col-left">
                    {h.last_seen
                      ? new Date(h.last_seen).toLocaleString()
                      : "—"}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {/* ================= PAGINATION ================= */}
      {totalPages > 1 && (
        <div className="pagination">
          <button
            disabled={page === 1}
            onClick={() => setPage(page - 1)}
          >
            Prev
          </button>

          <span>
            Page {page} of {totalPages}
          </span>

          <button
            disabled={page === totalPages}
            onClick={() => setPage(page + 1)}
          >
            Next
          </button>
        </div>
      )}
    </div>
  )
}
