const tbody = document.querySelector("#hosts tbody")
const table = document.querySelector("#hosts")
const loginBox = document.querySelector("#login")
const errorBox = document.querySelector("#error")

let token = sessionStorage.getItem("token")

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
      <td class="${h.alive ? "alive" : "dead"}">
        ${h.alive ? "Alive" : "Down"}
      </td>
      <td>${timeAgo(h.last_seen)}</td>
    `
    tbody.appendChild(tr)
  })
}

function authFetch(url) {
  return fetch(url, {
    headers: {
      "Authorization": "Bearer " + token
    }
  })
}

function loadHosts() {
  authFetch("/api/hosts")
    .then(res => {
      if (res.status === 401) throw new Error("unauthorized")
      return res.json()
    })
    .then(data => {
      errorBox.textContent = ""
      renderHosts(data)
    })
    .catch(() => {
      errorBox.textContent = "Invalid token"
      logout()
    })
}

function login() {
  const input = document.getElementById("tokenInput").value.trim()
  if (!input) return

  token = input
  sessionStorage.setItem("token", token)

  // validate token first
  authFetch("/api/hosts")
    .then(res => {
      if (!res.ok) throw new Error()
      loginBox.style.display = "none"
      table.style.display = "table"
      loadHosts()
    })
    .catch(() => {
      errorBox.textContent = "Invalid token"
      sessionStorage.removeItem("token")
      token = null
    })
}

function logout() {
  sessionStorage.removeItem("token")
  token = null
  loginBox.style.display = "block"
  table.style.display = "none"
}

// auto-login
if (token) {
  loginBox.style.display = "none"
  table.style.display = "table"
  loadHosts()
}

// refresh every 5 seconds
setInterval(() => {
  if (token) loadHosts()
}, 5000)
