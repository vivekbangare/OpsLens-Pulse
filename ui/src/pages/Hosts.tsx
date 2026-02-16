import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"
import { fetchHosts } from "../api/client"

type Host = {
  agent_id: string
  alive: boolean
  hostname: string
  ip: string
  os?: string
  last_seen?: string
  tags?: Record<string, string>
}

export default function Hosts() {
  const [hosts, setHosts] = useState<Host[]>([])
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    load()
    const interval = setInterval(load, 5000)
    return () => clearInterval(interval)
  }, [])

  async function load() {
    const data = await fetchHosts()
    setHosts(Array.isArray(data) ? data : [])
    setLoading(false)
  }

  return (
    <div style={{ padding: "20px" }}>
      <h2>Hosts</h2>

      {loading && <p>Loading...</p>}

      <div className="tableWrap" style={{ marginTop: "20px" }}>
        <table>
          <thead>
            <tr>
              <th>Status</th>
              <th>Host</th>
              <th>OS</th>
              <th>Tags</th>
              <th>Last Seen</th>
            </tr>
          </thead>

          <tbody>
            {hosts.map((h) => {
              const tagsArray = Object.entries(h.tags || {})

              return (
                <tr
                  key={h.agent_id}
                  className="row"
                  onClick={() => navigate(`/hosts/${h.agent_id}`)}
                  style={{ cursor: "pointer" }}
                >
                  {/* Status */}
                  <td>
                    <span className="status">
                      <span
                        className={`dot ${h.alive ? "good" : "bad"}`}
                      ></span>
                      {h.alive ? "Active" : "Dead"}
                    </span>
                  </td>

                  {/* Hostname + IP */}
                  <td>
                    <b>{h.hostname}</b>
                    <div style={{ fontSize: 12, opacity: 0.6 }}>
                      {h.ip}
                    </div>
                  </td>

                  {/* OS */}
                  <td>{h.os || "—"}</td>

                  {/* Tags */}
                  <td style={{ position: "relative" }}>
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
                            style={{
                              marginRight: "6px",
                              padding: "4px 8px",
                              background: "#374151",
                              borderRadius: "6px",
                              fontSize: "12px",
                              cursor: "pointer",
                            }}
                            className="more-tags"
                            onClick={(e) => e.stopPropagation()} // prevent row click
                          >
                            +{tagsArray.length - 3} more

                            <div className="tooltip">
                              {tagsArray.slice(3).map(([key, value]) => (
                                <div key={key} style={{ marginBottom: "4px" }}>
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
                  <td>
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
    </div>
  )
}