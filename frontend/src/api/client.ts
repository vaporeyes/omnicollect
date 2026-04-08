// ABOUTME: Centralized fetch-based HTTP client for the OmniCollect REST API.
// ABOUTME: Replaces Wails IPC imports with typed HTTP calls to /api/v1/ endpoints.

const BASE_URL = (import.meta as any).env?.VITE_API_URL || ''

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let msg = `HTTP ${res.status}`
    try {
      const body = await res.json()
      if (body.error) msg = body.error
    } catch { /* use status text */ }
    throw new Error(msg)
  }
  return res.json()
}

export async function get<T>(path: string): Promise<T> {
  const res = await fetch(BASE_URL + path)
  return handleResponse<T>(res)
}

export async function post<T>(path: string, body: any): Promise<T> {
  const res = await fetch(BASE_URL + path, {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(body),
  })
  return handleResponse<T>(res)
}

export async function put<T>(path: string, body: any): Promise<T> {
  const res = await fetch(BASE_URL + path, {
    method: 'PUT',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(body),
  })
  return handleResponse<T>(res)
}

export async function del(path: string): Promise<void> {
  const res = await fetch(BASE_URL + path, {method: 'DELETE'})
  if (!res.ok) {
    let msg = `HTTP ${res.status}`
    try {
      const body = await res.json()
      if (body.error) msg = body.error
    } catch { /* use status text */ }
    throw new Error(msg)
  }
}

export async function postFile<T>(path: string, file: File, fieldName = 'image'): Promise<T> {
  const form = new FormData()
  form.append(fieldName, file)
  const res = await fetch(BASE_URL + path, {method: 'POST', body: form})
  return handleResponse<T>(res)
}

export async function downloadFile(path: string, body?: any): Promise<void> {
  const opts: RequestInit = body
    ? {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(body)}
    : {method: 'GET'}
  const res = await fetch(BASE_URL + path, opts)
  if (!res.ok) throw new Error(`Download failed: HTTP ${res.status}`)

  const blob = await res.blob()
  const disposition = res.headers.get('Content-Disposition') || ''
  const match = disposition.match(/filename="?([^"]+)"?/)
  const filename = match?.[1] || 'download'

  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}
