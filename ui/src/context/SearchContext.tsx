import { createContext, useContext, useState } from "react"

interface SearchConfig {
  placeholder: string
  onSearch: (value: string) => void
}

interface SearchContextType {
  config: SearchConfig | null
  setSearchConfig: (config: SearchConfig | null) => void
}

const SearchContext = createContext<SearchContextType | undefined>(undefined)

export function SearchProvider({ children }: { children: React.ReactNode }) {
  const [config, setSearchConfig] = useState<SearchConfig | null>(null)

  return (
    <SearchContext.Provider value={{ config, setSearchConfig }}>
      {children}
    </SearchContext.Provider>
  )
}

export function useSearch() {
  const ctx = useContext(SearchContext)
  if (!ctx) throw new Error("useSearch must be used inside SearchProvider")
  return ctx
}
