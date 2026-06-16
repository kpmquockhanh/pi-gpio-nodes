<template>
  <div class="node trigger-node">
    <div class="node-header">
      <Zap :size="16" class="text-blue-500" />
      <span class="font-medium text-sm">Trigger</span>
    </div>
    <div class="node-body">
      <div class="text-xs text-gray-500 mb-2">{{ data.subtype || 'pin_state' }}</div>
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
        <select v-model="config.condition" class="w-full text-xs p-1 border rounded">
          <option value="HIGH">HIGH</option>
          <option value="LOW">LOW</option>
          <option value="rising_edge">Rising Edge</option>
          <option value="falling_edge">Falling Edge</option>
          <option value="change">Any Change</option>
        </select>
      </div>
    </div>
    <Handle type="source" position="bottom" class="handle" />
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Handle } from '@vue-flow/core'
import { Zap } from '@lucide/vue'

const props = defineProps({
  data: Object,
})

const config = ref(props.data.config || {})

watch(config, (newConfig) => {
  props.data.config = newConfig
}, { deep: true })
</script>

<style scoped>
.trigger-node {
  @apply bg-white border-2 border-blue-400 rounded-xl p-3 min-w-[180px];
}

.node-header {
  @apply flex items-center gap-2 mb-2 pb-2 border-b border-blue-100;
}

.node-body {
  @apply space-y-2;
}

.handle {
  @apply w-3 h-3 bg-blue-500 border-2 border-white rounded-full;
}
</style>
