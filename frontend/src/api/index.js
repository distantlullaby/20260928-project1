// 极简 API 封装：自动携带 JWT，统一错误处理
const TOKEN_KEY = 'mbti_token'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || ''
}
export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token)
}
export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

async function request(method, path, body) {
  const headers = { 'Content-Type': 'application/json' }
  const token = getToken()
  if (token) headers.Authorization = `Bearer ${token}`

  const res = await fetch(`/api${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })

  let data = null
  const text = await res.text()
  if (text) {
    try { data = JSON.parse(text) } catch { data = { raw: text } }
  }

  if (!res.ok) {
    const message = (data && data.error) || `请求失败 (${res.status})`
    const err = new Error(message)
    err.status = res.status
    throw err
  }
  return data
}

export const api = {
  get: (p) => request('GET', p),
  post: (p, b) => request('POST', p, b),
}
