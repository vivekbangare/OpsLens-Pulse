export type AgentHealth = {
  label: string
  className: string
  reason: string
}

export function getAgentHealth(h: any): AgentHealth {
  if (!h.alive) {
    return {
      label: "Down",
      className: "critical",
      reason: "Agent not reachable",
    }
  }

  if (
    h.cpu_percent !== undefined &&
    h.mem_used_mb !== undefined &&
    h.mem_total_mb !== undefined
  ) {
    const memPercent =
      (h.mem_used_mb / h.mem_total_mb) * 100

    if (h.cpu_percent > 85 || memPercent > 90) {
      return {
        label: "Critical",
        className: "critical",
        reason: "High CPU or Memory",
      }
    }

    if (h.cpu_percent > 70 || memPercent > 80) {
      return {
        label: "Degraded",
        className: "warning",
        reason: "Elevated resource usage",
      }
    }

    return {
      label: "Healthy",
      className: "healthy",
      reason: "Operating normally",
    }
  }

  return {
    label: "Unknown",
    className: "unknown",
    reason: "No metrics available",
  }
}