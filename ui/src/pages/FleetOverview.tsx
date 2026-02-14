import { useEffect, useState } from "react"
import { fetchHosts } from "../api/client"
import Drawer from "../components/Drawer"

export default function FleetOverview() {
  const [hosts, setHosts] = useState<any[]>([])
  const [selected, setSelected] = useState<any | null>(null)

  useEffect(() => {
    load()
    const interval = setInterval(load, 5000)
    return () => clearInterval(interval)
  }, [])

  async function load() {
    const data = await fetchHosts()
    setHosts(data)
  }

  return (
    <>
      <section className="grid">

        <div className="card">
          <div className="cardHeader">
            <h2>Fleet Health</h2>
          </div>

          <div className="tableWrap">
            <table>
              <thead>
                <tr>
                  <th>Status</th>
                  <th>Host</th>
                  <th>OS</th>
                  <th>Last Seen</th>
                </tr>
              </thead>
              <tbody>
                {hosts.map((h) => (
                  <tr
                    key={h.agent_id}
                    className="row"
                    onClick={() => setSelected(h)}
                  >
                    <td>
                      <span className="status">
                        <span
                          className={`dot ${
                            h.alive ? "good" : "bad"
                          }`}
                        ></span>
                        {h.alive ? "Active" : "Dead"}
                      </span>
                    </td>
                    <td>
                      <b>{h.hostname}</b>
                      <div style={{ fontSize: 12, opacity: .6 }}>
                        {h.ip}
                      </div>
                    </td>
                    <td>{h.os}</td>
                    <td>
                      {new Date(h.last_seen).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

      </section>

      {selected && (
        <Drawer host={selected} onClose={() => setSelected(null)} />
      )}
    </>
  )
}
