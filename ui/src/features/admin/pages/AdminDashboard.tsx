import StatCard from "../components/StatCard";
import { useNavigate } from "react-router-dom";

export default function AdminDashboard() {
  const navigate = useNavigate();

  return (
    <div className="admin-page">
      <h1>Admin</h1>

      <div className="stat-grid">
        <StatCard
          title="Users"
          count={14}
          subtitle="users"
          onClick={() => navigate("/admin/users")}
        />
        <StatCard
          title="Groups"
          count={5}
          subtitle="groups"
          onClick={() => navigate("/admin/groups")}
        />
        <StatCard
          title="Tenants"
          count={3}
          subtitle="tenants"
          onClick={() => navigate("/admin/tenants")}
        />
      </div>
    </div>
  );
}