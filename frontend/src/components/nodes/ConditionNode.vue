<template>
  <div class="node condition-node">
    <Handle type="target" position="top" class="handle" />
    <div class="node-header">
      <GitBranch :size="16" class="text-purple-500" />
      <span class="font-medium text-sm">Condition</span>
    </div>
    <div class="node-body">
      <div class="space-y-2">
        <select v-model="config.type" class="w-full text-xs p-1 border rounded">
          <option value="pin_state">Pin State</option>
          <option value="value">Value</option>
          <option value="time">Time</option>
        </select>
        <select v-model="config.pin" class="w-full text-xs p-1 border rounded">
          <option value="">Select pin</option>
          <option value="pc-power">PC Power</option>
          <option value="status-led">Status LED</option>
          <option value="power-button">Power Button</option>
        </select>
        <select v-model="config.operator" class="w-full text-xs p-1 border rounded">
          <option value="equals">Equals</option>
          <option value="not_equals">Not Equals</option>
          <option value="greater_than">Greater Than</option>
          <option value="less_than">Less Than</option>
        </select>
        <input 
          v-model="config.value" 
          type="text" 
          placeholder="Value"
          class="w-full text-xs p-1 border rounded"
        />
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
  @apply bg-white border-2 border-purple-400 rounded-xl p-3 min-w-[180px];
}

.node-header {
  @apply flex items-center gap-2 mb-2 pb-2 border-b border-purple-100;
}

.node-body {
  @apply space-y-2;
}

.handle {
  @apply w-3 h-3 bg-purple-500 border-2 border-white rounded-full;
}
</style>
