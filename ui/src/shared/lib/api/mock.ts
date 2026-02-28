function base64UrlEncode(obj: any) {
  return btoa(JSON.stringify(obj))
    .replace(/=/g, "")
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
}

export async function mockGet(url: string) {
  if (url.includes("/api/hosts")) {
    return [
      {
        agent_id: "1",
        hostname: "opslens-agent",
        alive: true,
        ip: "172.18.0.7",
        public_ip: "1.1.1.1",
        os: "linux",
        first_seen: new Date().toISOString(),
        last_seen: new Date().toISOString(),
      },
    ]
  }

  if (url.includes("/api/logs")) {
    return [
      {
        ts: Math.floor(Date.now() / 1000),
        level: "ERROR",
        message: "Mock log error",
        source_name: "container-1",
      },
    ]
  }
  return []
}



export async function mockPost(url: string, body: any) {
  if (url.includes("/api/login")) {
    const header = { alg: "HS256", typ: "JWT" }

    const payload = {
      sub: "1",
      username: "demo",
      email: "demo@opslens.io",
      is_super_admin: true,
      tenants: [
        { id: "t1", name: "Demo Tenant", slug: "demo" },
      ],
      currentTenantId: "t1",
    }

    const token =
      base64UrlEncode(header) +
      "." +
      base64UrlEncode(payload) +
      ".mocksignature"

    return { token }
  }

  return []
}