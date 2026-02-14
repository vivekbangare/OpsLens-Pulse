// -------------------------------
// API CLIENT (Production Safe)
// -------------------------------

function getHeaders(): HeadersInit {
  const token = sessionStorage.getItem("apiKey")

  if (!token) return {}

  return {
    Authorization: "Bearer " + token,
  }
}

async function safeFetch(url: string) {
  try {
    const res = await fetch(url, {
      headers: getHeaders(),
    })

    if (!res.ok) {
      console.warn(`API error: ${res.status} → ${url}`)
      return null
    }

    const data = await res.json()
    return data ?? null
  } catch (err) {
    console.error("Network/API failure:", err)
    return null
  }
}

// -------------------------------
// Hosts
// -------------------------------

export async function fetchHosts(): Promise<any[]> {
  const data = await safeFetch("/api/hosts")
  return Array.isArray(data) ? data : []
}

// -------------------------------
// Logs
// -------------------------------

export async function fetchLogs(agentId: string): Promise<any[]> {
  const data = await safeFetch(
    `/api/logs/fetch?account_id=default&agent_id=${agentId}`
  )

  if (Array.isArray(data)) return data
  if (Array.isArray(data?.logs)) return data.logs

  return []
}

// -------------------------------
// Host Summary
// -------------------------------

export async function fetchHostSummary(agentId: string) {
  const data = await safeFetch("/api/hosts/summary")

  if (!data || typeof data !== "object") return null

  return data[agentId] ?? null
}

// -------------------------------
// Containers
// -------------------------------

export async function fetchContainers(agentId: string): Promise<any[]> {
  const data = await safeFetch(`/api/containers?agent_id=${agentId}`)
  return Array.isArray(data) ? data : []
}
