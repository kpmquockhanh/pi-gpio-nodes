<template>
  <div class="visual-builder">
    <div class="builder-header">
      <h2 class="text-2xl font-bold mb-2">Visual Rule Builder</h2>
      <p class="text-gray-600 mb-4">Drag and drop nodes to create automation rules</p>
      
      <div class="flex gap-2 mb-4">
        <button
          @click="saveRule"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2"
        >
          <SaveIcon :size="18" />
          Save Rule
        </button>
        <button
          @click="clearCanvas"
          class="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors flex items-center gap-2"
        >
          <TrashIcon :size="18" />
          Clear
        </button>
        <button
          @click="loadRule"
          class="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors flex items-center gap-2"
        >
          <FolderOpenIcon :size="18" />
          Load Rule
        </button>
      </div>
    </div>

    <div class="builder-container">
      <!-- Sidebar with node types -->
      <div class="sidebar bg-gray-50 border-r border-gray-200 p-4 w-64">
        <h3 class="font-semibold mb-4 text-gray-700">Node Types</h3>
        
        <div class="space-y-3">
          <div class="text-sm font-medium text-gray-500 uppercase">Triggers</div>
          <div
            v-for="type in triggerTypes"
            :key="type.type"
            class="node-template p-3 bg-white rounded-lg border border-gray-200 cursor-move hover:shadow-md transition-shadow"
            draggable="true"
            @dragstart="onDragStart($event, type)"
          >
            <div class="flex items-center gap-2">
              <component :is="type.icon" :size="20" class="text-blue-500" />
              <span class="font-medium">{{ type.label }}</span>
            </div>
            <div class="text-xs text-gray-500 mt-1">{{ type.description }}</div>
          </div>
        </div>

        <div class="space-y-3 mt-6">
          <div class="text-sm font-medium text-gray-500 uppercase">Actions</div>
          <div
            v-for="type in actionTypes"
            :key="type.type"
            class="node-template p-3 bg-white rounded-lg border border-gray-200 cursor-move hover:shadow-md transition-shadow"
            draggable="true"
            @dragstart="onDragStart($event, type)"
          >
            <div class="flex items-center gap-2">
              <component :is="type.icon" :size="20" class="text-green-500" />
              <span class="font-medium">{{ type.label }}</span>
            </div>
            <div class="text-xs text-gray-500 mt-1">{{ type.description }}</div>
          </div>
        </div>

        <div class="space-y-3 mt-6">
          <div class="text-sm font-medium text-gray-500 uppercase">Logic</div>
          <div
            v-for="type in logicTypes"
            :key="type.type"
            class="node-template p-3 bg-white rounded-lg border border-gray-200 cursor-move hover:shadow-md transition-shadow"
            draggable="true"
            @dragstart="onDragStart($event, type)"
          >
            <div class="flex items-center gap-2">
              <component :is="type.icon" :size="20" class="text-purple-500" />
              <span class="font-medium">{{ type.label }}</span>
            </div>
            <div class="text-xs text-gray-500 mt-1">{{ type.description }}</div>
          </div>
        </div>
      </div>

      <!-- Vue Flow Canvas -->
      <div class="canvas-container flex-1">
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
    <div v-if="showSaveModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-96 max-w-full mx-4">
        <h3 class="text-lg font-semibold mb-4">Save Rule</h3>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Rule Name</label>
            <input
              v-model="ruleName"
              type="text"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter rule name"
            />
          </div>
          
          <div class="flex items-center gap-2">
            <input
              v-model="ruleEnabled"
              type="checkbox"
              id="enabled"
              class="rounded text-blue-600"
            />
            <label for="enabled" class="text-sm">Enable rule</label>
          </div>
        </div>

        <div class="flex gap-2 mt-6">
          <button
            @click="showSaveModal = false"
            class="flex-1 px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
          >
            Cancel
          </button>
          <button
            @click="confirmSave"
            class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Save
          </button>
        </div>
      </div>
    </div>

    <!-- Rule Load Modal -->
    <div v-if="showLoadModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-96 max-w-full mx-4">
        <h3 class="text-lg font-semibold mb-4">Load Rule</h3>
        
        <div class="space-y-2 max-h-60 overflow-y-auto">
          <div
            v-for="rule in existingRules"
            :key="rule.id"
            @click="selectRule(rule)"
            class="p-3 border border-gray-200 rounded-lg cursor-pointer hover:bg-gray-50 transition-colors"
            :class="{ 'bg-blue-50 border-blue-300': selectedRule?.id === rule.id }"
          >
            <div class="font-medium">{{ rule.name }}</div>
            <div class="text-sm text-gray-500">
              {{ rule.enabled ? 'Enabled' : 'Disabled' }} • {{ rule.actions?.length || 0 }} actions
            </div>
          </div>
          
          <div v-if="existingRules.length === 0" class="text-center text-gray-500 py-4">
            No rules found
          </div>
        </div>

        <div class="flex gap-2 mt-6">
          <button
            @click="showLoadModal = false"
            class="flex-1 px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
          >
            Cancel
          </button>
          <button
            @click="confirmLoad"
            class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            :disabled="!selectedRule"
            :class="{ 'opacity-50 cursor-not-allowed': !selectedRule }"
          >
            Load
          </button>
        </div>
      </div>
    </div>
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

// Node type definitions
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

// State
const elements = ref([])
const showSaveModal = ref(false)
const showLoadModal = ref(false)
const ruleName = ref('')
const ruleEnabled = ref(true)
const existingRules = ref([])
const selectedRule = ref(null)

// Drag and drop
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

// Canvas operations
function clearCanvas() {
  if (confirm('Clear all nodes?')) {
    elements.value = []
  }
}

// Save rule
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

  // Validate: must have at least one trigger and one action
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

  // Build rule from flow
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

// Load rule
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

  // Create trigger node
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

  // Create action nodes
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

    // Connect nodes
    elements.value.push({
      id: `e${prevNode.id}-${actionNode.id}`,
      source: prevNode.id,
      target: actionNode.id,
      animated: true,
    })

    prevNode = actionNode
  })

  // Add end node
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
  // Add initial nodes
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
}

.builder-header {
  padding: 1rem 2rem;
  border-bottom: 1px solid #e5e7eb;
}

.builder-container {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.sidebar {
  overflow-y: auto;
}

.canvas-container {
  position: relative;
}

.node-template {
  transition: all 0.2s;
}

.node-template:hover {
  transform: translateX(4px);
}

:deep(.vue-flow) {
  height: 100%;
}
</style>
