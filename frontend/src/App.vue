<template>
  <a-config-provider :theme="customTheme">
    <ApiKeySetup v-if="!apiKeyReady" @connected="onApiKeyConnected" />
    <div v-else class="dashboard">
      <header class="header">
        <div class="header-content">
          <div class="header-left">
            <div class="brand">
              <LayoutDashboard :size="28" class="header-icon" />
              <div class="brand-text">
                <h1>Pi GPIO Dashboard</h1>
                <span class="subtitle">Multi-Node IoT Control</span>
              </div>
            </div>
          </div>
          <div class="header-right">
            <a-segmented
              v-model:value="currentView"
              :options="[
                { label: 'Dashboard', value: 'dashboard' },
                { label: 'Automation', value: 'automation' },
              ]"
              class="nav-tabs"
            />
            <div class="header-actions">
              <a-tag :color="connectionColor" class="connection-badge">
                <component :is="connectionIcon" :size="14" />
                {{ connectionText }}
              </a-tag>
            </div>
          </div>
        </div>
      </header>

      <main class="main">
        <!-- Dashboard View -->
        <div v-if="currentView === 'dashboard'">
          <a-spin v-if="nodesStore.loading && nodesStore.nodes.length === 0" size="large" tip="Loading nodes...">
            <div class="loading-container" />
          </a-spin>
          
          <a-result
            v-else-if="nodesStore.error"
            status="error"
            :title="nodesStore.error"
            class="error-state"
          >
            <template #extra>
              <a-button type="primary" @click="retryConnection">
                <RotateCw :size="14" />
                Retry
              </a-button>
            </template>
          </a-result>
          
          <div v-else class="content">
            <div class="nodes-section">
              <a-card
                v-for="node in nodesStore.nodes"
                :key="node.id"
                class="node-card"
                :bordered="true"
              >
                <template #title>
                  <div class="node-header">
                    <div class="node-info">
                      <div class="node-title-row">
                        <h2>
                          <Cpu :size="20" />
                          {{ node.name }}
                        </h2>
                        <a-tag v-if="node.mock_gpio" color="warning" class="node-mock-badge">
                          <FlaskConical :size="12" />
                          Mock GPIO
                        </a-tag>
                      </div>
                      <span class="node-meta">{{ node.id }} • {{ node.role }}</span>
                    </div>
                    <a-tag :color="node.status === 'online' ? 'success' : 'error'" class="node-status">
                      <component :is="node.status === 'online' ? CheckCircle2 : XCircle" :size="12" />
                      {{ node.status }}
                    </a-tag>
                  </div>
                </template>
                
                <div v-if="Object.keys(node.pins || {}).length > 0" class="pins-grid">
                  <PinCard
                    v-for="pin in Object.values(node.pins || {})"
                    :key="pin.id"
                    :pin="pin"
                    :node-id="node.id"
                    @action="handleAction"
                  />
                </div>
                <div v-else class="empty-pins">
                  <div class="empty-pins-icon">
                    <Pin :size="32" />
                  </div>
                  <p class="empty-pins-text">No pins configured</p>
                  <p class="empty-pins-subtext">Add pins to this node in the config file to control them here.</p>
                </div>
              </a-card>
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
  </a-config-provider>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import {
  LayoutDashboard, Wifi, WifiOff, Activity,
  RotateCw, CheckCircle2, XCircle, Cpu, FlaskConical, Pin
} from '@lucide/vue'
import PinCard from './components/PinCard.vue'
import LogViewer from './components/LogViewer.vue'
import AutomationList from './components/automation/AutomationList.vue'
import ApiKeySetup from './components/ApiKeySetup.vue'
import { useNodesStore } from './store/nodes.js'
import { useLogsStore } from './store/logs.js'
import { connectionManager, hasApiKey } from './services/api.js'

const nodesStore = useNodesStore()
const logsStore = useLogsStore()
const currentView = ref('dashboard')
const apiKeyReady = ref(hasApiKey())

const connectionColor = computed(() => {
  switch (nodesStore.connectionStatus) {
    case 'connected': return 'success'
    case 'polling': return 'warning'
    case 'disconnected': return 'error'
    default: return 'default'
  }
})

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

const customTheme = {
  token: {
    colorPrimary: '#0d9488',
    colorSuccess: '#10b981',
    colorWarning: '#f59e0b',
    colorError: '#ef4444',
    colorInfo: '#06b6d4',
    colorBgBase: '#ffffff',
    colorTextBase: '#0f172a',
    borderRadius: 8,
    wireframe: false,
  },
}

const unsubscribeEvents = ref([])

function initConnection() {
  const wsUrl = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`
  connectionManager.connect(wsUrl)

  unsubscribeEvents.value.push(
    connectionManager.on('connected', (data) => {
      nodesStore.setConnectionStatus(data.mode)
    })
  )

  unsubscribeEvents.value.push(
    connectionManager.on('disconnected', () => {
      if (connectionManager.connectionMode !== 'websocket') {
        nodesStore.setConnectionStatus('polling')
      }
    })
  )

  unsubscribeEvents.value.push(
    connectionManager.on('state_update', (data) => {
      nodesStore.updatePinState(data.node, data.pin, data.state)
    })
  )

  unsubscribeEvents.value.push(
    connectionManager.on('polling_data', (nodes) => {
      nodesStore.nodes = nodes
      nodesStore.lastUpdate = new Date()
    })
  )
}

onMounted(async () => {
  if (apiKeyReady.value) {
    await nodesStore.fetchNodes()
    initConnection()
  }
})

onBeforeUnmount(() => {
  unsubscribeEvents.value.forEach(unsub => unsub())
  connectionManager.disconnect()
})

async function onApiKeyConnected() {
  apiKeyReady.value = true
  await nodesStore.fetchNodes()
  initConnection()
}

async function handleAction({ nodeId, pinId, action, params }) {
  try {
    await nodesStore.executeAction(nodeId, pinId, action, params)
    logsStore.fetchLogs(50)
  } catch (err) {
    alert('Action failed: ' + err.message)
  }
}

function retryConnection() {
  nodesStore.fetchNodes()
  const wsUrl = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`
  connectionManager.connect(wsUrl)
}
</script>

<style scoped>
.dashboard {
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px;
  min-height: 100vh;
  background: var(--light-bg);
}

.header {
  margin-bottom: 24px;
  padding: 20px 0;
  border-bottom: 1px solid var(--light-border);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  gap: 16px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.header-left h1 {
  font-size: 24px;
  font-weight: 700;
  color: var(--light-text);
  line-height: 1.2;
}

.header-icon {
  color: var(--primary-color);
  flex-shrink: 0;
}

.subtitle {
  font-size: 14px;
  color: var(--light-text-muted);
  line-height: 1.4;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.connection-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
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
  background: var(--light-surface);
  border-color: var(--light-border);
  border-radius: var(--radius-xl);
}

.node-card :deep(.ant-card-head) {
  border-bottom: 1px solid var(--light-border);
  padding: 16px 24px;
}

.node-card :deep(.ant-card-body) {
  padding: 20px;
}

.node-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.node-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.node-info h2 {
  font-size: 18px;
  font-weight: 600;
  color: var(--light-text);
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
}

.node-meta {
  font-size: 13px;
  color: var(--light-text-muted);
  padding-left: 28px;
}

.node-status {
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 600;
  text-transform: uppercase;
  font-size: 12px;
}

.node-mock-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 600;
  text-transform: uppercase;
  font-size: 11px;
  flex-shrink: 0;
}

.pins-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}

.loading-container {
  min-height: 200px;
}

.error-state {
  padding: 60px 20px;
}

.empty-pins {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  text-align: center;
  background: var(--light-bg);
  border-radius: var(--radius-lg);
  border: 1px dashed var(--light-border);
}

.empty-pins-icon {
  color: var(--light-text-muted);
  margin-bottom: 12px;
  opacity: 0.6;
}

.empty-pins-text {
  font-size: 15px;
  font-weight: 600;
  color: var(--light-text-muted);
  margin: 0 0 4px;
}

.empty-pins-subtext {
  font-size: 13px;
  color: var(--light-text-muted);
  opacity: 0.8;
  margin: 0;
  max-width: 320px;
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
  
  .header-content {
    flex-direction: column;
    gap: 16px;
    text-align: center;
  }

  .brand {
    justify-content: center;
  }

  .header-right {
    flex-direction: column;
    gap: 12px;
  }

  .pins-grid {
    grid-template-columns: 1fr;
  }
}
</style>
