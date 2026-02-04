const tbody = document.querySelector("#hosts tbody")
const searchInput = document.getElementById("searchInput")
const tagFilter = document.getElementById("tagFilter")
const logoutBtn = document.getElementById("logoutBtn")

// redirect if not logged in
if (!sessionStorage.getItem("loggedIn")) window.location.href = "/login.html"

let hostsData = []
const API_TOKEN = "mysecrettoken"; // use the token that works

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

function renderHosts(data) {
  tbody.innerHTML = ""
  data.forEach(h => {
    const tr = document.createElement("tr")
    tr.innerHTML = `
      <td>${h.hostname}</td>
      <td>${h.cpu_percent.toFixed(1)}%</td>
      <td>${h.mem_used_mb} MB</td>
      <td>${renderTags(h.tags)}</td>
      <td class="${h.alive ? "alive" : "dead"}">${h.alive ? "Alive" : "Down"}</td>
      <td>${timeAgo(h.last_seen)}</td>
    `
    tbody.appendChild(tr)
  })
}

function populateTagFilter() {
  const allTags = new Set()
  hostsData.forEach(h => {
    if (h.tags) Object.keys(h.tags).forEach(t => allTags.add(t))
  })
  tagFilter.innerHTML = '<option value="">All Tags</option>'
  allTags.forEach(tag => {
    const opt = document.createElement("option")
    opt.value = tag
    opt.textContent = tag
    tagFilter.appendChild(opt)
  })
}

function applyFilter() {
  const search = searchInput.value.toLowerCase()
  const selectedTag = tagFilter.value

  const filtered = hostsData.filter(h => {
    const matchesSearch = h.hostname.toLowerCase().includes(search)
    const matchesTag = !selectedTag || (h.tags && selectedTag in h.tags)
    return matchesSearch && matchesTag
  })
  renderHosts(filtered)
}

searchInput.addEventListener("input", applyFilter)
tagFilter.addEventListener("change", applyFilter)

logoutBtn.addEventListener("click", () => {
  sessionStorage.removeItem("loggedIn")
  window.location.href = "/login.html"
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
      populateTagFilter() // only once per fetch
      applyFilter()
    })
    .catch(err => console.error("Failed to load hosts:", err))
}

loadHosts()
setInterval(loadHosts, 5000)
