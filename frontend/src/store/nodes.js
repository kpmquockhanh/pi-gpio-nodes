import { defineStore } from 'pinia'
import { api } from '../services/api.js'

export const useNodesStore = defineStore('nodes', {
  state: () => ({
    nodes: [],
    loading: false,
    error: null,
    connectionStatus: 'disconnected', // 'connected', 'polling', 'disconnected'
    lastUpdate: null,
  }),

  getters: {
    onlineNodes: (state) => state.nodes.filter(n => n.status === 'online'),
    offlineNodes: (state) => state.nodes.filter(n => n.status !== 'online'),
    allPins: (state) => {
      const pins = []
      state.nodes.forEach(node => {
        if (node.pins) {
          Object.values(node.pins).forEach(pin => {
            pins.push({ ...pin, nodeId: node.id, nodeName: node.name })
          })
        }
      })
      return pins
    },
    getNodeById: (state) => (id) => state.nodes.find(n => n.id === id),
    getPinById: (state) => (nodeId, pinId) => {
      const node = state.nodes.find(n => n.id === nodeId)
      return node?.pins?.[pinId] || null
    },
  },

  actions: {
    async fetchNodes() {
      this.loading = true
      this.error = null
      try {
        const response = await api.get('/api/nodes')
        this.nodes = response.data.nodes || []
        this.lastUpdate = new Date()
      } catch (err) {
        this.error = err.message
        console.error('Failed to fetch nodes:', err)
      } finally {
        this.loading = false
      }
    },

    updatePinState(nodeId, pinId, state) {
      const node = this.nodes.find(n => n.id === nodeId)
      if (node && node.pins && node.pins[pinId]) {
        node.pins[pinId].state = state
        node.pins[pinId].last_update = new Date().toISOString()
      }
    },

    updateNodeStatus(nodeId, status) {
      const node = this.nodes.find(n => n.id === nodeId)
      if (node) {
        node.status = status
      }
    },

    setConnectionStatus(status) {
      this.connectionStatus = status
    },

    async executeAction(nodeId, pinId, action, params = {}) {
      try {
        const response = await api.post(
          `/api/nodes/${nodeId}/pins/${pinId}/action`,
          { action, params }
        )
        // Optimistically update state
        if (response.data.new_state !== undefined) {
          this.updatePinState(nodeId, pinId, response.data.new_state)
        }
        return response.data
      } catch (err) {
        console.error('Action failed:', err)
        throw err
      }
    },
  },
})