export async function loginRequest(username: string, password: string) {
  const res = await fetch("/api/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  })

  if (!res.ok) {
    throw new Error("Invalid credentials")
  }

  const json = await res.json()

  // Expected backend format:
  // {
  //   data: { token: "..." },
  //   message: "...",
  //   request_id: "..."
  // }

  if (!json?.data?.token) {
    console.error("LOGIN RAW RESPONSE:", json)
    throw new Error("Token missing in response")
  }

  return {
    token: json.data.token,
  }
}