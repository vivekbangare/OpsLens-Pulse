import { NavLink } from "react-router-dom"

export default function Sidebar() {
  return (
    <aside className="sidebar">
      <div className="brand">
        <h1>OpsLens</h1>
        <span>Fleet Monitoring</span>
      </div>

      <nav className="nav">
        <p className="navTitle">Overview</p>

        <NavLink to="/fleet" className={({ isActive }) => isActive ? "active" : ""}>
          Fleet Overview
        </NavLink>

        <NavLink to="/alerts" className={({ isActive }) => isActive ? "active" : ""}>
          Alerts
        </NavLink>

        <p className="navTitle">Infrastructure</p>

        <NavLink to="/hosts" className={({ isActive }) => isActive ? "active" : ""}>
          Hosts
        </NavLink>

        <NavLink to="/logs" className={({ isActive }) => isActive ? "active" : ""}>
          Log Explorer
        </NavLink>

        <p className="navTitle">Intelligence</p>

        <NavLink to="/ai" className={({ isActive }) => isActive ? "active" : ""}>
          AI Insights
        </NavLink>

         <p className="navTitle">Admin</p>

        <NavLink to="/admin" className={({ isActive }) => isActive ? "active" : ""}>
          Admin
        </NavLink>
      </nav>
    </aside>
  )
}