export default function Sidebar() {
  return (
    <aside className="sidebar">
      <div className="brand">
        <div className="logo"></div>
        <div>
          <h1>OpsLens</h1>
          <p>Multi-tenant fleet monitoring</p>
        </div>
      </div>

      <nav className="nav">
        <a className="active">⚡ Fleet Overview</a>
        <a>🧩 Hosts</a>
        <a>🧾 Logs Explorer</a>
        <a>🚨 Alerts</a>
        <a>🛡️ Admin</a>
      </nav>
    </aside>
  )
}
