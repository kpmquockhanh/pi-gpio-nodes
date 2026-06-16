<template>
  <div class="dashboard">
    <header class="header">
      <div class="header-left">
        <h1>
          <LayoutDashboard :size="28" class="header-icon" />
          Pi GPIO Dashboard
        </h1>
        <span class="subtitle">Multi-Node IoT Control</span>
      </div>
      <div class="header-right">
        <nav class="nav-tabs">
          <button 
            class="nav-tab" 
            :class="{ active: currentView === 'dashboard' }"
            @click="currentView = 'dashboard'"
          >
            <LayoutDashboard :size="16" />
            Dashboard
          </button>
          <button 
            class="nav-tab" 
            :class="{ active: currentView === 'automation' }"
            @click="currentView = 'automation'"
          >
            <Workflow :size="16" />
            Automation
          </button>
        </nav>
        <div class="flex items-center gap-3">
          <button 
            class="dark-mode-toggle"
            @click="toggleDarkMode"
            :title="isDarkMode ? 'Switch to light mode' : 'Switch to dark mode'"
          >
            <Sun v-if="isDarkMode" :size="18" />
            <Moon v-else :size="18" />
          </button>
          <div class="connection-badge" :class="connectionClass">
            <component :is="connectionIcon" :size="14" class="status-icon" />
            <span class="status-text">{{ connectionText }}</span>
          </div>
        </div>
      </div>
    </header>

    <main class="main">
      <!-- Dashboard View -->
      <div v-if="currentView === 'dashboard'">
        <div v-if="nodesStore.loading && nodesStore.nodes.length === 0" class="loading-state">
          <div class="spinner"></div>
          <p>Loading nodes...</p>
        </div>
        
        <div v-else-if="nodesStore.error" class="error-state">
          <AlertCircle :size="32" class="error-icon" />
          <p>{{ nodesStore.error }}</p>
          <button class="btn btn-primary" @click="retryConnection">
            <RotateCw :size="14" />
            Retry
          </button>
        </div>
        
        <div v-else class="content">
          <div class="nodes-section">
            <div v-for="node in nodesStore.nodes" :key="node.id" class="node-card">
              <div class="node-header">
                <div class="node-info">
                  <h2>
                    <Cpu :size="20" />
                    {{ node.name }}
                  </h2>
                  <span class="node-meta">{{ node.id }} • {{ node.role }}</span>
                </div>
                <div class="node-status" :class="node.status">
                  <component :is="node.status === 'online' ? CheckCircle2 : XCircle" :size="12" />
                  {{ node.status }}
                </div>
              </div>
              
              <div class="pins-grid">
                <PinCard
                  v-for="pin in Object.values(node.pins || {})"
                  :key="pin.id"
                  :pin="pin"
                  :node-id="node.id"
                  @action="handleAction"
                />
              </div>
            </div>
          </div>

          <div class="sidebar">
            <LogViewer />
          </div>
        </div>
      </div>

      <!-- Automation View -->
      <div v-else-if="currentView === 'automation'">
        <AutomationList />
      </div>
    </main>
  </div>
</template>

<script>
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { 
  LayoutDashboard, Workflow, Wifi, WifiOff, Activity, 
  RotateCw, CheckCircle2, XCircle, AlertCircle, Cpu,
  Sun, Moon
} from '@lucide/vue'
import PinCard from './components/PinCard.vue'
import LogViewer from './components/LogViewer.vue'
import AutomationList from './components/automation/AutomationList.vue'
import { useNodesStore } from './store/nodes.js'
import { useLogsStore } from './store/logs.js'
import { connectionManager } from './services/api.js'

export default {
  name: 'App',
  components: { 
    PinCard, LogViewer, AutomationList,
    LayoutDashboard, Workflow, Wifi, WifiOff, Activity, 
    RotateCw, CheckCircle2, XCircle, AlertCircle, Cpu,
    Sun, Moon
  },
  setup() {
    const nodesStore = useNodesStore()
    const logsStore = useLogsStore()
    const currentView = ref('dashboard')
    const isDarkMode = ref(localStorage.getItem('darkMode') === 'true' || 
      (!localStorage.getItem('darkMode') && window.matchMedia('(prefers-color-scheme: dark)').matches))

    // Apply dark mode class
    if (isDarkMode.value) {
      document.documentElement.classList.add('dark')
    }

    const connectionClass = computed(() => ({
      'status-connected': nodesStore.connectionStatus === 'connected',
      'status-polling': nodesStore.connectionStatus === 'polling',
      'status-disconnected': nodesStore.connectionStatus === 'disconnected',
    }))

    const connectionText = computed(() => {
      switch (nodesStore.connectionStatus) {
        case 'connected': return 'Live'
        case 'polling': return 'Polling'
        case 'disconnected': return 'Offline'
        default: return 'Unknown'
      }
    })

    const connectionIcon = computed(() => {
      switch (nodesStore.connectionStatus) {
        case 'connected': return Wifi
        case 'polling': return Activity
        case 'disconnected': return WifiOff
        default: return WifiOff
      }
    })

    const toggleDarkMode = () => {
      isDarkMode.value = !isDarkMode.value
      localStorage.setItem('darkMode', isDarkMode.value)
      if (isDarkMode.value) {
        document.documentElement.classList.add('dark')
      } else {
        document.documentElement.classList.remove('dark')
      }
    }

    return {
      nodesStore,
      logsStore,
      currentView,
      isDarkMode,
      toggleDarkMode,
      connectionClass,
      connectionText,
      connectionIcon,
    }
  },
  data() {
    return {
      unsubscribeEvents: [],
    }
  },
  async mounted() {
    await this.nodesStore.fetchNodes()

    const wsUrl = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`
    connectionManager.connect(wsUrl)

    this.unsubscribeEvents.push(
      connectionManager.on('connected', (data) => {
        this.nodesStore.setConnectionStatus(data.mode)
      })
    )

    this.unsubscribeEvents.push(
      connectionManager.on('disconnected', () => {
        if (connectionManager.connectionMode !== 'websocket') {
          this.nodesStore.setConnectionStatus('polling')
        }
      })
    )

    this.unsubscribeEvents.push(
      connectionManager.on('state_update', (data) => {
        this.nodesStore.updatePinState(data.node, data.pin, data.state)
      })
    )

    this.unsubscribeEvents.push(
      connectionManager.on('polling_data', (nodes) => {
        this.nodesStore.nodes = nodes
        this.nodesStore.lastUpdate = new Date()
      })
    )
  },
  beforeUnmount() {
    this.unsubscribeEvents.forEach(unsub => unsub())
    connectionManager.disconnect()
  },
  methods: {
    async handleAction({ nodeId, pinId, action, params }) {
      try {
        await this.nodesStore.executeAction(nodeId, pinId, action, params)
        this.logsStore.fetchLogs(50)
      } catch (err) {
        alert('Action failed: ' + err.message)
      }
    },
    retryConnection() {
      this.nodesStore.fetchNodes()
      const wsUrl = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`
      connectionManager.connect(wsUrl)
    },
  },
}
</script>

<style>
.dashboard {
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px;
  min-height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid #334155;
}

.header-left h1 {
  font-size: 28px;
  font-weight: 700;
  color: #f8fafc;
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-icon {
  color: #3b82f6;
  flex-shrink: 0;
}

.subtitle {
  font-size: 14px;
  color: #64748b;
  padding-left: 38px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.dark-mode-toggle {
  padding: 8px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.15s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dark-mode-toggle:hover {
  color: #f1f5f9;
  background: #334155;
}

.nav-tabs {
  display: flex;
  gap: 4px;
  background: #1e293b;
  padding: 4px;
  border-radius: 8px;
}

.nav-tab {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: #94a3b8;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  display: flex;
  align-items: center;
  gap: 6px;
}

.nav-tab:hover {
  color: #f1f5f9;
}

.nav-tab.active {
  background: #334155;
  color: #f1f5f9;
}

.connection-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 500;
}

.status-icon {
  flex-shrink: 0;
}

.status-connected {
  background: rgba(34, 197, 94, 0.15);
  color: #4ade80;
}

.status-polling {
  background: rgba(234, 179, 8, 0.15);
  color: #facc15;
}

.status-disconnected {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.content {
  display: grid;
  grid-template-columns: 1fr 350px;
  gap: 24px;
}

.nodes-section {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.node-card {
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 16px;
  padding: 24px;
}

.node-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #334155;
}

.node-info h2 {
  font-size: 18px;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.node-meta {
  font-size: 13px;
  color: #64748b;
  padding-left: 28px;
}

.node-status {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  display: flex;
  align-items: center;
  gap: 4px;
}

.node-status.online {
  background: rgba(34, 197, 94, 0.2);
  color: #4ade80;
}

.node-status.offline {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
}

.pins-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}

.loading-state, .error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #94a3b8;
}

.error-icon {
  color: #f87171;
  margin-bottom: 12px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #334155;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-state p {
  margin-bottom: 16px;
  color: #f87171;
}

.sidebar {
  position: sticky;
  top: 20px;
  align-self: start;
}

/* Responsive */
@media (max-width: 1024px) {
  .content {
    grid-template-columns: 1fr;
  }
  
  .sidebar {
    position: static;
  }
}

@media (max-width: 640px) {
  .dashboard {
    padding: 12px;
  }
  
  .header {
    flex-direction: column;
    gap: 12px;
    text-align: center;
  }
  
  .header-right {
    flex-direction: column;
    gap: 12px;
  }
  
  .subtitle {
    padding-left: 0;
  }
  
  .pins-grid {
    grid-template-columns: 1fr;
  }
}
</style>
