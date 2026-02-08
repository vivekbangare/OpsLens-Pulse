/*********************************
 * DOM ELEMENTS
 *********************************/
const tbody = document.querySelector("#hosts tbody")
const searchInput = document.getElementById("searchInput")
const tagFilterInput = document.getElementById("tagFilterInput")
const logoutBtn = document.getElementById("logoutBtn")

/*********************************
 * GLOBAL STATE
 *********************************/
let hostsData = []
let currentAgentId = null
let logRefreshTimer = null
let detailsViewOpen = false
let lastLogTimestamp = null

const REFRESH_INTERVAL = 5000 // 5s
const API_TOKEN = sessionStorage.getItem("apiKey")

/*********************************
 * AUTH GUARD
 *********************************/
if (!sessionStorage.getItem("loggedIn") || !API_TOKEN) {
  window.location.href = "/index.html"
}

/*********************************
 * API KEY DISPLAY
 *********************************/
const keyInfo = document.getElementById("apiKeyInfo")
if (keyInfo) {
  keyInfo.textContent = "API Key: ****" + API_TOKEN.slice(-4)
}

/*********************************
 * UTILITIES
 *********************************/
function formatUptime(seconds) {
  if (!seconds || seconds <= 0) return "—"

  const units = [
    { label: "y", value: 31536000 },
    { label: "mo", value: 2592000 },
    { label: "d", value: 86400 },
    { label: "h", value: 3600 },
    { label: "m", value: 60 },
    { label: "s", value: 1 }
  ]

  let remaining = Math.floor(seconds)
  const parts = []

  for (const u of units) {
    const amount = Math.floor(remaining / u.value)
    if (amount > 0) {
      parts.push(`${amount}${u.label}`)
      remaining %= u.value
    }
    if (parts.length === 3) break
  }

  return parts.join(" ")
}

function timeAgo(date) {
  const seconds = Math.floor((Date.now() - new Date(date)) / 1000)
  if (seconds < 10) return "just now"
  if (seconds < 60) return `${seconds}s ago`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  return `${Math.floor(seconds / 3600)}h ago`
}

function renderTags(tags) {
  if (!tags || Object.keys(tags).length === 0) {
    return "<span style='color:#999'>—</span>"
  }

  return Object.entries(tags)
    .map(([k, v]) =>
      `<span class="tag"><span class="tag-key">${k}</span>: ${v}</span>`
    )
    .join("")
}

function getTimeRange() {
  const range = document.getElementById("logTimeRange")?.value || "1h"; // ✅ Fallback
  const now = new Date()
  const from = new Date(now)

  if (range === "5m") from.setMinutes(now.getMinutes() - 5)
  if (range === "15m") from.setMinutes(now.getMinutes() - 15)
  if (range === "1h") from.setHours(now.getHours() - 1)
  if (range === "24h") from.setHours(now.getHours() - 24)

  return {
    from: from.toISOString(),
    to: now.toISOString()
  }
}

/*********************************
 * HOST LIST
 *********************************/
function renderHostsList(data) {
  tbody.innerHTML = ""

  data.forEach(h => {
    const tr = document.createElement("tr")
    tr.innerHTML = `
      <td class="hostname">${h.hostname}</td>
      <td class="${h.alive ? "alive" : "dead"}">${h.alive ? "Alive" : "Down"}</td>
      <td>${timeAgo(h.last_seen)}</td>
    `
    tr.querySelector(".hostname")
      .addEventListener("click", () => showDetails(h))

    tbody.appendChild(tr)
  })
}

/*********************************
 * HOST DETAILS
 *********************************/
function showDetails(host) {
  if (!host || !host.agent_id) return

  detailsViewOpen = true
  currentAgentId = host.agent_id

  document.getElementById("hosts").style.display = "none"
  document.getElementById("hostDetails").style.display = "block"

  updateHostDetails(host)
  startLogAutoRefresh()
}

function updateHostDetails(host) {
  document.getElementById("detailsHostname").textContent = host.hostname
  document.getElementById("detailsIP").textContent = host.ip || "—"
  document.getElementById("detailsOS").textContent = host.os || "—"

  document.getElementById("detailsCPU").textContent =
    typeof host.cpu_percent === "number"
      ? host.cpu_percent.toFixed(1)
      : "—"

  document.getElementById("detailsMem").textContent =
    host.mem_used_mb ? `${host.mem_used_mb}` : "—"

  const uptimeSeconds = host.uptime_seconds ?? host.uptime_sec ?? 0
  const uptimeEl = document.getElementById("detailsUptime")
  uptimeEl.textContent = formatUptime(uptimeSeconds)
  uptimeEl.title = `${uptimeSeconds} seconds`

  document.getElementById("detailsTags").innerHTML =
    renderTags(host.tags || {})
}

/*********************************
 * LOGS
 *********************************/
function startLogAutoRefresh() {
  stopLogAutoRefresh()
  lastLogTimestamp = null
  loadLogs(currentAgentId, false)

  logRefreshTimer = setInterval(() => {
    if (currentAgentId) loadLogs(currentAgentId, true)
  }, REFRESH_INTERVAL)
}

function stopLogAutoRefresh() {
  if (logRefreshTimer) {
    clearInterval(logRefreshTimer)
    logRefreshTimer = null
  }
}

function loadLogs(agentId, incremental = false) {
  if (!agentId) return;

  let from, to;
  if (incremental && lastLogTimestamp) {
    from = lastLogTimestamp;
    to = new Date().toISOString();
  } else {
    const range = getTimeRange();
    from = range.from;
    to = range.to;
    lastLogTimestamp = null;
    document.getElementById("hostLogs").innerHTML = "";
  }

  const level = document.getElementById("logLevel")?.value || "";

  fetch(`/api/logs/fetch?agent_id=${agentId}&from=${from}&to=${to}&level=${level}`, {
    headers: { Authorization: "Bearer " + API_TOKEN }
  })
    .then(res => {
      if (!res.ok) throw new Error("Failed to fetch logs");
      return res.json();
    })
    .then(data => {
      const logsDiv = document.getElementById("hostLogs");

      if (!data.logs || data.logs.length === 0) {
        if (!incremental) logsDiv.textContent = "No logs available";
        return;
      }

      data.logs.forEach(l => {
        const div = document.createElement("div");
        const agentId = l.agent_id || l.agentId || "-";

        // Convert timestamp (seconds) to human-readable format
        const ts = l.timestamp ? new Date(Number(l.timestamp) * 1000) : null;
        const humanTime = ts && !isNaN(ts.getTime()) ? ts.getFullYear() + '-' +
                              String(ts.getMonth() + 1).padStart(2, '0') + '-' +
                              String(ts.getDate()).padStart(2, '0') + ' ' +
                              String(ts.getHours()).padStart(2, '0') + ':' +
                              String(ts.getMinutes()).padStart(2, '0') + ':' +
                              String(ts.getSeconds()).padStart(2, '0') : "—";

        // Default level and message if missing
        const logLevel = l.level || "info";
        const message = l.message || "—";

        div.textContent = `${humanTime} [${logLevel}] ${message}`;
        logsDiv.appendChild(div);
      });

      // Update last log timestamp (for incremental refresh)
      lastLogTimestamp = data.logs[data.logs.length - 1].timestamp;
      logsDiv.scrollTop = logsDiv.scrollHeight;
    })
    .catch(err => console.error("Log fetch error:", err));
}

/*********************************
 * BACK BUTTON
 *********************************/
document.getElementById("backBtn")
  .addEventListener("click", () => {
    detailsViewOpen = false
    currentAgentId = null
    stopLogAutoRefresh()

    document.getElementById("hostDetails").style.display = "none"
    document.getElementById("hosts").style.display = "table"
  })

/*********************************
 * FILTERING
 *********************************/
function applyFilter() {
  const search = searchInput.value.toLowerCase().trim()
  const tagFilter = tagFilterInput.value.trim()

  const filtered = hostsData.filter(h => {
    const matchesSearch = h.hostname.toLowerCase().includes(search)

    let matchesTag = true
    if (tagFilter) {
      const [key, value] = tagFilter.split("=")
      matchesTag = key && value && h.tags && h.tags[key] === value
    }

    return matchesSearch && matchesTag
  })

  renderHostsList(filtered)
}

searchInput.addEventListener("input", applyFilter)
tagFilterInput.addEventListener("input", applyFilter)

/*********************************
 * LOGOUT
 *********************************/
logoutBtn.addEventListener("click", () => {
  sessionStorage.clear()
  window.location.href = "/index.html"
})

/*********************************
 * HOST FETCH (AUTO-REFRESH DETAILS)
 *********************************/
function loadHosts() {
  fetch("/api/hosts", {
    headers: { Authorization: "Bearer " + API_TOKEN }
  })
    .then(res => {
      if (!res.ok) throw new Error("Unauthorized")
      return res.json()
    })
    .then(data => {
      hostsData = data
      applyFilter()
      
      // 🔥 AUTO-REFRESH DETAILS VIEW
      if (detailsViewOpen && currentAgentId) {
        const updatedHost = hostsData.find(
          h => h.agent_id === currentAgentId
        )
        if (updatedHost) {
          updateHostDetails(updatedHost)
        }
      }
    })
    .catch(err =>
      console.error("Failed to load hosts:", err)
    )
}

document.getElementById("logLevel")
  ?.addEventListener("change", () => {
    lastLogTimestamp = null
    loadLogs(currentAgentId, false)
  })
/*********************************
 * INITIAL LOAD
 *********************************/
loadHosts()
setInterval(loadHosts, REFRESH_INTERVAL)
