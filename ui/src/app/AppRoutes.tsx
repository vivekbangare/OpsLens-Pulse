import { Routes, Route, Navigate } from "react-router-dom"

import AppLayout from "../layout/AppLayout"
import Login from "../features/auth/Login"
import FleetPage from "../features/fleet/FleetPage"
import Hosts from "../features/hosts/Hosts"
import HostDetails from "../features/hosts/HostDetails"
import LogsExplorer from "../features/logs/LogExplores"
import Alerts from "../features/alerts/Alerts"
import ProtectedRoute from "./ProtectedRoute"
import AIPage from "../features/ai/AIPage"
import { AdminRoutes } from "../features/admin/routes";

export default function AppRoutes() {
  return (
    <Routes>
      
      <Route path="/login" element={<Login />} />

      {/* Protected */}
      <Route element={<ProtectedRoute />}>
        <Route path="/" element={<AppLayout />}>
          <Route index element={<Navigate to="/fleet" replace />} />
          <Route path="fleet" element={<FleetPage />} />
          <Route path="hosts" element={<Hosts />} />
          <Route path="hosts/:agentId" element={<HostDetails />} />
          <Route path="logs" element={<LogsExplorer />} />
          <Route path="alerts" element={<Alerts />} />
          <Route path="ai" element={<AIPage />} />
          {AdminRoutes}
        </Route>
      </Route>

      {/* Fallback */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
