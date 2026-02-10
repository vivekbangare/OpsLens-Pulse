const API_TOKEN = sessionStorage.getItem("apiKey")
if (!API_TOKEN) location.href = "/index.html"

document.getElementById("apiKeyInfo").textContent =
  "API Key: ****" + API_TOKEN.slice(-4)

const hostsTbody = document.querySelector("#hostsTable tbody")
const containersTbody = document.querySelector("#containersTable tbody")

let hosts = []
let summaryCache = {}
let currentAgentId = null
let detailsTimer = null
let containersLoadedFor = null

const REFRESH = 5000

/* ---------------- UTIL ---------------- */

function timeAgo(ts) {
  if (!ts) return "—"
  const diff = Math.floor((Date.now() - new Date(ts)) / 1000)
  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  if (diff < 31536000) return `${Math.floor(diff / 86400)}d ago`
  return `${Math.floor(diff / 31536000)}y ago`
}

function renderTags(tags) {
  if (!tags || typeof tags !== "object") return "—"
  return Object.entries(tags)
    .map(([k, v]) => `<span class="tag">${k}:${v}</span>`)
    .join(" ")
}

/* ---------------- HOST LIST ---------------- */

function renderHosts() {
  hostsTbody.innerHTML = ""
  hosts.forEach(h => {
    const tr = document.createElement("tr")
    tr.innerHTML = `
      <td class="link">
        <span class="dot ${h.alive ? "alive" : "dead"}"></span>
        ${h.hostname}
      </td>
      <td class="${h.alive ? "alive" : "dead"}">${h.alive ? "Alive" : "Down"}</td>
      <td>${timeAgo(h.last_seen)}</td>
      <td>${h.ip || "—"}</td>
      <td>${h.os || "—"}</td>
      <td>${renderTags(h.tags)}</td>
    `
    tr.querySelector(".link").onclick = () => openHost(h.agent_id)
    hostsTbody.appendChild(tr)
  })
}

async function loadHosts() {
  const res = await fetch("/api/hosts", {
    headers: { Authorization: "Bearer " + API_TOKEN }
  })
  hosts = await res.json()

  const summaryRes = await fetch("/api/hosts/summary", {
    headers: { Authorization: "Bearer " + API_TOKEN }
  })
  summaryCache = await summaryRes.json()

  hosts = hosts.map(h => {
    const s = summaryCache[h.agent_id]
    if (!s) return h
    return {
      ...h,
      cpu_percent: s.cpu_percent,
      mem_used_mb: s.mem_used_mb,
      mem_total_mb: s.mem_total_mb,
      uptime_sec: s.uptime_sec
    }
  })

  renderHosts()
}

setInterval(loadHosts, REFRESH)
loadHosts()

/* ---------------- HOST DETAILS ---------------- */

function openHost(agentId) {
  currentAgentId = agentId
  containersLoadedFor = null

  document.getElementById("hostsTable").style.display = "none"
  document.getElementById("hostDetails").style.display = "block"

  refreshHostDetails()
  loadContainersOnce()

  detailsTimer = setInterval(refreshHostDetails, REFRESH)
}

function refreshHostDetails() {
  if (!currentAgentId) return

  const host = hosts.find(h => h.agent_id === currentAgentId)
  if (!host) return

  document.getElementById("detailsHostname").textContent = host.hostname
  document.getElementById("detailsStatus").textContent =
    host.alive ? "🟢 Alive" : "🔴 Down"

  document.getElementById("detailsCPU").textContent =
    typeof host.cpu_percent === "number"
      ? host.cpu_percent.toFixed(1)
      : "—"

  document.getElementById("detailsMem").textContent =
    host.mem_used_mb ?? "—"

  document.getElementById("detailsUptime").textContent =
    host.uptime_sec
      ? `${Math.floor(host.uptime_sec / 60)} min`
      : "—"

  document.getElementById("detailsTags").innerHTML =
    renderTags(host.tags)

  loadLogs()
}

document.getElementById("backBtn").onclick = () => {
  if (detailsTimer) {
    clearInterval(detailsTimer)
    detailsTimer = null
  }

  currentAgentId = null
  containersLoadedFor = null

  document.getElementById("hostDetails").style.display = "none"
  document.getElementById("hostsTable").style.display = "table"
}

/* ---------------- LOGS ---------------- */

async function loadLogs() {
  if (!currentAgentId) return

  const box = document.getElementById("hostLogs")

  let res
  try {
    res = await fetch(
      `/api/logs/fetch?account_id=default&agent_id=${currentAgentId}`,
      {
        headers: { Authorization: "Bearer " + API_TOKEN }
      }
    )
  } catch (e) {
    console.warn("Logs fetch failed:", e)
    box.innerHTML = "<span class='muted'>Unable to reach logs service</span>"
    return
  }

  // Handle HTTP errors
  if (!res.ok) {
    box.innerHTML =
      `<span class='muted'>Logs unavailable (${res.status})</span>`
    return
  }

  let data
  try {
    data = await res.json()
  } catch (e) {
    console.warn("Invalid JSON from logs API:", e)
    box.innerHTML = "<span class='muted'>Invalid logs response</span>"
    return
  }

  // Support multiple response shapes safely
  const logs = Array.isArray(data?.logs)
    ? data.logs
    : Array.isArray(data)
    ? data
    : []

  if (!logs.length) {
    box.innerHTML = "<span class='muted'>No logs found</span>"
    return
  }

  box.innerHTML = logs
    .map(l => {
      const ts = l.timestamp
        ? new Date(l.timestamp * 1000).toLocaleString()
        : "—"
      const level = l.level || "info"
      const msg = l.message || ""
      return `${ts} [${level}] ${msg}`
    })
    .join("<br/>")
}



/* ---------------- CONTAINERS (ONCE ONLY) ---------------- */

async function loadContainersOnce() {
  if (!currentAgentId) return
  if (containersLoadedFor === currentAgentId) return

  containersLoadedFor = currentAgentId

  let res
  try {
    res = await fetch(`/api/containers?agent_id=${currentAgentId}`, {
      headers: { Authorization: "Bearer " + API_TOKEN }
    })
  } catch {
    containersTbody.innerHTML =
      `<tr><td colspan="5" class="muted">Containers API unreachable</td></tr>`
    return
  }

  if (!res.ok) {
    containersTbody.innerHTML =
      `<tr><td colspan="5" class="muted">Containers not supported</td></tr>`
    return
  }

  let containers = []
  try {
    containers = await res.json()
  } catch {
    containersTbody.innerHTML =
      `<tr><td colspan="5" class="muted">Invalid containers response</td></tr>`
    return
  }

  containersTbody.innerHTML = containers.length
    ? containers.map(c => `
        <tr>
          <td>${c.name}</td>
          <td>${c.image}</td>
          <td>${c.status}</td>
          <td>${c.cpu_percent?.toFixed(1) ?? "—"}%</td>
          <td>${c.mem_used_mb ?? "—"} MB</td>
        </tr>
      `).join("")
    : `<tr><td colspan="5" class="muted">No containers detected</td></tr>`
}

