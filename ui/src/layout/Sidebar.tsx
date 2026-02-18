import { NavLink } from "react-router-dom"
import { useAuth } from "../auth/AuthContext"

export default function Sidebar() {
  const { user, updateTenant } = useAuth()

  const tenants = user?.tenants || []
  const currentTenant = tenants.find(
    (t) => t.id === user?.currentTenantId
  )
  const hasPermission = (perm: string) =>
    user?.permissions?.includes(perm) || false

  const canViewAdmin = user?.is_super_admin
    // user?.permissions?.includes("users.read") ||
    // user?.permissions?.includes("roles.read") ||
    // user?.permissions?.includes("groups.read") ||

  return (
    <aside className="sidebar">
      <div className="brand">
        <div className="logo"></div>
        <div>
          <h1>OpsLens</h1>
          <p>Fleet Monitoring</p>
        </div>
      </div>

      {/* Navigation */}
      <nav className="nav">
        {/* Overview */}
        <div className="navSection">
          <p className="navTitle">Overview</p>
          <NavLink to="/fleet">Fleet Overview</NavLink>
        </div>

        {/* Infrastructure */}
        <div className="navSection">
          <p className="navTitle">Infrastructure</p>
          <NavLink to="/hosts">Hosts</NavLink>
        </div>

        {/* Observability */}
        <div className="navSection">
          <p className="navTitle">Observability</p>
          <NavLink to="/logs">Log Explorer</NavLink>
          <NavLink to="/alerts">Alerts</NavLink>
        </div>

        {/* Intelligence */}
        <div className="navSection">
          <p className="navTitle">Intelligence</p>
          <NavLink to="/ai">AI Insights</NavLink>
        </div>

        {/* Administration */}
        {canViewAdmin && (
          <div className="navSection">
            <p className="navTitle">Administration</p>

            {user?.permissions?.includes("users.read") && (
              <NavLink to="/admin/users">Users</NavLink>
            )}

            {user?.permissions?.includes("roles.read") && (
              <NavLink to="/admin/roles">Roles</NavLink>
            )}

            {user?.permissions?.includes("groups.read") && (
              <NavLink to="/admin/groups">Groups</NavLink>
            )}
          </div>
        )}
      </nav>

      {/* Tenant Switch */}
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