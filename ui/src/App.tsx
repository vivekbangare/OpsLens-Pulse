import { Routes, Route, Navigate, useParams } from "react-router-dom"
import { useAuth } from "./auth/AuthContext"

import Login from "./pages/Login"
import FleetOverview from "./pages/FleetOverview"
import Hosts from "./pages/Hosts"
import LogsExplorer from "./pages/logs/LogsExplorer"
import Alerts from "./pages/Alerts"
import Admin from "./pages/Admin"
import HostDetails from "./pages/HostDetails"
import AppLayout from "./layout/AppLayout"

export default function App() {
  const { token } = useAuth()

  return (
    <Routes>
      {/* ---------------- PUBLIC ---------------- */}
      {!token && (
        <Route path="*" element={<Login />} />
      )}

      {/* ---------------- PROTECTED ---------------- */}
      {token && (
        <Route path="/" element={<AppLayout />}>
          <Route index element={<Navigate to="/fleet" replace />} />
          <Route path="fleet" element={<FleetOverview />} />
          <Route path="hosts" element={<Hosts />} />
          <Route path="hosts/:agentId" element={<HostDetailsWrapper />} />
          <Route path="logs" element={<LogsExplorer />} />
          <Route path="alerts" element={<Alerts />} />
          <Route path="admin" element={<Admin />} />
          <Route path="*" element={<Navigate to="/fleet" replace />} />
        </Route>
      )}
    </Routes>
  )
}

function HostDetailsWrapper() {
  const { agentId } = useParams()

  if (!agentId) return null

  return <HostDetails agentId={agentId} />
}
