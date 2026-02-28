import Sidebar from "./Sidebar"
import Topbar from "./Topbar"
import { Outlet } from "react-router-dom"

export default function AppLayout() {
  return (
    <div className="layout">
      <Sidebar />
      <div className="contentArea">
        <Topbar />
        <div className="pageContainer">
          <Outlet />
        </div>
      </div>
    </div>
  )
}
