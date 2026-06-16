import { defineStore } from 'pinia'
import { api } from '../services/api.js'

export const useLogsStore = defineStore('logs', {
  state: () => ({
    logs: [],
    loading: false,
    error: null,
    autoRefresh: true,
    refreshInterval: 5000,
  }),

  getters: {
    recentLogs: (state) => state.logs.slice(0, 50),
    logsByNode: (state) => {
      const grouped = {}
      state.logs.forEach(log => {
        if (!grouped[log.NodeID]) grouped[log.NodeID] = []
        grouped[log.NodeID].push(log)
      })
      return grouped
    },
  },

  actions: {
    async fetchLogs(limit = 50, nodeId = '') {
      this.loading = true
      try {
        const params = { limit }
        if (nodeId) params.node = nodeId
        const response = await api.get('/api/logs', { params })
        this.logs = response.data.logs || []
      } catch (err) {
        this.error = err.message
        console.error('Failed to fetch logs:', err)
      } finally {
        this.loading = false
      }
    },

    addLog(log) {
      this.logs.unshift(log)
      // Keep only last 100 logs in memory
      if (this.logs.length > 100) {
        this.logs = this.logs.slice(0, 100)
      }
    },

    startAutoRefresh() {
      if (this.refreshTimer) return
      this.refreshTimer = setInterval(() => {
        if (this.autoRefresh) {
          this.fetchLogs(50)
        }
      }, this.refreshInterval)
    },

    stopAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer)
        this.refreshTimer = null
      }
    },
  },
})