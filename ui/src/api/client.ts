// -------------------------------
// API CLIENT (Production Safe)
// -------------------------------

import { data } from "react-router-dom"
import { User } from "../auth/AuthContext"

function safeParse<T>(value: string | null): T | null {
  if (!value || value === "undefined" || value === "null") {
    return null
  }

  try {
    return JSON.parse(value)
  } catch {
    return null
  }
}

function getHeaders(): HeadersInit {
  const token = 
    sessionStorage.getItem("token") || localStorage.getItem("token")

    const user = safeParse<User>(
      sessionStorage.getItem("user") || localStorage.getItem("user")
    ) 


  if (!token) return {}

  return {
    Authorization: token ? "Bearer " + token : "",
    "X-Tenant-ID": user?.currentTenantId || "",
    "Content-Type": "application/json",
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
// Host Summary
// -------------------------------

export async function fetchHostSummary(agentId: string) {
  try {
    const res = await fetch("/api/hosts/summary", {
      method: "POST",
      headers: getHeaders(),
      body: JSON.stringify({
        agent_id: agentId,
      }),
    })

    if (!res.ok) {
      console.warn(`API error: ${res.status} → host summary`)
      return null
    }
    const data = await res.json()

    if (data && data[agentId]) {
      return data[agentId]
    }

    return null
  } catch (err) {
    console.error("Host summary error:", err)
    return null
  }
}

// -------------------------------
// Logs
// -------------------------------

export async function fetchLogs(agentId: string): Promise<any[]> {
  const data = await safeFetch(
    `/api/logs/fetch?agent_id=${agentId}`
  )

  if (Array.isArray(data)) return data
  if (Array.isArray(data?.logs)) return data.logs

  return []
}

// -------------------------------
// Container Metrics (FIXED)
// -------------------------------

export async function fetchContainerMetrics(agentId: string): Promise<any[]> {
  try {
    const res = await fetch("/api/container/metrics", {
      method: "POST",
      headers: {
        ...getHeaders(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        agent_id: agentId,
      }),
    })

    if (!res.ok) {
      console.warn(`API error: ${res.status} → container metrics`)
      return []
    }

    const data = await res.json()
    return Array.isArray(data) ? data : []
  } catch (err) {
    console.error("Container metrics error:", err)
    return []
  }
}


// -------------------------------
// Container Logs (FIXED)
// -------------------------------

export async function fetchContainerLogs(agentId: string): Promise<any[]> {
  const data = await safeFetch(
    `/api/container/logs?agent_id=${agentId}`
  )

  return Array.isArray(data) ? data : []
}
