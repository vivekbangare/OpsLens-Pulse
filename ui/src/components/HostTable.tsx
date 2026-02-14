interface Props {
  hosts: any[] | null | undefined
  onSelect: (id: string) => void
}

export default function HostTable({ hosts, onSelect }: Props) {
  if (!Array.isArray(hosts) || hosts.length === 0) {
    return null
  }

  return (
    <table>
      <thead>
        <tr>
          <th>Host</th>
          <th>Status</th>
          <th>Last Seen</th>
          <th>IP</th>
          <th>OS</th>
        </tr>
      </thead>
      <tbody>
        {hosts.map((h) => (
          <tr key={h.agent_id} onClick={() => onSelect(h.agent_id)}>
            <td>{h.hostname}</td>
            <td className={h.alive ? "alive" : "dead"}>
              {h.alive ? "Alive" : "Down"}
            </td>
            <td>{h.last_seen ? new Date(h.last_seen).toLocaleString() : "—"}</td>
            <td>{h.ip || "—"}</td>
            <td>{h.os || "—"}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
