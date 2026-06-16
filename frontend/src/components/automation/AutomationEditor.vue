<template>
  <div class="modal-overlay" @click.self="$emit('cancel')">
    <div class="modal-content editor-modal">
      <h2>
        <Workflow :size="24" />
        Create Automation Rule
      </h2>
      
      <div class="form-section">
        <label>
          <Settings :size="14" />
          Rule Name
        </label>
        <input v-model="rule.name" type="text" placeholder="e.g., Turn on PC" />
      </div>

      <!-- Trigger Section -->
      <div class="form-section trigger-section">
        <h3>
          <Activity :size="16" />
          When (Trigger)
        </h3>
        <div class="form-row">
          <select v-model="rule.trigger.type">
            <option value="pin_state">Pin State Change</option>
            <option value="value_threshold">Value Threshold</option>
          </select>
        </div>
        
        <div class="form-row">
          <select v-model="rule.trigger.node">
            <option value="">Local Node</option>
            <option v-for="node in nodes" :key="node.id" :value="node.id">
              {{ node.name }}
            </option>
          </select>
        </div>
        
        <div class="form-row">
          <select v-model="rule.trigger.pin">
            <option value="">Select Pin</option>
            <option v-for="pin in pins" :key="pin.id" :value="pin.id">
              {{ pin.name }} (BCM {{ pin.bcm }})
            </option>
          </select>
        </div>
        
        <div class="form-row">
          <select v-model="rule.trigger.condition">
            <option value="HIGH">Is HIGH</option>
            <option value="LOW">Is LOW</option>
            <option value="rising_edge">Rising Edge (LOW → HIGH)</option>
            <option value="falling_edge">Falling Edge (HIGH → LOW)</option>
          </select>
        </div>
      </div>

      <!-- Actions Section -->
      <div class="form-section actions-section">
        <h3>
          <Zap :size="16" />
          Then (Actions)
        </h3>
        <div 
          v-for="(action, idx) in rule.actions" 
          :key="idx"
          class="action-row"
        >
          <div class="action-number">{{ idx + 1 }}</div>
          <div class="action-fields">
            <select v-model="action.type">
              <option value="pin_action">Pin Action</option>
              <option value="delay">Delay</option>
            </select>
            
            <template v-if="action.type === 'pin_action'">
              <select v-model="action.node">
                <option value="">Local Node</option>
                <option v-for="node in nodes" :key="node.id" :value="node.id">
                  {{ node.name }}
                </option>
              </select>
              <select v-model="action.pin">
                <option value="">Select Pin</option>
                <option v-for="pin in pins" :key="pin.id" :value="pin.id">
                  {{ pin.name }}
                </option>
              </select>
              <select v-model="action.action">
                <option value="set">Set State</option>
                <option value="toggle">Toggle</option>
                <option value="pulse">Pulse</option>
                <option value="blink">Blink</option>
              </select>
            </template>
            
            <template v-if="action.type === 'delay'">
              <input 
                v-model.number="action.params.ms" 
                type="number" 
                placeholder="Milliseconds"
                min="100"
                step="100"
              />
            </template>
          </div>
          <button class="btn btn-danger btn-sm" @click="removeAction(idx)">
            <X :size="14" />
          </button>
        </div>
        
        <button class="btn btn-secondary" @click="addAction">
          <Plus :size="14" />
          Add Action
        </button>
      </div>

      <div class="modal-actions">
        <button class="btn btn-primary" @click="save">
          <Check :size="14" />
          Create Rule
        </button>
        <button class="btn btn-secondary" @click="$emit('cancel')">
          <X :size="14" />
          Cancel
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { reactive } from 'vue'
import { 
  Workflow, Settings, Activity, Zap, Plus, X, Check
} from '@lucide/vue'

export default {
  name: 'AutomationEditor',
  components: { 
    Workflow, Settings, Activity, Zap, Plus, X, Check
  },
  props: {
    pins: { type: Array, default: () => [] },
  },
  emits: ['save', 'cancel'],
  setup() {
    const rule = reactive({
      name: '',
      enabled: true,
      trigger: {
        type: 'pin_state',
        node: '',
        pin: '',
        condition: 'HIGH',
      },
      actions: [
        {
          type: 'pin_action',
          node: '',
          pin: '',
          action: 'set',
          params: {},
        },
      ],
    })

    return { rule }
  },
  computed: {
    nodes() {
      // Get unique nodes from pins
      const nodeMap = {}
      this.pins.forEach(pin => {
        if (!nodeMap[pin.nodeId]) {
          nodeMap[pin.nodeId] = { id: pin.nodeId, name: pin.nodeName }
        }
      })
      return Object.values(nodeMap)
    },
  },
  methods: {
    addAction() {
      this.rule.actions.push({
        type: 'pin_action',
        node: '',
        pin: '',
        action: 'set',
        params: {},
      })
    },
    removeAction(idx) {
      this.rule.actions.splice(idx, 1)
    },
    save() {
      if (!this.rule.name || !this.rule.trigger.pin) {
        alert('Please fill in all required fields')
        return
      }
      
      // Generate ID
      const id = 'auto-' + Date.now()
      
      this.$emit('save', {
        id,
        ...this.rule,
      })
    },
  },
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  overflow-y: auto;
  padding: 20px;
}

.editor-modal {
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 16px;
  padding: 32px;
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}

.editor-modal h2 {
  color: #f1f5f9;
  margin-bottom: 24px;
  font-size: 20px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.form-section {
  margin-bottom: 24px;
  padding: 16px;
  background: rgba(15, 23, 42, 0.5);
  border-radius: 8px;
}

.form-section h3 {
  color: #94a3b8;
  font-size: 14px;
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 1px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-row {
  margin-bottom: 10px;
}

.form-row select,
.form-row input {
  width: 100%;
  padding: 10px 12px;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #e2e8f0;
  font-size: 14px;
}

.form-row select:focus,
.form-row input:focus {
  outline: none;
  border-color: #3b82f6;
}

label {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
  color: #94a3b8;
  font-size: 13px;
}

input[type="text"] {
  width: 100%;
  padding: 10px 12px;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #e2e8f0;
  font-size: 14px;
}

.action-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  padding: 10px;
  background: rgba(59, 130, 246, 0.1);
  border-radius: 6px;
}

.action-number {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #3b82f6;
  color: white;
  border-radius: 50%;
  font-size: 12px;
  font-weight: 700;
  flex-shrink: 0;
}

.action-fields {
  flex: 1;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.action-fields select,
.action-fields input {
  flex: 1;
  min-width: 120px;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #334155;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  color: white;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-primary {
  background: #3b82f6;
}

.btn-primary:hover {
  background: #2563eb;
}

.btn-secondary {
  background: #64748b;
}

.btn-secondary:hover {
  background: #475569;
}

.btn-danger {
  background: #ef4444;
  padding: 6px;
}

.btn-danger:hover {
  background: #dc2626;
}

.btn-sm {
  padding: 4px 10px;
  font-size: 12px;
}

.trigger-section {
  border-left: 3px solid #22c55e;
}

.actions-section {
  border-left: 3px solid #3b82f6;
}
</style>
