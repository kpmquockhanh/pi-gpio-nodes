<template>
  <div class="node action-node">
    <Handle type="target" position="top" class="handle" />
    <div class="node-header">
      <ToggleLeft :size="16" class="text-green-500" />
      <span class="font-medium text-sm">Action</span>
    </div>
    <div class="node-body">
      <div class="text-xs text-gray-500 mb-2">{{ data.subtype || 'pin_action' }}</div>
      <div class="space-y-2">
        <a-select v-model:value="config.node" size="small" class="w-full">
          <a-select-option value="">Select node</a-select-option>
          <a-select-option value="local">Local</a-select-option>
        </a-select>
        <a-select v-model:value="config.pin" size="small" class="w-full">
          <a-select-option value="">Select pin</a-select-option>
          <a-select-option value="pc-power">PC Power</a-select-option>
          <a-select-option value="status-led">Status LED</a-select-option>
          <a-select-option value="power-button">Power Button</a-select-option>
        </a-select>
        <a-select v-model:value="config.action" size="small" class="w-full">
          <a-select-option value="toggle">Toggle</a-select-option>
          <a-select-option value="pulse">Pulse</a-select-option>
          <a-select-option value="set">Set</a-select-option>
          <a-select-option value="blink">Blink</a-select-option>
        </a-select>
        <div v-if="config.action === 'pulse' || config.action === 'blink'">
          <a-input-number
            v-model:value="config.params.ms"
            placeholder="Duration (ms)"
            size="small"
            class="w-full"
          />
        </div>
        <div v-if="config.action === 'set'">
          <a-select v-model:value="config.params.value" size="small" class="w-full">
            <a-select-option :value="true">HIGH</a-select-option>
            <a-select-option :value="false">LOW</a-select-option>
          </a-select>
        </div>
      </div>
    </div>
    <Handle type="source" position="bottom" class="handle" />
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Handle } from '@vue-flow/core'
import { ToggleLeft } from '@lucide/vue'

const props = defineProps({
  data: Object,
})

const config = ref(props.data.config || {})

watch(config, (newConfig) => {
  props.data.config = newConfig
}, { deep: true })
</script>

<style scoped>
.action-node {
  background: var(--light-surface);
  border: 2px solid var(--success-color);
  border-radius: var(--radius-lg);
  padding: 12px;
  min-width: 180px;
}

.node-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(16, 185, 129, 0.2);
}

.node-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.handle {
  width: 12px;
  height: 12px;
  background: var(--success-color);
  border: 2px solid var(--light-surface);
  border-radius: 50%;
}

.text-green-500 {
  color: var(--success-color);
}

.font-medium {
  font-weight: 500;
}

.text-sm {
  font-size: 14px;
}

.text-xs {
  font-size: 12px;
}

.text-gray-500 {
  color: var(--light-text-muted);
}

.mb-2 {
  margin-bottom: 8px;
}

.space-y-2 > * + * {
  margin-top: 8px;
}

.w-full {
  width: 100%;
}
</style>
