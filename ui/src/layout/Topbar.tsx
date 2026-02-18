import { useState } from "react"
import { useAuth } from "../auth/AuthContext"
import { useSearch } from "../context/SearchContext"

export default function TopBar() {
  const { user, logout } = useAuth()
  const [open, setOpen] = useState(false)
  const { config } = useSearch()

  return (
    <div className="topbar">

      {/* <div className="topbarLeft">
        {config && (
          <div className="search">
            <input
              placeholder={config.placeholder}
              onChange={(e) => config.onSearch(e.target.value)}
            />
          </div>
        )}
      </div> */}

      {/* RIGHT */}
      <div className="topbarRight">

        {/* Notification */}
        <div className="notification">🔔</div>

        {/* User Dropdown */}
        <div className="userWrapper" onClick={() => setOpen(!open)}>
          <div className="avatar">
            {user?.username?.charAt(0).toUpperCase()}
          </div>

          <span className="username">{user?.username}</span>

          {open && (
            <div className="dropdown">

              <div className="dropdownItem">Profile</div>
              <div className="dropdownItem">Settings</div>
              <div className="dropdownSub">
                <div className="dropdownItem">API Tokens</div>
                <div className="dropdownItem">Change Password</div>
              </div>

              <div className="divider"></div>

              <div className="dropdownItem logout" onClick={logout}>
                Logout
              </div>

            </div>
          )}
        </div>

      </div>
    </div>
  )
}
