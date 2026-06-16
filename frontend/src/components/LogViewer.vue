<template>
  <div class="log-viewer">
    <div class="log-header">
      <h3>
        <ActivitySquare :size="18" />
        Action Logs
      </h3>
      <div class="log-controls">
        <button 
          class="btn btn-sm" 
          :class="logsStore.autoRefresh ? 'btn-active' : 'btn-inactive'"
          @click="logsStore.autoRefresh = !logsStore.autoRefresh"
        >
          <component :is="logsStore.autoRefresh ? PlayCircle : PauseCircle" :size="14" />
          {{ logsStore.autoRefresh ? 'Auto' : 'Manual' }}
        </button>
        <button class="btn btn-sm btn-primary" @click="refreshLogs">
          <RotateCw :size="14" />
          Refresh
        </button>
      </div>
    </div>

    <div class="log-list" v-if="logsStore.logs.length > 0">
      <div 
        v-for="log in logsStore.recentLogs" 
        :key="log.ID"
        class="log-entry"
        :class="`log-${log.TriggeredBy}`"
      >
        <div class="log-time">
          <Clock :size="10" />
          {{ formatTime(log.Timestamp) }}
        </div>
        <div class="log-content">
          <span class="log-node">
            <Cpu :size="10" />
            {{ log.NodeID }}
          </span>
          <span class="log-pin">
            <Pin :size="10" />
            {{ log.PinID || 'system' }}
          </span>
          <span class="log-action">
            <Zap :size="10" />
            {{ log.Action }}
          </span>
          <span class="log-trigger">
            <User :size="10" />
            {{ log.TriggeredBy }}
          </span>
        </div>
        <div class="log-result" :class="log.Result === 'success' ? 'success' : ''">
          <component :is="log.Result === 'success' ? CheckCircle2 : AlertCircle" :size="12" />
          {{ log.Result }}
        </div>
      </div>
    </div>
    <div v-else class="log-empty">
      <Inbox :size="32" />
      <p>No actions recorded yet</p>
    </div>
  </div>
</template>

<script>
import { useLogsStore } from '../store/logs.js'
import { 
  ActivitySquare, RotateCw, PlayCircle, PauseCircle, 
  Clock, Cpu, Pin, Zap, User, CheckCircle2, AlertCircle, Inbox
} from '@lucide/vue'

export default {
  name: 'LogViewer',
  components: {
    ActivitySquare, RotateCw, PlayCircle, PauseCircle, 
    Clock, Cpu, Pin, Zap, User, CheckCircle2, AlertCircle, Inbox
  },
  setup() {
    const logsStore = useLogsStore()
    return { logsStore }
  },
  mounted() {
    this.logsStore.fetchLogs(50)
    this.logsStore.startAutoRefresh()
  },
  beforeUnmount() {
    this.logsStore.stopAutoRefresh()
  },
  methods: {
    refreshLogs() {
      this.logsStore.fetchLogs(50)
    },
    formatTime(timestamp) {
      if (!timestamp) return '-'
      const date = new Date(timestamp * 1000)
      return date.toLocaleTimeString()
    },
  },
}
</script>

<style scoped>
.log-viewer {
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 12px;
  padding: 16px;
  max-height: 400px;
  display: flex;
  flex-direction: column;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #334155;
}

.log-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #f1f5f9;
  display: flex;
  align-items: center;
  gap: 8px;
}

.log-controls {
  display: flex;
  gap: 8px;
}

.btn-sm {
  padding: 4px 10px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
  font-weight: 500;
}

.btn-active {
  background: rgba(34, 197, 94, 0.2);
  color: #4ade80;
}

.btn-inactive {
  background: rgba(100, 116, 139, 0.2);
  color: #94a3b8;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover {
  background: #2563eb;
}

.log-list {
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.log-entry {
  display: grid;
  grid-template-columns: 70px 1fr auto;
  gap: 8px;
  padding: 8px;
  background: rgba(15, 23, 42, 0.5);
  border-radius: 6px;
  font-size: 13px;
  align-items: center;
}

.log-time {
  color: #64748b;
  font-size: 11px;
  font-family: monospace;
  display: flex;
  align-items: center;
  gap: 4px;
}

.log-content {
  display: flex;
  gap: 6px;
  align-items: center;
  flex-wrap: wrap;
}

.log-content span {
  display: flex;
  align-items: center;
  gap: 3px;
}

.log-node {
  color: #60a5fa;
  font-weight: 500;
}

.log-pin {
  color: #a78bfa;
}

.log-action {
  background: rgba(59, 130, 246, 0.2);
  color: #60a5fa;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
  text-transform: uppercase;
}

.log-trigger {
  color: #64748b;
  font-size: 11px;
}

.log-result {
  font-size: 11px;
  color: #94a3b8;
  display: flex;
  align-items: center;
  gap: 3px;
}

.log-result.success {
  color: #4ade80;
}

.log-empty {
  text-align: center;
  padding: 20px;
  color: #64748b;
  font-size: 14px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

/* Scrollbar styling */
.log-list::-webkit-scrollbar {
  width: 6px;
}

.log-list::-webkit-scrollbar-track {
  background: transparent;
}

.log-list::-webkit-scrollbar-thumb {
  background: #475569;
  border-radius: 3px;
}
</style>
