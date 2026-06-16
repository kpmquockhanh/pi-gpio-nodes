<template>
  <div class="visual-builder">
    <div class="builder-header">
      <h2>Visual Rule Builder</h2>
      <p class="builder-desc">Drag and drop nodes to create automation rules</p>
      
      <div class="header-actions">
        <a-button type="primary" @click="saveRule">
          <SaveIcon :size="18" />
          Save Rule
        </a-button>
        <a-button @click="clearCanvas">
          <TrashIcon :size="18" />
          Clear
        </a-button>
        <a-button @click="loadRule">
          <FolderOpenIcon :size="18" />
          Load Rule
        </a-button>
      </div>
    </div>

    <div class="builder-container">
      <!-- Sidebar with node types -->
      <div class="sidebar">
        <h3>Node Types</h3>
        
        <div class="node-group">
          <div class="node-group-title">Triggers</div>
          <div
            v-for="type in triggerTypes"
            :key="type.type"
            class="node-template"
            draggable="true"
            @dragstart="onDragStart($event, type)"
          >
            <div class="node-template-header">
              <component :is="type.icon" :size="20" class="node-icon-blue" />
              <span class="font-medium">{{ type.label }}</span>
            </div>
            <div class="node-template-desc">{{ type.description }}</div>
          </div>
        </div>

        <div class="node-group">
          <div class="node-group-title">Actions</div>
          <div
            v-for="type in actionTypes"
            :key="type.type"
            class="node-template"
            draggable="true"
            @dragstart="onDragStart($event, type)"
          >
            <div class="node-template-header">
              <component :is="type.icon" :size="20" class="node-icon-green" />
              <span class="font-medium">{{ type.label }}</span>
            </div>
            <div class="node-template-desc">{{ type.description }}</div>
          </div>
        </div>

        <div class="node-group">
          <div class="node-group-title">Logic</div>
          <div
            v-for="type in logicTypes"
            :key="type.type"
            class="node-template"
            draggable="true"
            @dragstart="onDragStart($event, type)"
          >
            <div class="node-template-header">
              <component :is="type.icon" :size="20" class="node-icon-purple" />
              <span class="font-medium">{{ type.label }}</span>
            </div>
            <div class="node-template-desc">{{ type.description }}</div>
          </div>
        </div>
      </div>

      <!-- Vue Flow Canvas -->
      <div class="canvas-container">
        <VueFlow
          v-model="elements"
          :default-zoom="1"
          :min-zoom="0.2"
          :max-zoom="4"
          @connect="onConnect"
          @dragover="onDragOver"
          @drop="onDrop"
          fit-view-on-init
        >
          <Background pattern-color="#BBDBF1" :gap="20" />
          <Controls />
          <MiniMap />
          
          <!-- Custom nodes -->
          <template #node-trigger="props">
            <TriggerNode :data="props.data" />
          </template>
          
          <template #node-action="props">
            <ActionNode :data="props.data" />
          </template>
          
          <template #node-condition="props">
            <ConditionNode :data="props.data" />
          </template>
          
          <template #node-delay="props">
            <DelayNode :data="props.data" />
          </template>
          
          <template #node-end="props">
            <EndNode :data="props.data" />
          </template>
        </VueFlow>
      </div>
    </div>

    <!-- Rule Save Modal -->
    <a-modal
      v-model:open="showSaveModal"
      title="Save Rule"
      @ok="confirmSave"
      @cancel="showSaveModal = false"
      ok-text="Save"
    >
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Rule Name</label>
          <a-input
            v-model:value="ruleName"
            placeholder="Enter rule name"
          />
        </div>
        
        <div class="flex items-center gap-2">
          <a-checkbox v-model:checked="ruleEnabled">Enable rule</a-checkbox>
        </div>
      </div>
    </a-modal>

    <!-- Rule Load Modal -->
    <a-modal
      v-model:open="showLoadModal"
      title="Load Rule"
      @ok="confirmLoad"
      @cancel="showLoadModal = false"
      ok-text="Load"
      :ok-button-props="{ disabled: !selectedRule }"
    >
      <div class="space-y-2 max-h-60 overflow-y-auto">
        <div
          v-for="rule in existingRules"
          :key="rule.id"
          @click="selectRule(rule)"
          class="rule-item"
          :class="{ 'rule-selected': selectedRule?.id === rule.id }"
        >
          <div class="font-medium">{{ rule.name }}</div>
          <div class="text-sm text-gray-500">
            {{ rule.enabled ? 'Enabled' : 'Disabled' }} • {{ rule.actions?.length || 0 }} actions
          </div>
        </div>
        
        <a-empty v-if="existingRules.length === 0" description="No rules found" />
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { VueFlow } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'

import { 
  Zap, 
  ToggleLeft, 
  Clock, 
  Bell, 
  Activity,
  ArrowRight,
  GitBranch,
  Timer,
  Save as SaveIcon,
  Trash2 as TrashIcon,
  FolderOpen,
  MousePointerClick,
  Gauge
} from '@lucide/vue'

import TriggerNode from './nodes/TriggerNode.vue'
import ActionNode from './nodes/ActionNode.vue'
import ConditionNode from './nodes/ConditionNode.vue'
import DelayNode from './nodes/DelayNode.vue'
import EndNode from './nodes/EndNode.vue'

import { api } from '../services/api'

const triggerTypes = [
  { type: 'trigger', subtype: 'pin_state', label: 'Pin State', description: 'Trigger on pin state change', icon: Zap },
  { type: 'trigger', subtype: 'value_threshold', label: 'Threshold', description: 'Trigger on value crossing', icon: Gauge },
  { type: 'trigger', subtype: 'timer', label: 'Timer', description: 'Trigger on interval', icon: Clock },
  { type: 'trigger', subtype: 'long_press', label: 'Long Press', description: 'Trigger on button hold', icon: MousePointerClick },
]

const actionTypes = [
  { type: 'action', subtype: 'pin_action', label: 'Pin Action', description: 'Control a pin', icon: ToggleLeft },
  { type: 'action', subtype: 'notify', label: 'Notify', description: 'Send notification', icon: Bell },
  { type: 'action', subtype: 'delay', label: 'Delay', description: 'Wait for duration', icon: Timer },
]

const logicTypes = [
  { type: 'condition', label: 'Condition', description: 'Check condition', icon: GitBranch },
  { type: 'end', label: 'End', description: 'End of flow', icon: ArrowRight },
]

const elements = ref([])
const showSaveModal = ref(false)
const showLoadModal = ref(false)
const ruleName = ref('')
const ruleEnabled = ref(true)
const existingRules = ref([])
const selectedRule = ref(null)

function onDragStart(event, nodeType) {
  event.dataTransfer.setData('application/vueflow', JSON.stringify(nodeType))
  event.dataTransfer.effectAllowed = 'move'
}

function onDragOver(event) {
  event.preventDefault()
  event.dataTransfer.dropEffect = 'move'
}

function onDrop(event) {
  const type = JSON.parse(event.dataTransfer.getData('application/vueflow'))
  
  const position = {
    x: event.clientX - event.target.getBoundingClientRect().left,
    y: event.clientY - event.target.getBoundingClientRect().top,
  }

  const newNode = {
    id: `${type.type}-${Date.now()}`,
    type: type.type,
    position,
    data: {
      label: type.label,
      subtype: type.subtype,
      config: {},
    },
  }

  elements.value.push(newNode)
}

function onConnect(connection) {
  const edge = {
    id: `e${connection.source}-${connection.target}`,
    source: connection.source,
    target: connection.target,
    animated: true,
  }
  elements.value.push(edge)
}

function clearCanvas() {
  if (confirm('Clear all nodes?')) {
    elements.value = []
  }
}

function saveRule() {
  showSaveModal.value = true
  ruleName.value = ''
  ruleEnabled.value = true
}

async function confirmSave() {
  if (!ruleName.value) {
    alert('Please enter a rule name')
    return
  }

  const nodes = elements.value.filter(e => !e.source)
  const edges = elements.value.filter(e => e.source)

  const triggers = nodes.filter(n => n.type === 'trigger')
  const actions = nodes.filter(n => n.type === 'action')

  if (triggers.length === 0) {
    alert('Rule must have at least one trigger')
    return
  }

  if (actions.length === 0) {
    alert('Rule must have at least one action')
    return
  }

  const rule = {
    name: ruleName.value,
    enabled: ruleEnabled.value,
    trigger: buildTriggerFromNode(triggers[0]),
    actions: actions.map(buildActionFromNode),
  }

  try {
    await api.post('/automations', rule)
    showSaveModal.value = false
    alert('Rule saved!')
  } catch (error) {
    console.error('Failed to save rule:', error)
    alert('Failed to save rule')
  }
}

function buildTriggerFromNode(node) {
  return {
    type: node.data.subtype || 'pin_state',
    node: node.data.config.node || '',
    pin: node.data.config.pin || '',
    condition: node.data.config.condition || 'HIGH',
    threshold: node.data.config.threshold || 0,
    duration_ms: node.data.config.duration_ms || 1000,
    interval: node.data.config.interval || '',
  }
}

function buildActionFromNode(node) {
  return {
    type: node.data.subtype || 'pin_action',
    node: node.data.config.node || '',
    pin: node.data.config.pin || '',
    action: node.data.config.action || 'toggle',
    params: node.data.config.params || {},
  }
}

async function loadRule() {
  try {
    const response = await api.get('/automations')
    existingRules.value = response.data.automations || []
    showLoadModal.value = true
    selectedRule.value = null
  } catch (error) {
    console.error('Failed to load rules:', error)
    alert('Failed to load rules')
  }
}

function selectRule(rule) {
  selectedRule.value = rule
}

function confirmLoad() {
  if (!selectedRule.value) return

  const rule = selectedRule.value
  elements.value = []

  const triggerNode = {
    id: 'trigger-0',
    type: 'trigger',
    position: { x: 250, y: 50 },
    data: {
      label: 'Trigger',
      subtype: rule.trigger.type,
      config: rule.trigger,
    },
  }
  elements.value.push(triggerNode)

  let prevNode = triggerNode
  rule.actions.forEach((action, index) => {
    const actionNode = {
      id: `action-${index}`,
      type: 'action',
      position: { x: 250, y: 150 + index * 100 },
      data: {
        label: action.action || 'Action',
        subtype: action.type,
        config: action,
      },
    }
    elements.value.push(actionNode)

    elements.value.push({
      id: `e${prevNode.id}-${actionNode.id}`,
      source: prevNode.id,
      target: actionNode.id,
      animated: true,
    })

    prevNode = actionNode
  })

  const endNode = {
    id: 'end',
    type: 'end',
    position: { x: 250, y: 150 + rule.actions.length * 100 },
    data: { label: 'End' },
  }
  elements.value.push(endNode)

  elements.value.push({
    id: `e${prevNode.id}-end`,
    source: prevNode.id,
    target: 'end',
    animated: true,
  })

  showLoadModal.value = false
}

onMounted(() => {
  elements.value = [
    {
      id: 'trigger-0',
      type: 'trigger',
      position: { x: 250, y: 50 },
      data: { label: 'Trigger', subtype: 'pin_state', config: {} },
    },
    {
      id: 'action-0',
      type: 'action',
      position: { x: 250, y: 200 },
      data: { label: 'Action', subtype: 'pin_action', config: {} },
    },
    {
      id: 'e-trigger-0-action-0',
      source: 'trigger-0',
      target: 'action-0',
      animated: true,
    },
  ]
})
</script>

<style scoped>
.visual-builder {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--light-bg);
}

.builder-header {
  padding: 1rem 2rem;
  border-bottom: 1px solid var(--light-border);
  background: var(--light-surface);
}

.builder-header h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--light-text);
  margin-bottom: 4px;
}

.builder-desc {
  color: var(--light-text-muted);
  margin-bottom: 16px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.builder-container {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.sidebar {
  overflow-y: auto;
  background: var(--light-surface);
  border-right: 1px solid var(--light-border);
  padding: 16px;
  width: 256px;
}

.sidebar h3 {
  font-weight: 600;
  margin-bottom: 16px;
  color: var(--light-text);
}

.node-group {
  margin-bottom: 24px;
}

.node-group-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--light-text-muted);
  text-transform: uppercase;
  margin-bottom: 8px;
}

.node-template {
  padding: 12px;
  background: var(--light-bg);
  border: 1px solid var(--light-border);
  border-radius: var(--radius-md);
  cursor: move;
  margin-bottom: 8px;
  transition: all 0.2s;
}

.node-template:hover {
  border-color: var(--primary-color);
  transform: translateX(4px);
}

.node-template-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.node-template-desc {
  font-size: 12px;
  color: var(--light-text-muted);
  margin-top: 4px;
}

.node-icon-blue { color: var(--primary-color); }
.node-icon-green { color: var(--success-color); }
.node-icon-purple { color: var(--secondary-color); }

.canvas-container {
  position: relative;
  flex: 1;
}

.rule-item {
  padding: 12px;
  border: 1px solid var(--light-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 8px;
}

.rule-item:hover {
  background: var(--light-bg);
}

.rule-selected {
  background: rgba(13, 148, 136, 0.08);
  border-color: var(--primary-color);
}

:deep(.vue-flow) {
  height: 100%;
}

.space-y-4 > * + * {
  margin-top: 16px;
}

.space-y-2 > * + * {
  margin-top: 8px;
}

.block {
  display: block;
}

.font-medium {
  font-weight: 500;
}

.text-sm {
  font-size: 14px;
}

.text-gray-500 {
  color: var(--light-text-muted);
}

.text-gray-700 {
  color: var(--light-text);
}

.mb-1 {
  margin-bottom: 4px;
}

.max-h-60 {
  max-height: 240px;
}

.overflow-y-auto {
  overflow-y: auto;
}
</style>
