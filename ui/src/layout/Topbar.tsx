export default function Topbar() {
  return (
    <div className="topbar">
      <div className="search">
        <input placeholder="Search hosts, logs, containers..." />
      </div>

      <div className="topbar-right">
        <span className="badge">Default Account</span>
        <div className="avatar">VB</div>
      </div>
    </div>
  )
}
