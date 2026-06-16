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
  background: var(--light-surface);
  border: 2px solid var(--light-border);
  border-radius: var(--radius-lg);
  padding: 12px;
  min-width: 120px;
}

.node-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--light-border);
}

.node-body {
  text-align: center;
}

.handle {
  width: 12px;
  height: 12px;
  background: var(--light-text-muted);
  border: 2px solid var(--light-surface);
  border-radius: 50%;
}

.text-gray-500 {
  color: var(--light-text-muted);
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
</style>
