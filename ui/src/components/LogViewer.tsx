interface Props {
  logs: any[] | null | undefined
}

export default function LogViewer({ logs }: Props) {
  if (!Array.isArray(logs) || logs.length === 0) {
    return (
      <div className="log-box">
        No logs available
      </div>
    )
  }

  return (
    <div className="log-box">
      {logs.map((l, i) => (
        <div key={i}>
          {l.timestamp
            ? new Date(l.timestamp * 1000).toLocaleString()
            : "—"}{" "}
          [{l.level || "info"}] {l.message || ""}
        </div>
      ))}
    </div>
  )
}
