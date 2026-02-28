import {
  useEffect,
  useState,
  useRef,
  KeyboardEvent,
} from "react"
import { useNavigate } from "react-router-dom"
import { Search } from "lucide-react"
import { apiPost } from "../lib/api/client"

interface ResultItem {
  type: "host" | "log" | "alert"
  label: string
  sub?: string
  route: string
}

export default function GlobalSearch() {
  const [query, setQuery] = useState("")
  const [results, setResults] = useState<ResultItem[]>([])
  const [open, setOpen] = useState(false)
  const [selectedIndex, setSelectedIndex] = useState(0)

  const navigate = useNavigate()
  const inputRef = useRef<HTMLInputElement>(null)
  const wrapperRef = useRef<HTMLDivElement>(null)

  // ----------------------------------
  // ⌘K / Ctrl+K Shortcut
  // ----------------------------------
  useEffect(() => {
    function handleShortcut(e: KeyboardEvent | any) {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault()
        inputRef.current?.focus()
        setOpen(true)
      }
    }

    window.addEventListener("keydown", handleShortcut)
    return () => window.removeEventListener("keydown", handleShortcut)
  }, [])

  // ----------------------------------
  // Debounced Search
  // ----------------------------------
  useEffect(() => {
    if (!query) {
      setResults([])
      return
    }

    const timer = setTimeout(async () => {
      try {
        const data = await apiPost("/api/search/global", { query })

        const combined: ResultItem[] = []

        data?.hosts?.forEach((h: any) => {
          combined.push({
            type: "host",
            label: h.hostname,
            sub: h.ip,
            route: `/hosts/${h.agent_id}`,
          })
        })

        data?.logs?.length &&
          combined.push({
            type: "log",
            label: `Logs containing "${query}"`,
            route: `/logs?query=${query}`,
          })

        data?.alerts?.forEach((a: any) => {
          combined.push({
            type: "alert",
            label: a.title,
            route: `/alerts`,
          })
        })

        setResults(combined)
        setOpen(true)
        setSelectedIndex(0)

      } catch (err) {
        console.error("Search error:", err)
      }
    }, 250)

    return () => clearTimeout(timer)
  }, [query])

  // ----------------------------------
  // Close on outside click
  // ----------------------------------
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (
        wrapperRef.current &&
        !wrapperRef.current.contains(e.target as Node)
      ) {
        setOpen(false)
      }
    }
    document.addEventListener("mousedown", handleClickOutside)
    return () =>
      document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  // ----------------------------------
  // Keyboard Navigation
  // ----------------------------------
  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (!open) return

    if (e.key === "ArrowDown") {
      e.preventDefault()
      setSelectedIndex((prev) =>
        prev < results.length - 1 ? prev + 1 : prev
      )
    }

    if (e.key === "ArrowUp") {
      e.preventDefault()
      setSelectedIndex((prev) =>
        prev > 0 ? prev - 1 : prev
      )
    }

    if (e.key === "Enter") {
      e.preventDefault()
      const selected = results[selectedIndex]
      if (selected) {
        navigate(selected.route)
        setOpen(false)
        setQuery("")
      }
    }

    if (e.key === "Escape") {
      setOpen(false)
    }
  }

  // ----------------------------------
  // Highlight matching text
  // ----------------------------------
  function highlight(text: string) {
    if (!query) return text
    const regex = new RegExp(`(${query})`, "gi")
    return text.split(regex).map((part, i) =>
      regex.test(part) ? (
        <span key={i} className="highlight">
          {part}
        </span>
      ) : (
        part
      )
    )
  }

  return (
    <div className="globalSearch" ref={wrapperRef}>
      <Search size={16} className="searchIcon" />
      <input
        ref={inputRef}
        placeholder="Search anything... (⌘K)"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        onFocus={() => query && setOpen(true)}
        onKeyDown={handleKeyDown}
      />

      {open && results.length > 0 && (
        <div className="searchDropdown">
          {results.map((item, index) => (
            <div
              key={index}
              className={`searchItem ${
                selectedIndex === index ? "active" : ""
              }`}
              onClick={() => {
                navigate(item.route)
                setOpen(false)
                setQuery("")
              }}
            >
              <div>
                {highlight(item.label)}
              </div>
              {item.sub && (
                <div className="subText">
                  {item.sub}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}