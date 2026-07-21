const TOKEN_KEY = 'token'
const USER_ID_KEY = 'user_id'
const USERNAME_KEY = 'username'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(t) {
  localStorage.setItem(TOKEN_KEY, t)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_ID_KEY)
  localStorage.removeItem(USERNAME_KEY)
}

export function getUserId() {
  return localStorage.getItem(USER_ID_KEY) || null
}

export function setUserId(id) {
  localStorage.setItem(USER_ID_KEY, id)
}

export function getUsername() {
  return localStorage.getItem(USERNAME_KEY) || ''
}

export function setUsername(name) {
  localStorage.setItem(USERNAME_KEY, name)
}

export async function api(endpoint, method = 'GET', body = null, isFormData = false) {
  const headers = {}
  const token = getToken()
  if (token && !endpoint.includes('/login') && !endpoint.includes('/register')) {
    headers['Authorization'] = `Bearer ${token}`
  }
  const options = { method, headers }
  if (body) {
    options.body = isFormData ? body : JSON.stringify(body)
    if (!isFormData) headers['Content-Type'] = 'application/json'
  }
  const res = await fetch(endpoint, options)
  let data
  try { data = await res.json() } catch { data = { error: await res.text() } }
  if (!res.ok) {
    if (res.status === 401) {
      clearToken()
      throw new Error(data.error || '登录已过期，请重新登录')
    }
    if (res.status === 429) throw new Error('请求过于频繁，请稍后再试')
    throw new Error(data.error || '请求失败')
  }
  return data
}
