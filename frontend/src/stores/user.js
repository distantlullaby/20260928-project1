import { defineStore } from 'pinia'
import api from '../api'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('mbti_token') || '',
    user: null
  }),
  getters: {
    isLoggedIn: (s) => !!s.token,
    coins: (s) => s.user?.coin_balance ?? 0,
    frozen: (s) => s.user?.frozen_balance ?? 0
  },
  actions: {
    setAuth(token, user) {
      this.token = token
      this.user = user
      localStorage.setItem('mbti_token', token)
    },
    async fetchMe() {
      if (!this.token) return null
      try {
        const { user } = await api.get('/me')
        this.user = user
        return user
      } catch {
        this.logout()
        return null
      }
    },
    async login(username, password) {
      const { token, user } = await api.post('/auth/login', { username, password })
      this.setAuth(token, user)
    },
    async register(username, password, nickname) {
      const { token, user } = await api.post('/auth/register', { username, password, nickname })
      this.setAuth(token, user)
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('mbti_token')
    }
  }
})
