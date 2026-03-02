import { useState, useEffect } from "react"
import GlobalSearch from "../shared/components/GlobalSearch"
import { useAuth } from "../features/auth/AuthContext"
import { useNavigate } from "react-router-dom"
import {
  Bell,
  HelpCircle,
  Moon,
  Sun,
  ChevronDown,
} from "lucide-react"

export default function TopBar() {
  const [isDark, setIsDark] = useState(true)
  const [open, setOpen] = useState(false)
  const { logout } = useAuth()
  const navigate = useNavigate()

  // Load saved theme
  useEffect(() => {
    const saved = localStorage.getItem("theme")

    if (saved === "light") {
      document.documentElement.classList.remove("dark")
      document.documentElement.classList.add("light")
      setIsDark(false)
    } else {
      document.documentElement.classList.remove("light")
      document.documentElement.classList.add("dark")
      setIsDark(true)
    }
  }, [])

  function toggleTheme() {
    const html = document.documentElement
    const current = html.getAttribute("data-theme")

    if (current === "dark") {
      html.setAttribute("data-theme", "light")
      localStorage.setItem("theme", "light")
      setIsDark(false)
    } else {
      html.setAttribute("data-theme", "dark")
      localStorage.setItem("theme", "dark")
      setIsDark(true)
    }
  }

  function handleLogout() {
    logout()
    navigate("/login", { replace: true })
  }

  return (
    <div className="topbar">

      {/* 🔍 Global Search (Left Side) */}
      <div className="topbarLeft">
         <GlobalSearch />
      </div>

      {/* Right Section */}
      <div className="topbarRight">

        {/* Help */}
        <div className="topIcon">
          <HelpCircle size={18} />
        </div>

        {/* Notifications */}
        <div className="topIcon notification">
          <Bell size={18} />
          <span className="badge">3</span>
        </div>

        {/* Theme Toggle */}
        <div
          className="topIcon theme"
          onClick={toggleTheme}
        >
          {isDark ? (
            <Sun size={18} />
          ) : (
            <Moon size={18} />
          )}
        </div>

        {/* User Dropdown */}
        <div
          className="userSection"
          onClick={() => setOpen(!open)}
        >
          <div className="avatar">V</div>
          <span>Vivek</span>
          <ChevronDown size={16} />

          {open && (
            <div className="dropdown">
              <div>Profile</div>
              <div>Settings</div>
              <div className="danger" onClick={handleLogout}>
                Logout
              </div>
            </div>
          )}
        </div>

      </div>
    </div>
  )
}