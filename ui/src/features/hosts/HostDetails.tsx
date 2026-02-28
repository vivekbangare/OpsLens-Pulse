import { useEffect, useState } from "react"
import { useNavigate, useParams } from "react-router-dom"

import {
  fetchLogs,
  fetchHostSummary,
  fetchHosts,
  fetchContainerMetrics,
} from "./api"

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
  source_name?: string
  source_type?: string
}

interface ContainerMetric {
  name: string
  status: string
}

export default function HostDetails() {
  const navigate = useNavigate()
  const { agentId } = useParams<{ agentId: string }>()
  const id = agentId ?? ""

  if (!id) return null

  const [logs, setLogs] = useState<LogEntry[]>([])
  const [summary, setSummary] = useState<HostSummary | null>(null)
  const [containers, setContainers] = useState<ContainerMetric[]>([])
  const [loading, setLoading] = useState(true)

  const [selectedSource, setSelectedSource] = useState("all")
  const [selectedType, setSelectedType] = useState("all")

  useEffect(() => {
    loadAll()
    const interval = setInterval(loadAll, 5000)
    return () => clearInterval(interval)
  }, [agentId])

  async function loadAll() {
    setLoading(true)
    try {
      const [summaryData, logData, containerData, hostsData] =
        await Promise.all([
          fetchHostSummary(id),
          fetchLogs(id),
          fetchContainerMetrics(id).catch(() => []),
          fetchHosts(),
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

  const filteredLogs = logs.filter(l => {
    const sourceMatch =
      selectedSource === "all" || l.source_name === selectedSource

    const typeMatch =
      selectedType === "all" || l.source_type === selectedType

    return sourceMatch && typeMatch
  })

  return (
    <div className="pageContainer">

      {/* HEADER */}
      <div className="detailsHeader">
        <div>
          <h2>Host Details</h2>
          <div className="subText">Agent ID: {agentId}</div>
        </div>

        <button className="filterBtn" onClick={() => navigate(-1)}>
          ← Back
        </button>
      </div>

      {/* METRICS GRID */}
      <div className="metricsGridCompact">

        <Metric label="CPU">
          {summary?.cpu_percent?.toFixed(1) ?? "—"} %
        </Metric>

        <Metric label="RAM">
          {summary?.mem_used_mb ?? "—"} /{" "}
          {summary?.mem_total_mb ?? "—"} MB
        </Metric>

        <Metric label="Disk">
          {summary?.disk_used_mb ?? "—"} MB
        </Metric>

        <Metric label="Net In">
          {summary?.network_in_mb ?? "—"} MB
        </Metric>

        <Metric label="Net Out">
          {summary?.network_out_mb ?? "—"} MB
        </Metric>

        <Metric label="Uptime">
          {summary?.uptime_sec ?? "—"}
        </Metric>

      </div>

      {/* TAGS */}
      <div className="cardPanel compactCard">
        <strong>Tags:</strong>{" "}
        {summary?.tags && Object.keys(summary.tags).length > 0
          ? Object.entries(summary.tags).map(([k, v]) => (
              <span key={k} className="tagChip">
                {k}: {v}
              </span>
            ))
          : "No tags"}
      </div>

      {/* LOGS */}
      <div className="cardPanel compactCard">
        <div className="logsHeaderRow">
          <h3>Logs</h3>

          <div className="logsFiltersCompact">
            <select
              className="selectCompact"
              value={selectedSource}
              onChange={(e) => setSelectedSource(e.target.value)}
            >
              <option value="all">All Sources</option>
              {Array.from(new Set(logs.map(l => l.source_name).filter(Boolean)))
                .map((src) => (
                  <option key={src} value={src}>
                    {src}
                  </option>
                ))}
            </select>

            <select
              className="selectCompact"
              value={selectedType}
              onChange={(e) => setSelectedType(e.target.value)}
            >
              <option value="all">All Types</option>
              {Array.from(new Set(logs.map(l => l.source_type).filter(Boolean)))
                .map((t) => (
                  <option key={t} value={t}>
                    {t}
                  </option>
                ))}
            </select>
          </div>
        </div>

        <div className="logsCompact">
          {filteredLogs.length === 0 && (
            <div className="subText">No logs available</div>
          )}

          {filteredLogs.map((l, i) => (
            <div key={i} className="logLineCompact">
              [{l.level}] {l.message}
            </div>
          ))}
        </div>
      </div>

      {/* CONTAINERS */}
      <div className="cardPanel compactCard">
        <h3>Containers</h3>
        {containers.length === 0
          ? <div className="subText">No containers detected</div>
          : containers.map((c, i) => (
              <div key={i} className="logLineCompact">
                {c.name} — {c.status}
              </div>
            ))}
      </div>

    </div>
  )
}

function Metric({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) {
  return (
    <div className="metricCompact">
      <div className="metricLabel">{label}</div>
      <div className="metricValue">{children}</div>
    </div>
  )
}