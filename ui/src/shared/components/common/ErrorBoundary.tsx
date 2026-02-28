import React from "react"

interface State {
  hasError: boolean
  error?: Error
}

export default class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  State
> {
  constructor(props: any) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: any) {
    console.error("🔥 UI CRASH CAUGHT:", error, errorInfo)
  }

  handleReload = () => {
    window.location.reload()
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{ padding: "40px", textAlign: "center" }}>
          <h2>⚠️ Something went wrong</h2>
          <p>The dashboard encountered an unexpected error.</p>
          <button onClick={this.handleReload}>Reload</button>
        </div>
      )
    }

    return this.props.children
  }
}
