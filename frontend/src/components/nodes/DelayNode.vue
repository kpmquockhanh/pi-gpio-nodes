<template>
  <div class="node delay-node">
    <Handle type="target" position="top" class="handle" />
    <div class="node-header">
      <Timer :size="16" class="text-yellow-500" />
      <span class="font-medium text-sm">Delay</span>
    </div>
    <div class="node-body">
      <div class="space-y-2">
        <input 
          v-model="config.ms" 
          type="number" 
          placeholder="Duration (ms)"
          class="w-full text-xs p-1 border rounded"
        />
        <select v-model="config.unit" class="w-full text-xs p-1 border rounded">
          <option value="ms">Milliseconds</option>
          <option value="s">Seconds</option>
          <option value="m">Minutes</option>
        </select>
      </div>
    </div>
    <Handle type="source" position="bottom" class="handle" />
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Handle } from '@vue-flow/core'
import { Timer } from '@lucide/vue'

const props = defineProps({
  data: Object,
})

const config = ref(props.data.config || {})

watch(config, (newConfig) => {
  props.data.config = newConfig
}, { deep: true })
</script>

<style scoped>
.delay-node {
  @apply bg-white border-2 border-yellow-400 rounded-xl p-3 min-w-[150px];
}

.node-header {
  @apply flex items-center gap-2 mb-2 pb-2 border-b border-yellow-100;
}

.node-body {
  @apply space-y-2;
}

.handle {
  @apply w-3 h-3 bg-yellow-500 border-2 border-white rounded-full;
}
</style>
