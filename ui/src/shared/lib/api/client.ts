import { User } from "../../../features/auth/AuthContext"
import { mockGet, mockPost } from "./mock"
import { ENV } from "../../config/env"

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

export function getHeaders(): HeadersInit {
  const token = sessionStorage.getItem("token")

  const user = safeParse<User>(
    sessionStorage.getItem("user")
  )

  if (!token) return {}

  return {
    Authorization: "Bearer " + token,
    "X-Tenant-ID": user?.currentTenantId || "",
    "Content-Type": "application/json",
  }
}

export async function apiGet(url: string) {
  if (ENV.USE_MOCKS) {
    return mockGet(url)
  }

  const res = await fetch(url, {
    headers: getHeaders(),
  })

  if (!res.ok) throw new Error(`API error ${res.status}`)
  return res.json()
}

export async function apiPost(url: string, body: any) {
  if (ENV.USE_MOCKS) {
    return mockPost(url, body)
  }

  const res = await fetch(url, {
    method: "POST",
    headers: getHeaders(),
    body: JSON.stringify(body),
  })

  if (!res.ok) throw new Error(`API error ${res.status}`)
  return res.json()
}