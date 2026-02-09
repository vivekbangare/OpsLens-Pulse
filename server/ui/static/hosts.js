const tbody = document.querySelector("#hosts tbody")
const containersTbody = document.querySelector("#containers tbody")

let hostsData = []
let currentAgentId = null
let currentContainerId = null

const API_TOKEN = sessionStorage.getItem("apiKey")
const REFRESH_INTERVAL = 5000

if (!API_TOKEN) window.location.href = "/index.html"

document.getElementById("apiKeyInfo").textContent =
  "API Key: ****" + API_TOKEN.slice(-4)

/* ---------------- HOST LIST ---------------- */

function renderHosts(data) {
  tbody.innerHTML = ""
  data.forEach(h => {
    const tr = document.createElement("tr")
    tr.innerHTML = `
      <td class="link">${h.hostname}</td>
      <td class="${h.alive ? "alive" : "dead"}">${h.alive ? "Alive" : "Down"}</td>
      <td>${new Date(h.last_seen).toLocaleString()}</td>
    `
    tr.querySelector(".link").onclick = () => openHost(h)
    tbody.appendChild(tr)
  })
}

function loadHosts() {
  fetch("/api/hosts", {
    headers: { Authorization: "Bearer " + API_TOKEN }
  })
    .then(r => r.json())
    .then(d => {
      hostsData = d
      renderHosts(d)
    })
}

setInterval(loadHosts, REFRESH_INTERVAL)
loadHosts()

/* ---------------- HOST DETAILS ---------------- */

function openHost(host) {
  currentAgentId = host.agent_id

  document.getElementById("hosts").style.display = "none"
  document.getElementById("hostDetails").style.display = "block"

  document.getElementById("detailsHostname").textContent = host.hostname
  document.getElementById("detailsIP").textContent = host.ip || "—"
  document.getElementById("detailsOS").textContent = host.os || "—"
  document.getElementById("detailsCPU").textContent =
    host.cpu_percent?.toFixed(1) ?? "—"
  document.getElementById("detailsMem").textContent =
    host.mem_used_mb ?? "—"
  document.getElementById("detailsUptime").textContent =
    host.uptime_sec ?? "—"
  document.getElementById("detailsTags").innerHTML = renderTags(host.tags)

  loadHostLogs()
  loadContainers()
}

document.getElementById("backBtn").onclick = () => {
  document.getElementById("hostDetails").style.display = "none"
  document.getElementById("hosts").style.display = "table"
  document.getElementById("containerLogsPanel").style.display = "none"
}

/* ---------------- TAG RENDER ---------------- */

function renderTags(tags = {}) {
  return Object.entries(tags)
    .map(([k, v]) => `<span class="tag">${k}:${v}</span>`)
    .join(" ") || "—"
}

/* ---------------- HOST LOGS ---------------- */

function loadHostLogs() {
  fetch(`/api/logs/fetch?agent_id=${currentAgentId}`, {
    headers: { Authorization: "Bearer " + API_TOKEN }
  })
    .then(r => r.json())
    .then(d => {
      const box = document.getElementById("hostLogs")
      box.innerHTML = ""
      d.logs?.forEach(l => {
        box.innerHTML += `${new Date(l.timestamp * 1000).toISOString()} [${l.level}] ${l.message}<br/>`
      })
    })
}

/* ---------------- CONTAINERS ---------------- */

function loadContainers() {
  fetch(`/api/containers?agent_id=${currentAgentId}`, {
    headers: { Authorization: "Bearer " + API_TOKEN }
  })
    .then(r => r.json())
    .then(containers => {
      containersTbody.innerHTML = ""
      containers.forEach(c => {
        const tr = document.createElement("tr")
        tr.innerHTML = `
          <td class="link">${c.name}</td>
          <td>${c.image}</td>
          <td>${c.status}</td>
          <td>${c.cpu_percent?.toFixed(1) ?? "—"}%</td>
          <td>${c.mem_used_mb ?? "—"} MB</td>
        `
        tr.querySelector(".link").onclick = () => openContainer(c)
        containersTbody.appendChild(tr)
      })
    })
}

function openContainer(c) {
  currentContainerId = c.container_id
  document.getElementById("containerLogsPanel").style.display = "block"
  document.getElementById("containerTitle").textContent = c.name

  fetch(`/api/container/logs/fetch?container_id=${currentContainerId}`, {
    headers: { Authorization: "Bearer " + API_TOKEN }
  })
    .then(r => r.json())
    .then(d => {
      const box = document.getElementById("containerLogs")
      box.innerHTML = ""
      d.logs?.forEach(l => {
        box.innerHTML += `${new Date(l.timestamp * 1000).toISOString()} ${l.message}<br/>`
      })
    })
}

/* ---------------- LOGOUT ---------------- */

document.getElementById("logoutBtn").onclick = () => {
  sessionStorage.clear()
  window.location.href = "/index.html"
}
