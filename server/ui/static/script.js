const tbody = document.querySelector("#hosts tbody")

function timeAgo(date) {
  const seconds = Math.floor((Date.now() - new Date(date)) / 1000)

  if (seconds < 10) return "just now"
  if (seconds < 60) return `${seconds}s ago`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  return `${Math.floor(seconds / 3600)}h ago`
}

function renderHosts(data) {
  tbody.innerHTML = ""

  data.forEach(h => {
    const tr = document.createElement("tr")

    tr.innerHTML = `
      <td>${h.hostname}</td>
      <td>${h.cpu_percent.toFixed(1)}%</td>
      <td>${h.mem_used_mb} MB</td>
      <td class="${h.alive ? 'alive' : 'dead'}">
        <span class="dot ${h.alive ? 'alive' : 'dead'}"></span>
        ${h.alive ? 'Alive' : 'Down'}
      </td>
      <td>${timeAgo(h.last_seen)}</td>
    `
    tbody.appendChild(tr)
  })
}

function loadHosts() {
  fetch("/api/hosts")
    .then(res => res.json())
    .then(renderHosts)
    .catch(err => console.error("Failed to load hosts", err))
}

// initial load
loadHosts()

// 🔁 refresh every 5 seconds
setInterval(loadHosts, 5000)
