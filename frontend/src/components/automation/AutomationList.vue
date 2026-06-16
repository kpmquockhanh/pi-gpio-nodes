<template>
  <div class="automation-list">
    <div class="automation-header">
      <h2>
        <Workflow :size="24" />
        Automation Rules
      </h2>
      <div class="flex items-center gap-2">
        <a-button type="primary" @click="showEditor = true">
          <Plus :size="16" />
          New Rule
        </a-button>
        <a-button @click="showVisualBuilder = true">
          <GitGraph :size="16" />
          Visual Builder
        </a-button>
      </div>
    </div>

    <a-empty v-if="automations.length === 0" description="No automation rules yet">
      <template #image>
        <Inbox :size="48" />
      </template>
      <template #description>
        <div class="empty-desc">
          <p>No automation rules yet</p>
          <p class="hint">Create rules to automate your GPIO pins</p>
        </div>
      </template>
    </a-empty>

    <div v-else class="automation-cards">
      <a-card
        v-for="auto in automations"
        :key="auto.id"
        class="automation-card"
        :class="{ disabled: !auto.enabled }"
        :bordered="true"
      >
        <div class="automation-header-row">
          <h3>
            <Settings :size="16" />
            {{ auto.name }}
          </h3>
          <div class="automation-controls">
            <a-button
              size="small"
              :type="auto.enabled ? 'primary' : 'default'"
              @click="toggleEnabled(auto)"
            >
              <component :is="auto.enabled ? ToggleRight : ToggleLeft" :size="14" />
              {{ auto.enabled ? 'ON' : 'OFF' }}
            </a-button>
            <a-button size="small" danger @click="deleteRule(auto.id)">
              <Trash2 :size="14" />
              Delete
            </a-button>
          </div>
        </div>

        <div class="automation-flow">
          <!-- Trigger -->
          <div class="flow-node trigger">
            <div class="node-label">
              <Activity :size="10" />
              WHEN
            </div>
            <div class="node-content">
              <span class="trigger-type">{{ auto.trigger.type }}</span>
              <span v-if="auto.trigger.pin" class="trigger-pin">
                <Pin :size="10" />
                {{ auto.trigger.pin }}
              </span>
              <span v-if="auto.trigger.condition" class="trigger-condition">
                <CheckCircle2 :size="10" />
                {{ auto.trigger.condition }}
              </span>
            </div>
          </div>

          <div class="flow-arrow">
            <ArrowRight :size="20" />
          </div>

          <!-- Actions -->
          <div class="flow-actions">
            <div
              v-for="(action, idx) in auto.actions"
              :key="idx"
              class="flow-node action"
            >
              <div class="node-label">
                <Zap :size="10" />
                THEN
              </div>
              <div class="node-content">
                <span class="action-type">{{ action.action || action.type }}</span>
                <span v-if="action.pin" class="action-pin">
                  <Pin :size="10" />
                  {{ action.pin }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </a-card>
    </div>

    <!-- Editor Modal -->
    <AutomationEditor
      v-if="showEditor"
      :pins="allPins"
      @save="saveAutomation"
      @cancel="showEditor = false"
    />

    <!-- Visual Builder Modal -->
    <a-modal
      v-model:open="showVisualBuilder"
      :footer="null"
      width="90%"
      wrap-class-name="visual-builder-modal"
      @cancel="showVisualBuilder = false"
    >
      <VisualBuilder @close="showVisualBuilder = false" />
    </a-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useNodesStore } from '../../store/nodes.js'
import {
  Workflow, Plus, Inbox, Settings, ToggleRight, ToggleLeft,
  Trash2, Activity, Pin, CheckCircle2, ArrowRight, Zap, GitGraph
} from '@lucide/vue'
import AutomationEditor from './AutomationEditor.vue'
import VisualBuilder from '../VisualBuilder.vue'

const nodesStore = useNodesStore()
const showEditor = ref(false)
const showVisualBuilder = ref(false)
const automations = ref([])

const allPins = computed(() => nodesStore.allPins)

onMounted(async () => {
  await loadAutomations()
})

async function loadAutomations() {
  try {
    const response = await fetch('/api/automations')
    const data = await response.json()
    automations.value = data.automations || []
  } catch (err) {
    console.error('Failed to load automations:', err)
  }
}

async function toggleEnabled(auto) {
  try {
    await fetch(`/api/automations/${auto.id}/enable`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled: !auto.enabled }),
    })
    auto.enabled = !auto.enabled
  } catch (err) {
    console.error('Failed to toggle automation:', err)
  }
}

async function deleteRule(id) {
  if (!confirm('Delete this automation rule?')) return
  try {
    await fetch(`/api/automations/${id}`, { method: 'DELETE' })
    automations.value = automations.value.filter(a => a.id !== id)
  } catch (err) {
    console.error('Failed to delete automation:', err)
  }
}

async function saveAutomation(rule) {
  try {
    const response = await fetch('/api/automations', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(rule),
    })
    if (response.ok) {
      showEditor.value = false
      await loadAutomations()
    }
  } catch (err) {
    console.error('Failed to save automation:', err)
  }
}
</script>

<style scoped>
.automation-list {
  padding: 20px;
}

.automation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.automation-header h2 {
  font-size: 20px;
  color: var(--light-text);
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
}

.empty-desc {
  text-align: center;
  color: var(--light-text-muted);
}

.empty-desc .hint {
  font-size: 14px;
  margin-top: 8px;
}

.automation-cards {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.automation-card {
  background: var(--light-surface);
  border-color: var(--light-border);
  transition: all 0.3s ease;
}

.automation-card.disabled {
  opacity: 0.6;
}

.automation-card :deep(.ant-card-body) {
  padding: 20px;
}

.automation-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.automation-header-row h3 {
  font-size: 16px;
  color: var(--light-text);
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
}

.automation-controls {
  display: flex;
  gap: 8px;
}

.automation-flow {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.flow-node {
  background: rgba(13, 148, 136, 0.1);
  border: 1px solid rgba(13, 148, 136, 0.25);
  border-radius: var(--radius-md);
  padding: 10px 14px;
  min-width: 140px;
}

.flow-node.trigger {
  background: rgba(16, 185, 129, 0.1);
  border-color: rgba(16, 185, 129, 0.25);
}

.flow-node.action {
  background: rgba(13, 148, 136, 0.1);
  border-color: rgba(13, 148, 136, 0.25);
}

.node-label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: var(--light-text-muted);
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.node-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.node-content span {
  display: flex;
  align-items: center;
  gap: 3px;
}

.trigger-type, .action-type {
  font-weight: 600;
  color: var(--light-text);
  font-size: 13px;
}

.trigger-pin, .action-pin {
  font-size: 12px;
  color: var(--secondary-color);
}

.trigger-condition {
  font-size: 11px;
  color: var(--success-color);
}

.flow-arrow {
  color: var(--light-text-muted);
  display: flex;
  align-items: center;
}

.flow-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

:deep(.visual-builder-modal .ant-modal-body) {
  padding: 0;
  height: 80vh;
}

:deep(.visual-builder-modal .ant-modal-content) {
  background: var(--light-bg);
}
</style>
