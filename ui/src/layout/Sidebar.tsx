import { NavLink } from "react-router-dom"
import { useAuth } from "../auth/AuthContext"

export default function Sidebar() {
  const { user, updateTenant } = useAuth()

  const tenants = user?.tenants || []
  const currentTenant = tenants.find(
    (t) => t.id === user?.currentTenantId
  )

  return (
    <aside className="sidebar">
      <div className="brand">
        <div className="logo"></div>
        <div>
          <h1>OpsLens</h1>
          <p>Fleet Monitoring</p>
        </div>
      </div>

      <nav className="nav">
        <NavLink to="/fleet">⚡ Fleet Overview</NavLink>
        <NavLink to="/hosts">🧩 Hosts</NavLink>
        <NavLink to="/logs">🧾 Logs Explorer</NavLink>
        <NavLink to="/alerts">🚨 Alerts</NavLink>
        <NavLink to="/admin">🛡️ Admin</NavLink>
      </nav>

      {/* Tenant Footer */}
      <div className="sidebarFooter">
        {tenants.length <= 1 ? (
          <div className="tenantName">
            {currentTenant?.name || "Default Tenant"}
          </div>
        ) : (
          <select
            className="tenantSelect"
            value={user?.currentTenantId}
            onChange={(e) => updateTenant(e.target.value)}
          >
            {tenants.map((t) => (
              <option key={t.id} value={t.id}>
                {t.name}
              </option>
            ))}
          </select>
        )}
      </div>
    </aside>
  )
}