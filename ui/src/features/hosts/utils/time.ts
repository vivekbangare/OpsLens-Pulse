export function formatLastSeen(dateString?: string) {
  if (!dateString) return { relative: "—", absolute: "" }

  const date = new Date(dateString)
  const now = new Date()

  const diffMs = now.getTime() - date.getTime()
  const diffSeconds = Math.floor(diffMs / 1000)
  const diffMinutes = Math.floor(diffSeconds / 60)
  const diffHours = Math.floor(diffMinutes / 60)
  const diffDays = Math.floor(diffHours / 24)
  const diffYears = Math.floor(diffDays / 365)

  let relative = ""

  if (diffSeconds < 60) {
    relative = `${diffSeconds}s ago`
  } else if (diffMinutes < 60) {
    relative = `${diffMinutes} min ago`
  } else if (diffHours < 24) {
    relative = `${diffHours} hr ago`
  } else if (diffDays < 365) {
    relative = `${diffDays} day${diffDays > 1 ? "s" : ""} ago`
  } else {
    relative = `${diffYears} year${diffYears > 1 ? "s" : ""} ago`
  }

  const absolute = date.toLocaleString()

  return { relative, absolute }
}