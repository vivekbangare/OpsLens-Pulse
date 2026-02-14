export default function Drawer({ host, onClose }: any) {
  return (
    <>
      <div className="backdrop show" onClick={onClose}></div>

      <aside className="drawer open">
        <div className="drawerHeader">
          <div>
            <h3>{host.hostname}</h3>
            <div className="meta">
              {host.os} · {host.ip}
            </div>
          </div>
          <button className="closeBtn" onClick={onClose}>
            Close
          </button>
        </div>

        <div className="drawerBody">
          <div className="panel">
            <h4>Summary</h4>
            <p>Status: {host.alive ? "Alive" : "Dead"}</p>
          </div>
        </div>
      </aside>
    </>
  )
}
