import ReactDOM from "react-dom/client"
import { BrowserRouter } from "react-router-dom"
import App from "./app/App"
import "./styles/tokens.css"
import "./styles/themes/light.css"
import "./styles/themes/dark.css"
import "./styles/base.css"
import "./styles/layout.css"
import "./styles/components.css"
import { AuthProvider } from "./features/auth/AuthContext"
import { SearchProvider } from "./context/SearchContext"
import ErrorBoundary from "./shared/components/common/ErrorBoundary"

// 🔥 Initialize theme BEFORE React loads
const savedTheme = localStorage.getItem("theme") || "dark"

ReactDOM.createRoot(document.getElementById("root")!).render(
  <BrowserRouter>
    <AuthProvider>
      <SearchProvider>
        <ErrorBoundary>
          <App />
        </ErrorBoundary>
      </SearchProvider>
    </AuthProvider>
  </BrowserRouter>
)