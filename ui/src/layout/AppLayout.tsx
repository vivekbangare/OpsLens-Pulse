import Sidebar from "./Sidebar"
import Topbar from "./Topbar"

export default function AppLayout({ children }: any) {
  return (
    <div className="app">
      <Sidebar />
      <main className="main">
        <Topbar />
        {children}
      </main>
    </div>
  )
}
