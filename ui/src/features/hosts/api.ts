import { apiGet, apiPost } from "../../shared/lib/api/client"

export async function fetchHosts() {
  return apiGet("/api/hosts")
}

export async function fetchHostSummary(agentId: string) {
  return apiGet(`/api/hosts/summary?agent_id=${agentId}`)
}

export async function fetchLogs(agentId: string) {
  return apiGet(`/api/logs/fetch?agent_id=${agentId}`)
}

export async function fetchContainerMetrics(agentId: string) {
  return apiGet(`/api/containers/metrics?agent_id=${agentId}`)
}
