export function getHeaders(): HeadersInit {
  const token = localStorage.getItem("token")

  const headers: HeadersInit = {
    "Content-Type": "application/json",
  }

  if (token) {
    headers["Authorization"] = "Bearer " + token
  }

  return headers
}

// Centralized unauthorized handler
function handleUnauthorized() {
  localStorage.removeItem("token")
  localStorage.removeItem("user")

  // Let React routing handle redirect via ProtectedRoute
  // Trigger a soft reload to re-evaluate auth state
  window.history.replaceState(null, "", "/login")
  window.dispatchEvent(new Event("auth:logout"))
}

export async function apiGet(url: string) {
  const res = await fetch(url, {
    headers: getHeaders(),
  })

  if (res.status === 401) {
    handleUnauthorized()
    throw new Error("Unauthorized")
  }

  if (!res.ok) {
    throw new Error(`API error ${res.status}`)
  }

  return res.json()
}

export async function apiPost(url: string, body: any) {
  const res = await fetch(url, {
    method: "POST",
    headers: getHeaders(),
    body: JSON.stringify(body),
  })

  if (res.status === 401) {
    handleUnauthorized()
    throw new Error("Unauthorized")
  }

  if (!res.ok) {
    throw new Error(`API error ${res.status}`)
  }

  return res.json()
}