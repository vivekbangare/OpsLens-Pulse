import { NavLink, Routes, Route } from "react-router-dom"


export default function Admin() {
  return (
    <div style={{ padding: "20px" }}>
      <h2>Admin Panel</h2>

      <div className="card" style={{ padding: "20px", marginTop: "20px" }}>
        <h3>Tenants</h3>
        <p>Tenant management UI coming soon.</p>
      </div>

      <div className="card" style={{ padding: "20px", marginTop: "20px" }}>
        <h3>Users</h3>
        <p>User management UI coming soon.</p>
      </div>

      <div className="card" style={{ padding: "20px", marginTop: "20px" }}>
        <h3>Roles</h3>
        <p>Role management UI coming soon.</p>
      </div>
    </div>
  )
}

