<template>
  <div class="node condition-node">
    <Handle type="target" position="top" class="handle" />
    <div class="node-header">
      <GitBranch :size="16" class="text-purple-500" />
      <span class="font-medium text-sm">Condition</span>
    </div>
    <div class="node-body">
      <div class="space-y-2">
        <a-select v-model:value="config.type" size="small" class="w-full">
          <a-select-option value="pin_state">Pin State</a-select-option>
          <a-select-option value="value">Value</a-select-option>
          <a-select-option value="time">Time</a-select-option>
        </a-select>
        <a-select v-model:value="config.pin" size="small" class="w-full">
          <a-select-option value="">Select pin</a-select-option>
          <a-select-option value="pc-power">PC Power</a-select-option>
          <a-select-option value="status-led">Status LED</a-select-option>
          <a-select-option value="power-button">Power Button</a-select-option>
        </a-select>
        <a-select v-model:value="config.operator" size="small" class="w-full">
          <a-select-option value="equals">Equals</a-select-option>
          <a-select-option value="not_equals">Not Equals</a-select-option>
          <a-select-option value="greater_than">Greater Than</a-select-option>
          <a-select-option value="less_than">Less Than</a-select-option>
        </a-select>
        <a-input v-model:value="config.value" placeholder="Value" size="small" class="w-full" />
      </div>
    </div>
    <Handle type="source" position="bottom" class="handle" id="true" />
    <Handle type="source" position="right" class="handle" id="false" />
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Handle } from '@vue-flow/core'
import { GitBranch } from '@lucide/vue'

const props = defineProps({
  data: Object,
})

const config = ref(props.data.config || {})

watch(config, (newConfig) => {
  props.data.config = newConfig
}, { deep: true })
</script>

<style scoped>
.condition-node {
  background: var(--light-surface);
  border: 2px solid var(--secondary-color);
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
  border-bottom: 1px solid rgba(99, 102, 241, 0.2);
}

.node-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.handle {
  width: 12px;
  height: 12px;
  background: var(--secondary-color);
  border: 2px solid var(--light-surface);
  border-radius: 50%;
}

.text-purple-500 {
  color: var(--secondary-color);
}

.font-medium {
  font-weight: 500;
}

.text-sm {
  font-size: 14px;
}

.space-y-2 > * + * {
  margin-top: 8px;
}

.w-full {
  width: 100%;
}
</style>
