<template>
  <div class="automation-list">
    <div class="automation-header">
      <h2>
        <Workflow :size="24" />
        Automation Rules
      </h2>
      <div class="flex gap-2">
        <button class="btn btn-primary" @click="showEditor = true">
          <Plus :size="16" />
          New Rule
        </button>
        <button class="btn btn-secondary" @click="showVisualBuilder = true">
          <GitGraph :size="16" />
          Visual Builder
        </button>
      </div>
    </div>

    <div v-if="automations.length === 0" class="empty-state">
      <Inbox :size="48" />
      <p>No automation rules yet</p>
      <p class="hint">Create rules to automate your GPIO pins</p>
    </div>

    <div v-else class="automation-cards">
      <div 
        v-for="auto in automations" 
        :key="auto.id"
        class="automation-card"
        :class="{ disabled: !auto.enabled }"
      >
        <div class="automation-header-row">
          <h3>
            <Settings :size="16" />
            {{ auto.name }}
          </h3>
          <div class="automation-controls">
            <button 
              class="btn btn-sm"
              :class="auto.enabled ? 'btn-success' : 'btn-secondary'"
              @click="toggleEnabled(auto)"
            >
              <component :is="auto.enabled ? ToggleRight : ToggleLeft" :size="14" />
              {{ auto.enabled ? 'ON' : 'OFF' }}
            </button>
            <button class="btn btn-sm btn-danger" @click="deleteRule(auto.id)">
              <Trash2 :size="14" />
              Delete
            </button>
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
      </div>
    </div>

    <!-- Editor Modal -->
    <AutomationEditor 
      v-if="showEditor"
      :pins="allPins"
      @save="saveAutomation"
      @cancel="showEditor = false"
    />

    <!-- Visual Builder Modal -->
    <div v-if="showVisualBuilder" class="fixed inset-0 bg-black/50 z-50 flex flex-col">
      <div class="flex-1 bg-white m-4 rounded-xl overflow-hidden">
        <VisualBuilder @close="showVisualBuilder = false" />
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed } from 'vue'
import { useNodesStore } from '../../store/nodes.js'
import { 
  Workflow, Plus, Inbox, Settings, ToggleRight, ToggleLeft, 
  Trash2, Activity, Pin, CheckCircle2, ArrowRight, Zap, GitGraph
} from '@lucide/vue'
import AutomationEditor from './AutomationEditor.vue'
import VisualBuilder from '../VisualBuilder.vue'

export default {
  name: 'AutomationList',
  components: { 
    AutomationEditor, VisualBuilder,
    Workflow, Plus, Inbox, Settings, ToggleRight, ToggleLeft, 
    Trash2, Activity, Pin, CheckCircle2, ArrowRight, Zap, GitGraph
  },
  setup() {
    const nodesStore = useNodesStore()
    const showEditor = ref(false)
    const showVisualBuilder = ref(false)
    const automations = ref([])

    const allPins = computed(() => nodesStore.allPins)

    return {
      showEditor,
      showVisualBuilder,
      automations,
      allPins,
    }
  },
  async mounted() {
    await this.loadAutomations()
  },
  methods: {
    async loadAutomations() {
      try {
        const response = await fetch('/api/automations')
        const data = await response.json()
        this.automations = data.automations || []
      } catch (err) {
        console.error('Failed to load automations:', err)
      }
    },
    async toggleEnabled(auto) {
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
    },
    async deleteRule(id) {
      if (!confirm('Delete this automation rule?')) return
      try {
        await fetch(`/api/automations/${id}`, { method: 'DELETE' })
        this.automations = this.automations.filter(a => a.id !== id)
      } catch (err) {
        console.error('Failed to delete automation:', err)
      }
    },
    async saveAutomation(rule) {
      try {
        const response = await fetch('/api/automations', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(rule),
        })
        if (response.ok) {
          this.showEditor = false
          await this.loadAutomations()
        }
      } catch (err) {
        console.error('Failed to save automation:', err)
      }
    },
  },
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
  color: #f1f5f9;
  display: flex;
  align-items: center;
  gap: 10px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #64748b;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.empty-state .hint {
  font-size: 14px;
  margin-top: 8px;
}

.automation-cards {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.automation-card {
  background: linear-gradient(145deg, #1e293b, #0f172a);
  border: 1px solid #334155;
  border-radius: 12px;
  padding: 20px;
  transition: all 0.3s ease;
}

.automation-card.disabled {
  opacity: 0.6;
}

.automation-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.automation-header-row h3 {
  font-size: 16px;
  color: #f1f5f9;
  display: flex;
  align-items: center;
  gap: 8px;
}

.automation-controls {
  display: flex;
  gap: 8px;
}

.btn {
  padding: 7px 14px;
  border: none;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  color: white;
  display: flex;
  align-items: center;
  gap: 4px;
}

.btn-primary {
  background: #3b82f6;
}

.btn-primary:hover {
  background: #2563eb;
}

.btn-success {
  background: #22c55e;
}

.btn-secondary {
  background: #64748b;
}

.btn-danger {
  background: #ef4444;
}

.btn-sm {
  padding: 4px 10px;
  font-size: 12px;
}

.automation-flow {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.flow-node {
  background: rgba(59, 130, 246, 0.15);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 8px;
  padding: 10px 14px;
  min-width: 140px;
}

.flow-node.trigger {
  background: rgba(34, 197, 94, 0.15);
  border-color: rgba(34, 197, 94, 0.3);
}

.flow-node.action {
  background: rgba(59, 130, 246, 0.15);
  border-color: rgba(59, 130, 246, 0.3);
}

.node-label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: #94a3b8;
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
  color: #f1f5f9;
  font-size: 13px;
}

.trigger-pin, .action-pin {
  font-size: 12px;
  color: #60a5fa;
}

.trigger-condition {
  font-size: 11px;
  color: #4ade80;
}

.flow-arrow {
  color: #64748b;
  display: flex;
  align-items: center;
}

.flow-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
