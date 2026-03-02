export function formatUptime(seconds?: number | null): string {
  if (!seconds || seconds <= 0) return "—"

  const units = [
    { label: "y", value: 60 * 60 * 24 * 365 },
    { label: "mo", value: 60 * 60 * 24 * 30 },
    { label: "d", value: 60 * 60 * 24 },
    { label: "h", value: 60 * 60 },
    { label: "m", value: 60 },
    { label: "s", value: 1 },
  ]

  let remaining = seconds
  const parts: string[] = []

  for (const unit of units) {
    const amount = Math.floor(remaining / unit.value)
    if (amount > 0) {
      parts.push(`${amount}${unit.label}`)
      remaining -= amount * unit.value
    }
    if (parts.length === 2) break // keep UI clean (max 2 units)
  }

  return parts.join(" ")
}