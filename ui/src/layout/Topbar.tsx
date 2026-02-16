import { useAuth } from "../auth/AuthContext"

export default function TopBar() {
  const { user, logout } = useAuth()
  console.log("TopBar User:", user)

  return (
    <div className="topbar">
      {/* LEFT SIDE */}
      <div className="topbarLeft">
        <div className="search">
          <input placeholder="Search hosts, logs..."/>
          <span className="kbd">⌘ K</span>
        </div>
      </div>

      {/* RIGHT SIDE */}
      <div className="topbarRight">
        <div className="userSection">
          <div className="signedIn">
            Signed in as <strong>{user?.username}</strong>
          </div>

          {user?.is_super_admin && (
            <span className="badge">Super Admin</span>
          )}
        </div>

        <button className="btn secondary" onClick={logout}>
          Logout
        </button>
      </div>
    </div>
  )
}