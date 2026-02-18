import ReactDOM from "react-dom/client"
import { BrowserRouter } from "react-router-dom"
import App from "./App"
import "./styles/dashboard.css"
import ErrorBoundary from "./components/ErrorBoundary"
import { AuthProvider } from "./auth/AuthContext"
import { SearchProvider } from "./context/SearchContext"

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
