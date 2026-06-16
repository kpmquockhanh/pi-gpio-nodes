<template>
  <a-modal
    :open="true"
    title="Create Automation Rule"
    width="600px"
    @ok="save"
    @cancel="$emit('cancel')"
    ok-text="Create Rule"
    class="editor-modal"
  >
    <div class="form-section">
      <label>
        <Settings :size="14" />
        Rule Name
      </label>
      <a-input v-model:value="rule.name" placeholder="e.g., Turn on PC" />
    </div>

    <!-- Trigger Section -->
    <div class="form-section trigger-section">
      <h3>
        <Activity :size="16" />
        When (Trigger)
      </h3>
      <div class="form-row">
        <a-select v-model:value="rule.trigger.type" class="w-full">
          <a-select-option value="pin_state">Pin State Change</a-select-option>
          <a-select-option value="value_threshold">Value Threshold</a-select-option>
        </a-select>
      </div>
      
      <div class="form-row">
        <a-select v-model:value="rule.trigger.node" class="w-full">
          <a-select-option value="">Local Node</a-select-option>
          <a-select-option v-for="node in nodes" :key="node.id" :value="node.id">
            {{ node.name }}
          </a-select-option>
        </a-select>
      </div>
      
      <div class="form-row">
        <a-select v-model:value="rule.trigger.pin" class="w-full">
          <a-select-option value="">Select Pin</a-select-option>
          <a-select-option v-for="pin in pins" :key="pin.id" :value="pin.id">
            {{ pin.name }} (BCM {{ pin.bcm }})
          </a-select-option>
        </a-select>
      </div>
      
      <div class="form-row">
        <a-select v-model:value="rule.trigger.condition" class="w-full">
          <a-select-option value="HIGH">Is HIGH</a-select-option>
          <a-select-option value="LOW">Is LOW</a-select-option>
          <a-select-option value="rising_edge">Rising Edge (LOW → HIGH)</a-select-option>
          <a-select-option value="falling_edge">Falling Edge (HIGH → LOW)</a-select-option>
        </a-select>
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
          <a-select v-model:value="action.type" class="w-full">
            <a-select-option value="pin_action">Pin Action</a-select-option>
            <a-select-option value="delay">Delay</a-select-option>
          </a-select>
          
          <template v-if="action.type === 'pin_action'">
            <a-select v-model:value="action.node" class="w-full">
              <a-select-option value="">Local Node</a-select-option>
              <a-select-option v-for="node in nodes" :key="node.id" :value="node.id">
                {{ node.name }}
              </a-select-option>
            </a-select>
            <a-select v-model:value="action.pin" class="w-full">
              <a-select-option value="">Select Pin</a-select-option>
              <a-select-option v-for="pin in pins" :key="pin.id" :value="pin.id">
                {{ pin.name }}
              </a-select-option>
            </a-select>
            <a-select v-model:value="action.action" class="w-full">
              <a-select-option value="set">Set State</a-select-option>
              <a-select-option value="toggle">Toggle</a-select-option>
              <a-select-option value="pulse">Pulse</a-select-option>
              <a-select-option value="blink">Blink</a-select-option>
            </a-select>
          </template>
          
          <template v-if="action.type === 'delay'">
            <a-input-number
              v-model:value="action.params.ms"
              placeholder="Milliseconds"
              :min="100"
              :step="100"
              class="w-full"
            />
          </template>
        </div>
        <a-button type="primary" danger size="small" @click="removeAction(idx)">
          <X :size="14" />
        </a-button>
      </div>
      
      <a-button type="dashed" class="w-full" @click="addAction">
        <Plus :size="14" />
        Add Action
      </a-button>
    </div>
  </a-modal>
</template>

<script setup>
import { computed, reactive } from 'vue'
import {
  Settings, Activity, Zap, Plus, X
} from '@lucide/vue'

const props = defineProps({
  pins: { type: Array, default: () => [] },
})

const emit = defineEmits(['save', 'cancel'])

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

const nodes = computed(() => {
  const nodeMap = {}
  props.pins.forEach(pin => {
    if (!nodeMap[pin.nodeId]) {
      nodeMap[pin.nodeId] = { id: pin.nodeId, name: pin.nodeName }
    }
  })
  return Object.values(nodeMap)
})

function addAction() {
  rule.actions.push({
    type: 'pin_action',
    node: '',
    pin: '',
    action: 'set',
    params: {},
  })
}

function removeAction(idx) {
  rule.actions.splice(idx, 1)
}

function save() {
  if (!rule.name || !rule.trigger.pin) {
    alert('Please fill in all required fields')
    return
  }

  const id = 'auto-' + Date.now()

  emit('save', {
    id,
    ...rule,
  })
}
</script>

<style scoped>
.editor-modal :deep(.ant-modal-body) {
  max-height: 70vh;
  overflow-y: auto;
}

.form-section {
  margin-bottom: 16px;
  padding: 16px;
  background: var(--light-bg);
  border-radius: var(--radius-md);
}

.form-section h3 {
  color: var(--light-text-muted);
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

.form-row :deep(.ant-select),
.form-row :deep(.ant-input-number) {
  width: 100%;
}

label {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
  color: var(--light-text-muted);
  font-size: 13px;
}

.action-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  padding: 10px;
  background: rgba(13, 148, 136, 0.08);
  border-radius: var(--radius-md);
}

.action-number {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--primary-color);
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

.action-fields :deep(.ant-select),
.action-fields :deep(.ant-input-number) {
  flex: 1;
  min-width: 120px;
}

.trigger-section {
  border-left: 3px solid var(--success-color);
}

.actions-section {
  border-left: 3px solid var(--primary-color);
}

.w-full {
  width: 100%;
}
</style>
