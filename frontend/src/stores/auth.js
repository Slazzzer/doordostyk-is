import { defineStore } from 'pinia'
import api from '../api.js'

const TOKEN_KEY = 'dd_token'
const USER_KEY = 'dd_user'

function storage() {
  return sessionStorage
}

/** Убрать устаревшую сессию из localStorage (раньше логинились «навсегда»). */
export function clearLegacyAuthStorage() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: '',
    user: null
  }),
  getters: {
    isAuth: s => !!s.token,
    isCustomer: s => s.user?.type === 'customer',
    isUser: s => s.user?.type === 'user',
    isAdmin: s => s.user?.role === 'administrator',
    isSeller: s => s.user?.role === 'seller' || s.user?.role === 'administrator',
    isStoreman: s => s.user?.role === 'storekeeper' || s.user?.role === 'administrator',
    displayName: s => s.user?.full_name || s.user?.email || 'Гость'
  },
  actions: {
    /** Публичные страницы: не показывать прошлую сессию в шапке. */
    enterAsGuest() {
      this.token = ''
      this.user = null
    },
    restore() {
      this.token = storage().getItem(TOKEN_KEY) || ''
      const u = storage().getItem(USER_KEY)
      this.user = u ? JSON.parse(u) : null
    },
    setSession(token, user) {
      this.token = token
      this.user = user
      storage().setItem(TOKEN_KEY, token)
      storage().setItem(USER_KEY, JSON.stringify(user))
    },
    async loginUser(login, password) {
      const r = await api.post('/auth/login/user', { login, password })
      this.setSession(r.data.token, {
        type: 'user',
        user_id: r.data.user_id,
        role: r.data.role,
        full_name: r.data.full_name
      })
    },
    async loginCustomer(email, password) {
      const r = await api.post('/auth/login/customer', { email, password })
      this.setSession(r.data.token, {
        type: 'customer',
        user_id: r.data.user_id,
        full_name: r.data.full_name
      })
    },
    async register(payload) {
      const r = await api.post('/auth/register', payload)
      this.setSession(r.data.token, {
        type: 'customer',
        user_id: r.data.user_id,
        full_name: r.data.full_name
      })
    },
    logout() {
      this.token = ''
      this.user = null
      storage().removeItem(TOKEN_KEY)
      storage().removeItem(USER_KEY)
      clearLegacyAuthStorage()
    }
  }
})
