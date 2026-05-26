import { defineStore } from 'pinia'

const DEFAULT_MS = {
  success: 5500,
  error: 8000,
  info: 6000,
  warn: 7000
}

export const useToastStore = defineStore('toast', {
  state: () => ({
    items: [],
    timers: {}
  }),
  actions: {
    push(message, type = 'success', duration) {
      const id = Date.now() + Math.random()
      const ms = duration ?? DEFAULT_MS[type] ?? 5500
      this.items.push({ id, message, type })
      this.timers[id] = setTimeout(() => this.remove(id), ms)
      return id
    },
    remove(id) {
      if (this.timers[id]) {
        clearTimeout(this.timers[id])
        delete this.timers[id]
      }
      this.items = this.items.filter(t => t.id !== id)
    }
  }
})
