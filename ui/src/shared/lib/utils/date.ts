export function formatDate(ts?: number | null): string {
  if (!ts) return "—"

  const date = new Date(ts * 1000)

  return date.toLocaleString(undefined, {
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  })
}