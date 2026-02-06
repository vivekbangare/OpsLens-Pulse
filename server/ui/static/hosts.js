const tbody = document.querySelector("#hosts tbody")
const searchInput = document.getElementById("searchInput")
const tagFilterInput = document.getElementById("tagFilterInput")
const logoutBtn = document.getElementById("logoutBtn")

// redirect if not logged in
if (!sessionStorage.getItem("loggedIn")) window.location.href = "/index.html"

let hostsData = []
const API_TOKEN = sessionStorage.getItem("apiKey")

if (!API_TOKEN) {
  window.location.href = "/index.html"
}

const keyInfo = document.getElementById("apiKeyInfo")
if (keyInfo) {
  keyInfo.textContent = "API Key: ****" + API_TOKEN.slice(-4)
}

function timeAgo(date) {
  const seconds = Math.floor((Date.now() - new Date(date)) / 1000)
  if (seconds < 10) return "just now"
  if (seconds < 60) return `${seconds}s ago`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  return `${Math.floor(seconds / 3600)}h ago`
}

function renderTags(tags) {
  if (!tags || Object.keys(tags).length === 0) return "<span style='color:#999'>—</span>"
  return Object.entries(tags)
    .map(([k,v]) => `<span class="tag"><span class="tag-key">${k}</span>: ${v}</span>`)
    .join("")
}

function renderHostsList(data) {
  tbody.innerHTML = ""
  data.forEach(h => {
    const tr = document.createElement("tr")
    tr.innerHTML = `
      <td class="hostname">${h.hostname}</td>
      <td class="${h.alive ? "alive" : "dead"}">${h.alive ? "Alive" : "Down"}</td>
      <td>${timeAgo(h.last_seen)}</td>
    `
    tbody.appendChild(tr)

    tr.querySelector(".hostname").addEventListener("click", () => showDetails(h))
  })
}

function showDetails(host) {
  document.getElementById("hosts").style.display = "none"
  document.getElementById("hostDetails").style.display = "block"

  document.getElementById("detailsHostname").textContent = host.hostname
  document.getElementById("detailsIP").textContent = host.ip || "—"
  document.getElementById("detailsOS").textContent = host.os
  document.getElementById("detailsCPU").textContent = host.cpu_percent.toFixed(1)
  document.getElementById("detailsMem").textContent = host.mem_used_mb
  document.getElementById("detailsUptime").textContent = host.uptime_sec
  document.getElementById("detailsTags").innerHTML = renderTags(host.tags)
}

document.getElementById("backBtn").addEventListener("click", () => {
  document.getElementById("hostDetails").style.display = "none"
  document.getElementById("hosts").style.display = "table"
})

// Filtering logic by hostname and tag key=value
function applyFilter() {
  const search = searchInput.value.toLowerCase().trim()
  const tagFilter = tagFilterInput.value.trim() // format: key=value

  const filtered = hostsData.filter(h => {
    // Check hostname match
    const matchesSearch = h.hostname.toLowerCase().includes(search)

    // Check tag filter
    let matchesTag = true
    if (tagFilter) {
      const [key, value] = tagFilter.split('=')
      if (!key || !value || !h.tags) {
        matchesTag = false
      } else {
        matchesTag = h.tags[key] === value
      }
    }

    return matchesSearch && matchesTag
  })

  renderHostsList(filtered)
}

searchInput.addEventListener("input", applyFilter)
tagFilterInput.addEventListener("input", applyFilter)

logoutBtn.addEventListener("click", () => {
  sessionStorage.removeItem("loggedIn")
  sessionStorage.removeItem("apiKey")
  window.location.href = "/index.html"
})

function loadHosts() {
  fetch("/api/hosts", {
    headers: {
      "Authorization": "Bearer " + API_TOKEN
    }
  })
    .then(res => {
      if (!res.ok) throw new Error("Unauthorized")
      return res.json()
    })
    .then(data => {
      hostsData = data
      applyFilter() // apply filter immediately after fetch
    })
    .catch(err => console.error("Failed to load hosts:", err))
}

// Initial load and periodic refresh
loadHosts()
setInterval(loadHosts, 5000)
