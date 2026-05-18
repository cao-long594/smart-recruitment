const TOKEN_KEY = 'recruitment_user_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(t: string | null) {
  if (t) localStorage.setItem(TOKEN_KEY, t)
  else localStorage.removeItem(TOKEN_KEY)
}

export async function api<T>(
  path: string,
  init: RequestInit = {}
): Promise<T> {
  const headers: Record<string, string> = {
    ...(init.headers as Record<string, string>),
  }
  if (init.body && !(init.body instanceof FormData) && !headers['Content-Type']) {
    headers['Content-Type'] = 'application/json'
  }
  const tok = getToken()
  if (tok) headers['Authorization'] = `Bearer ${tok}`
  const res = await fetch(path, { ...init, headers })
  if (!res.ok) {
    const j = await res.json().catch(() => ({}))
    throw new Error((j as { error?: string }).error || res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}
