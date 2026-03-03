import { useEffect, useState, useMemo } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { formatUptime } from "../../shared/lib/utils/time"

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
  tags?: Tags | null
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

  const maxLogs = 200

  useEffect(() => {
    loadAll()
  }, [agentId])

  async function loadAll() {
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

      const sortedLogs = Array.isArray(logData)
        ? [...logData]
            .sort((a, b) => (b.timestamp ?? 0) - (a.timestamp ?? 0))
            .slice(0, maxLogs)
        : []

      setSummary(summaryData ?? null)
      setLogs(sortedLogs)
      setContainers(Array.isArray(containerData) ? containerData : [])

    } catch (err) {
      console.error("Unexpected error:", err)
    }
  }

  async function refreshLogs() {
    const logData = await fetchLogs(id)

    const sortedLogs = Array.isArray(logData)
      ? [...logData]
          .sort((a, b) => (b.timestamp ?? 0) - (a.timestamp ?? 0))
          .slice(0, maxLogs)
      : []

    setLogs(sortedLogs)
  }

  function formatDate(ts?: number) {
    if (!ts) return "—"
    return new Date(ts * 1000).toLocaleString()
  }

  function truncate(message?: string) {
    if (!message) return ""
    return message.length > 300
      ? message.slice(0, 300) + "..."
      : message
  }

  function copyLogs() {
    const text = logs
      .map(
        l =>
          `${formatDate(l.timestamp)} [${l.level}] ${l.source_name} - ${l.message}`
      )
      .join("\n")

    navigator.clipboard.writeText(text)
  }

  function downloadLogs() {
    const text = logs
      .map(
        l =>
          `${formatDate(l.timestamp)} [${l.level}] ${l.source_name} - ${l.message}`
      )
      .join("\n")

    const blob = new Blob([text], { type: "text/plain" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `host-${agentId}-logs.txt`
    a.click()
    URL.revokeObjectURL(url)
  }

  // Last 1h summary
  const now = Date.now() / 1000
  const lastHourLogs = logs.filter(
    l => l.timestamp && l.timestamp >= now - 3600
  )

  const errorCount = lastHourLogs.filter(l => l.level === "error").length
  const warnCount = lastHourLogs.filter(l => l.level === "warn").length
  const infoCount = lastHourLogs.filter(l => l.level === "info").length

  function levelColor(level?: string) {
    if (level === "error") return "#ef4444"
    if (level === "warn") return "#f59e0b"
    if (level === "critical") return "#dc2626"
    return "#22c55e"
  }

  return (
    <div className="pageContainer">

      <div className="detailsHeader">
        <div>
          <h2>Host Details</h2>
          <div className="subText">Agent ID: {agentId}</div>
        </div>

        <button className="filterBtn" onClick={() => navigate(-1)}>
          ← Back
        </button>
      </div>

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
          {formatUptime(summary?.uptime_sec)}
        </Metric>
      </div>

      {/* LOGS */}
      <div className="cardPanel compactCard">
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "12px",
          }}
        >
          <h3>Logs (Last {maxLogs})</h3>

          <div className="logActions">
            <button className="btn secondary" onClick={refreshLogs}>
              Refresh
            </button>

            <button className="btn secondary" onClick={copyLogs}>
              Copy
            </button>

            <button className="btn secondary" onClick={downloadLogs}>
              Download
            </button>

            <button
              className="btn primary"
              onClick={() => navigate(`/logs?agent_id=${agentId}`)}
            >
              View in Log Explorer →
            </button>
          </div>
        </div>

        {/* Last 1h Summary */}
        <div
          style={{
            marginBottom: "16px",
            fontWeight: "bold",
          }}
        >
          Last 1h:
          <span style={{ color: "#ef4444", marginLeft: "12px" }}>
            {errorCount} Errors
          </span>
          <span style={{ color: "#f59e0b", marginLeft: "12px" }}>
            {warnCount} Warnings
          </span>
          <span style={{ color: "#3b82f6", marginLeft: "12px" }}>
            {infoCount} Info
          </span>
        </div>

        <div
          style={{
            maxHeight: "400px",
            overflowY: "auto",
            paddingRight: "8px",
          }}
        >
          {logs.length === 0 && (
            <div className="subText">No logs available</div>
          )}

          {logs.map((l, i) => (
            <div
              key={i}
              style={{
                marginBottom: "16px",
                borderBottom: "1px solid rgba(255,255,255,0.05)",
                paddingBottom: "10px",
              }}
            >
              <div
                style={{
                  display: "flex",
                  gap: "16px",
                  fontSize: "12px",
                  marginBottom: "4px",
                  alignItems: "center",
                }}
              >
                <span style={{ color: "#9ca3af", minWidth: "160px" }}>
                  {formatDate(l.timestamp)}
                </span>

                <span
                  style={{
                    color: levelColor(l.level),
                    fontWeight: 600,
                    minWidth: "80px",
                    textTransform: "uppercase",
                  }}
                >
                  {l.level}
                </span>

                <span style={{ color: "#60a5fa" }}>
                  {l.source_name ?? "unknown"}
                </span>
              </div>

              <div
                style={{
                  fontFamily: "monospace",
                  fontSize: "13px",
                  whiteSpace: "pre-wrap",
                  wordBreak: "break-word",
                }}
              >
                {truncate(l.message)}
              </div>
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
              <div key={i} style={{ marginBottom: "8px" }}>
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