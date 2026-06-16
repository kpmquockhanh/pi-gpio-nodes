<template>
  <a-card class="log-viewer" :bordered="true">
    <template #title>
      <div class="log-header">
        <h3>
          <ActivitySquare :size="18" />
          Action Logs
        </h3>
        <div class="log-controls">
          <a-button
            size="small"
            :type="logsStore.autoRefresh ? 'primary' : 'default'"
            @click="logsStore.autoRefresh = !logsStore.autoRefresh"
          >
            <component :is="logsStore.autoRefresh ? PlayCircle : PauseCircle" :size="14" />
            {{ logsStore.autoRefresh ? 'Auto' : 'Manual' }}
          </a-button>
          <a-button type="primary" size="small" @click="refreshLogs">
            <RotateCw :size="14" />
            Refresh
          </a-button>
        </div>
      </div>
    </template>

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
          <a-tag size="small" class="log-action">
            <Zap :size="10" />
            {{ log.Action }}
          </a-tag>
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
    <a-empty v-else description="No actions recorded yet" class="log-empty">
      <template #image>
        <Inbox :size="32" />
      </template>
    </a-empty>
  </a-card>
</template>

<script setup>
import { onMounted, onBeforeUnmount } from 'vue'
import { useLogsStore } from '../store/logs.js'
import {
  ActivitySquare, RotateCw, PlayCircle, PauseCircle,
  Clock, Cpu, Pin, Zap, User, CheckCircle2, AlertCircle, Inbox
} from '@lucide/vue'

const logsStore = useLogsStore()

onMounted(() => {
  logsStore.fetchLogs(50)
  logsStore.startAutoRefresh()
})

onBeforeUnmount(() => {
  logsStore.stopAutoRefresh()
})

function refreshLogs() {
  logsStore.fetchLogs(50)
}

function formatTime(timestamp) {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleTimeString()
}
</script>

<style scoped>
.log-viewer {
  background: var(--light-surface);
  border-color: var(--light-border);
  max-height: 400px;
  display: flex;
  flex-direction: column;
}

.log-viewer :deep(.ant-card-body) {
  padding: 12px;
  overflow-y: auto;
  flex: 1;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.log-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--light-text);
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
}

.log-controls {
  display: flex;
  gap: 8px;
}

.log-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.log-entry {
  display: grid;
  grid-template-columns: 70px 1fr auto;
  gap: 8px;
  padding: 8px;
  background: var(--light-bg);
  border-radius: var(--radius-md);
  font-size: 13px;
  align-items: center;
}

.log-time {
  color: var(--light-text-muted);
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
  color: var(--secondary-color);
  font-weight: 500;
}

.log-pin {
  color: #8b5cf6;
}

.log-action {
  background: rgba(13, 148, 136, 0.15);
  color: var(--primary-color);
  border: none;
  text-transform: uppercase;
  font-size: 11px;
}

.log-trigger {
  color: var(--light-text-muted);
  font-size: 11px;
}

.log-result {
  font-size: 11px;
  color: var(--light-text-muted);
  display: flex;
  align-items: center;
  gap: 3px;
}

.log-result.success {
  color: var(--success-color);
}

.log-empty {
  text-align: center;
  padding: 20px;
  color: var(--light-text-muted);
  font-size: 14px;
}
</style>
