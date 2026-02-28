import { apiPost } from "../../shared/lib/api/client"

export async function loginRequest(
  username: string,
  password: string
) {
  return apiPost("/api/login", {
    username,
    password,
  })
}