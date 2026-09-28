import { reactive } from 'vue'
import { api, getToken, setToken, clearToken } from '../api'

// 全局登录状态（响应式单例）
export const auth = reactive({
  token: getToken(),
  user: null,

  get isLoggedIn() {
    return !!this.token
  },

  async login(username, password) {
    const data = await api.post('/auth/login', { username, password })
    this.token = data.token
    this.user = data.user
    setToken(data.token)
    return data.user
  },

  async register(username, password, nickname) {
    const data = await api.post('/auth/register', { username, password, nickname })
    this.token = data.token
    this.user = data.user
    setToken(data.token)
    return data.user
  },

  async fetchMe() {
    if (!this.token) return null
    const data = await api.get('/me')
    this.user = data.user
    return data
  },

  // 本地更新余额（发帖/追加后即时反馈）
  setBalance(balance) {
    if (this.user) this.user.coinBalance = balance
  },

  logout() {
    this.token = ''
    this.user = null
    clearToken()
  },
})
