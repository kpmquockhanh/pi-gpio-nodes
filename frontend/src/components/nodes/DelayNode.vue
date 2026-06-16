<template>
  <div class="node delay-node">
    <Handle type="target" position="top" class="handle" />
    <div class="node-header">
      <Timer :size="16" class="text-yellow-500" />
      <span class="font-medium text-sm">Delay</span>
    </div>
    <div class="node-body">
      <div class="space-y-2">
        <a-input-number v-model:value="config.ms" placeholder="Duration (ms)" size="small" class="w-full" />
        <a-select v-model:value="config.unit" size="small" class="w-full">
          <a-select-option value="ms">Milliseconds</a-select-option>
          <a-select-option value="s">Seconds</a-select-option>
          <a-select-option value="m">Minutes</a-select-option>
        </a-select>
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
  background: var(--light-surface);
  border: 2px solid var(--warning-color);
  border-radius: var(--radius-lg);
  padding: 12px;
  min-width: 150px;
}

.node-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(245, 158, 11, 0.2);
}

.node-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.handle {
  width: 12px;
  height: 12px;
  background: var(--warning-color);
  border: 2px solid var(--light-surface);
  border-radius: 50%;
}

.text-yellow-500 {
  color: var(--warning-color);
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
