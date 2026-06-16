<template>
  <div class="node end-node">
    <Handle type="target" position="top" class="handle" />
    <div class="node-header">
      <ArrowRight :size="16" class="text-gray-500" />
      <span class="font-medium text-sm">End</span>
    </div>
    <div class="node-body">
      <div class="text-xs text-gray-500">End of flow</div>
    </div>
  </div>
</template>

<script setup>
import { Handle } from '@vue-flow/core'
import { ArrowRight } from '@lucide/vue'
</script>

<style scoped>
.end-node {
  @apply bg-white border-2 border-gray-400 rounded-xl p-3 min-w-[120px];
}

.node-header {
  @apply flex items-center gap-2 mb-2 pb-2 border-b border-gray-100;
}

.node-body {
  @apply text-center;
}

.handle {
  @apply w-3 h-3 bg-gray-500 border-2 border-white rounded-full;
}
</style>
