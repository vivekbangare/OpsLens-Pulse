import { getAgentHealth } from "../utils/health"
import StatusDots from "./StatusDots"
import { formatLastSeen } from "../utils/time"

interface Host {
  agent_id: string
  hostname: string
  ip: string
  public_ip?: string
  os?: string
  last_seen?: string
  alive: boolean
  tags?: Record<string, string>
  cpu_percent?: number
  mem_used_mb?: number
  mem_total_mb?: number
  cloud_provider?: string
}

interface Props {
  hosts: Host[] | null | undefined
  onSelect: (id: string) => void
}

export default function HostTable({ hosts, onSelect }: Props) {
  if (!Array.isArray(hosts) || hosts.length === 0) {
    return null
  }

  return (
    <table className="hostTable">
      <thead>
        <tr>
          <th>Status</th>
          <th>Host</th>
          <th>Public IP</th>
          <th>OS</th>
          <th>Cloud</th>
          <th>Tags</th>
          <th>Last Seen</th>
        </tr>
      </thead>

      <tbody>
        {hosts.map((h) => {
          const health = getAgentHealth(h)

          const memPercent =
            h.mem_used_mb && h.mem_total_mb
              ? ((h.mem_used_mb / h.mem_total_mb) * 100).toFixed(0)
              : null

          return (
            <tr
              key={h.agent_id}
              className="hostRow"
              onClick={() => onSelect(h.agent_id)}
            >

              {/* ================= STATUS (Dual Dot with Tooltips) ================= */}
              <td>
                <div className="statusDualWrapper">

                  {/* ================= HOST DOT ================= */}
                  <div className="dotWrapper">
                    <span
                      className={`dot ${h.alive ? "good" : "bad"}`}
                    ></span>

                    <div className="dotTooltip">
                      <div className="tooltipTitle">Host</div>
                      <div className="tooltipRow">
                        Status: {h.alive ? "Online" : "Offline"}
                      </div>

                      {h.cpu_percent !== undefined && (
                        <div className="tooltipRow">
                          CPU: {h.cpu_percent.toFixed(0)}%
                        </div>
                      )}

                      {memPercent && (
                        <div className="tooltipRow">
                          Memory: {memPercent}%
                        </div>
                      )}
                    </div>
                  </div>

                  {/* ================= AGENT DOT ================= */}
                  <div className="dotWrapper">
                    <span
                      className={`dot ${health.className}`}
                    ></span>

                    <div className="dotTooltip">
                      <div className="tooltipTitle">Agent</div>
                      <div className="tooltipRow">
                        Status: {health.label}
                      </div>

                      <div className="tooltipReason">
                        {health.reason}
                      </div>
                    </div>
                  </div>

                </div>
              </td>

              {/* ================= HOST ================= */}
              <td>
                <div className="hostCell">
                  <div className="hostName">{h.hostname}</div>
                  <div className="hostSub">{h.ip}</div>
                </div>
              </td>

              {/* ================= PUBLIC IP ================= */}
              <td>{h.public_ip || "—"}</td>

              {/* ================= OS ================= */}
              <td>{h.os || "—"}</td>

              {/* ================= CLOUD ================= */}
              <td>
                <span className="cloudBadge">
                  {h.cloud_provider || "—"}
                </span>
              </td>

              {/* ================= CLOUD ================= */}
              <td>
                {h.tags && typeof h.tags === "object" ? (
                  <div className="tagsWrapper">
                    {Object.entries(h.tags as Record<string, string>)
                      .slice(0, 3)
                      .map(([key, value]) => (
                        <span key={key} className="tagPill">
                          {String(key)}: {String(value)}
                        </span>
                      ))}

                    {Object.keys(h.tags).length > 3 && (
                      <div className="moreTagsWrapper">
                        <span className="tagMore">
                          +{Object.keys(h.tags).length - 3}
                        </span>

                        <div className="tagsTooltip">
                          {Object.entries(h.tags as Record<string, string>)
                            .slice(3)
                            .map(([key, value]) => (
                              <div key={key} className="tooltipRow">
                                {String(key)}: {String(value)}
                              </div>
                            ))}
                        </div>
                      </div>
                    )}
                  </div>
                ) : (
                  "—"
                )}
              </td>

              {/* ================= LAST SEEN ================= */}
              <td>
                {h.last_seen ? (
                  (() => {
                    const { relative, absolute } = formatLastSeen(h.last_seen)
                    return (
                      <div className="lastSeenCell">
                        <div className="lastSeenRelative">{relative}</div>
                        <div className="lastSeenAbsolute">{absolute}</div>
                      </div>
                    )
                  })()
                ) : (
                  "—"
                )}
              </td>

            </tr>
          )
        })}
      </tbody>
    </table>
  )
}