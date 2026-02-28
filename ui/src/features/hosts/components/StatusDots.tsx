import React from "react"

type Props = {
  hostAlive: boolean
  cpu?: number
  memUsed?: number
  memTotal?: number
  lastSeen?: string
}

function getHostHealth(cpu?: number, memUsed?: number, memTotal?: number) {
  if (!cpu || !memUsed || !memTotal) return "unknown"

  const memPercent = (memUsed / memTotal) * 100

  if (cpu > 95 || memPercent > 95) return "critical"
  if (cpu > 80 || memPercent > 85) return "warning"
  return "healthy"
}

function getAgentHealth(alive: boolean) {
  return alive ? "healthy" : "offline"
}

export default function StatusDots({
  hostAlive,
  cpu,
  memUsed,
  memTotal,
}: Props) {
  const hostHealth = getHostHealth(cpu, memUsed, memTotal)
  const agentHealth = getAgentHealth(hostAlive)

  return (
    <div className="statusDots">
      <span
        className={`dot ${hostHealth}`}
        title={`Host: ${hostHealth}`}
      />
      <span
        className={`dot ${agentHealth}`}
        title={`Agent: ${agentHealth}`}
      />
    </div>
  )
}