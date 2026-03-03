import { NavLink, Outlet } from "react-router-dom";

export default function AdminLayout() {
  return (
    <div className="adminContainer">
      <div className="adminHeader">
        <h1>Admin</h1>
        <p className="adminSub">
          Manage users, groups and tenants
        </p>
      </div>

      <div className="adminTabs">
        <NavLink to="/admin/users">Users</NavLink>
        <NavLink to="/admin/groups">Groups</NavLink>
        <NavLink to="/admin/tenants">Tenants</NavLink>
      </div>

      <div className="adminContent">
        <Outlet />
      </div>
    </div>
  );
}