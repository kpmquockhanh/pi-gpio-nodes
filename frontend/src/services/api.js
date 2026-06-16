import axios from 'axios'

// Create axios instance
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add API key to requests if available
const API_KEY = localStorage.getItem('api_key') || 'dev-api-key'
if (API_KEY) {
  api.defaults.headers.common['X-API-Key'] = API_KEY
}

// WebSocket Manager with automatic reconnect and polling fallback
class ConnectionManager {
  constructor() {
    this.ws = null
    this.url = null
    this.reconnectInterval = 3000
    this.maxReconnectAttempts = 5
    this.reconnectAttempts = 0
    this.pollingInterval = 2000
    this.pollingTimer = null
    this.listeners = {}
    this.isConnected = false
    this.connectionMode = 'disconnected' // 'websocket', 'polling', 'disconnected'
    this.shouldReconnect = true
  }

  connect(url, options = {}) {
    this.url = url
    this.options = options
    this.shouldReconnect = true
    this.reconnectAttempts = 0
    
    // Try WebSocket first
    this._connectWebSocket()
  }

  _connectWebSocket() {
    if (!this.shouldReconnect) return
    
    try {
      console.log('Connecting to WebSocket...')
      this.ws = new WebSocket(`${this.url}?api_key=${API_KEY}`)
      
      this.ws.onopen = () => {
        console.log('WebSocket connected')
        this.isConnected = true
        this.connectionMode = 'websocket'
        this.reconnectAttempts = 0
        this._stopPolling()
        this._emit('connected', { mode: 'websocket' })
      }
      
      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason)
        this.isConnected = false
        this.ws = null
        
        if (!this.shouldReconnect) {
          this.connectionMode = 'disconnected'
          this._emit('disconnected')
          return
        }
        
        // Start polling as fallback
        this.connectionMode = 'polling'
        this._startPolling()
        this._emit('disconnected', { willReconnect: true })
        
        // Attempt to reconnect WS
        this.reconnectAttempts++
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
          setTimeout(() => this._connectWebSocket(), this.reconnectInterval)
        }
      }
      
      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        this._emit('error', error)
      }
      
      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          this._emit('message', data)
          
          // Handle specific message types
          if (data.type === 'state_update') {
            this._emit('state_update', data)
          } else if (data.type === 'node_status') {
            this._emit('node_status', data)
          }
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e)
        }
      }
    } catch (err) {
      console.error('Failed to create WebSocket:', err)
      this._startPolling()
    }
  }

  _startPolling() {
    if (this.pollingTimer) return
    
    console.log('Starting HTTP polling fallback...')
    this.connectionMode = 'polling'
    this._emit('connected', { mode: 'polling' })
    
    this.pollingTimer = setInterval(async () => {
      if (this.connectionMode === 'websocket') {
        this._stopPolling()
        return
      }
      
      try {
        const response = await api.get('/api/nodes')
        if (response.data.nodes) {
          this._emit('polling_data', response.data.nodes)
        }
      } catch (e) {
        // Polling failed, keep trying
      }
    }, this.pollingInterval)
  }

  _stopPolling() {
    if (this.pollingTimer) {
      clearInterval(this.pollingTimer)
      this.pollingTimer = null
      console.log('Stopped HTTP polling')
    }
  }

  disconnect() {
    this.shouldReconnect = false
    this._stopPolling()
    
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    
    this.isConnected = false
    this.connectionMode = 'disconnected'
  }

  send(data) {
    if (this.ws && this.isConnected) {
      this.ws.send(JSON.stringify(data))
    }
  }

  // Event handling
  on(event, callback) {
    if (!this.listeners[event]) {
      this.listeners[event] = []
    }
    this.listeners[event].push(callback)
    
    // Return unsubscribe function
    return () => {
      this.listeners[event] = this.listeners[event].filter(cb => cb !== callback)
    }
  }

  _emit(event, data) {
    if (this.listeners[event]) {
      this.listeners[event].forEach(callback => {
        try {
          callback(data)
        } catch (e) {
          console.error('Event listener error:', e)
        }
      })
    }
  }

  getStatus() {
    return {
      isConnected: this.isConnected,
      mode: this.connectionMode,
      reconnectAttempts: this.reconnectAttempts,
    }
  }
}

export const connectionManager = new ConnectionManager()
export { api }