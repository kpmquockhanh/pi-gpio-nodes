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
        <select v-model="config.node" class="w-full text-xs p-1 border rounded">
          <option value="">Select node</option>
          <option value="local">Local</option>
        </select>
        <select v-model="config.pin" class="w-full text-xs p-1 border rounded">
          <option value="">Select pin</option>
          <option value="pc-power">PC Power</option>
          <option value="status-led">Status LED</option>
          <option value="power-button">Power Button</option>
        </select>
        <select v-model="config.action" class="w-full text-xs p-1 border rounded">
          <option value="toggle">Toggle</option>
          <option value="pulse">Pulse</option>
          <option value="set">Set</option>
          <option value="blink">Blink</option>
        </select>
        <div v-if="config.action === 'pulse' || config.action === 'blink'">
          <input 
            v-model="config.params.ms" 
            type="number" 
            placeholder="Duration (ms)"
            class="w-full text-xs p-1 border rounded"
          />
        </div>
        <div v-if="config.action === 'set'">
          <select v-model="config.params.value" class="w-full text-xs p-1 border rounded">
            <option :value="true">HIGH</option>
            <option :value="false">LOW</option>
          </select>
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
  @apply bg-white border-2 border-green-400 rounded-xl p-3 min-w-[180px];
}

.node-header {
  @apply flex items-center gap-2 mb-2 pb-2 border-b border-green-100;
}

.node-body {
  @apply space-y-2;
}

.handle {
  @apply w-3 h-3 bg-green-500 border-2 border-white rounded-full;
}
</style>
