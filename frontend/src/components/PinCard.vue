<template>
  <a-card
    class="pin-card"
    :class="{ 'pin-on': isOn, 'pin-active': isExecuting }"
    :bordered="true">
    <div class="pin-header">
      <div class="pin-icon" :class="pin.type">
        <component :is="pinIcon" :size="20" />
      </div>
      <div class="pin-info">
        <h3 class="pin-name">{{ pin.name }}</h3>
        <span class="pin-details">BCM {{ pin.bcm }} | {{ pin.mode }}</span>
      </div>
      <a-tag :color="stateColor" class="pin-state-badge">
        <component :is="stateIcon" :size="12" />
        {{ stateText }}
      </a-tag>
    </div>

    <div class="pin-actions">
      <a-button
        v-if="canToggle"
        :type="isOn ? 'primary' : 'default'"
        :class="['btn-toggle', { 'btn-on': isOn }]"
        @click="execute('toggle')"
        :loading="isExecuting"
        size="small">
        <component :is="isOn ? ToggleRight : ToggleLeft" :size="14" />
        {{ isOn ? "ON" : "OFF" }}
      </a-button>
      <a-button
        v-if="canPulse"
        type="primary"
        class="btn-warning"
        @click="executePulseDirect"
        :loading="isExecuting"
        size="small">
        <Zap :size="14" />
        Press
      </a-button>
      <a-button
        v-if="canPulse"
        type="default"
        @click="showPulse = true"
        :loading="isExecuting"
        size="small">
        <Settings :size="14" />
        Custom
      </a-button>
      <a-button
        v-if="canSet"
        type="primary"
        class="btn-on"
        @click="execute('set', { state: true })"
        :loading="isExecuting"
        :disabled="isOn"
        size="small">
        <Power :size="14" />
        ON
      </a-button>
      <a-button
        v-if="canSet"
        type="default"
        danger
        class="btn-off"
        @click="execute('set', { state: false })"
        :loading="isExecuting"
        :disabled="!isOn"
        size="small">
        <PowerOff :size="14" />
        OFF
      </a-button>
      <a-button
        v-if="canBlink"
        type="default"
        class="btn-warning"
        @click="showBlink = true"
        :loading="isExecuting"
        size="small">
        <Eye :size="14" />
        Blink
      </a-button>
      <a-button
        v-if="canRead"
        type="default"
        class="btn-info"
        @click="execute('read')"
        :loading="isExecuting"
        size="small">
        <Activity :size="14" />
        Read
      </a-button>
    </div>

    <div v-if="pin.last_update" class="pin-timestamp">
      <Clock :size="12" />
      {{ timeAgo }}
    </div>

    <!-- Pulse Modal -->
    <a-modal
      v-model:open="showPulse"
      title="Pulse Duration"
      @ok="executePulse"
      @cancel="showPulse = false"
      ok-text="Pulse"
      :confirm-loading="isExecuting">
      <div class="pulse-control">
        <a-slider
          v-model:value="pulseDuration"
          :min="100"
          :max="5000"
          :step="100" />
        <div class="pulse-value">{{ pulseDuration }}ms</div>
      </div>
      <div class="preset-buttons">
        <a-button size="small" @click="pulseDuration = 100">100ms</a-button>
        <a-button size="small" @click="pulseDuration = 500">500ms</a-button>
        <a-button size="small" @click="pulseDuration = 1000">1s</a-button>
        <a-button size="small" @click="pulseDuration = 2000">2s</a-button>
      </div>
    </a-modal>

    <!-- Blink Modal -->
    <a-modal
      v-model:open="showBlink"
      title="Blink Settings"
      @ok="executeBlink"
      @cancel="showBlink = false"
      ok-text="Blink"
      :confirm-loading="isExecuting">
      <div class="form-group">
        <label>Times:</label>
        <a-input-number v-model:value="blinkTimes" :min="1" :max="10" />
      </div>
      <div class="form-group">
        <label>Interval (ms):</label>
        <a-input-number
          v-model:value="blinkInterval"
          :min="100"
          :max="1000"
          :step="50" />
      </div>
    </a-modal>
  </a-card>
</template>

<script setup>
import { computed, ref } from 'vue'
import {
  Zap,
  Lightbulb,
  MousePointerClick,
  Thermometer,
  Pin,
  ToggleLeft,
  ToggleRight,
  Power,
  PowerOff,
  Eye,
  Activity,
  Clock,
  CircleDot,
  Settings,
} from '@lucide/vue'

const props = defineProps({
  pin: { type: Object, required: true },
  nodeId: { type: String, required: true },
})

const emit = defineEmits(['action'])

const showPulse = ref(false)
const showBlink = ref(false)
const pulseDuration = ref(500)
const blinkTimes = ref(3)
const blinkInterval = ref(200)
const isExecuting = ref(false)

const isOn = computed(() => props.pin.state === true || props.pin.state === 'HIGH')

const stateText = computed(() => {
  if (props.pin.mode === 'input') return isOn.value ? 'HIGH' : 'LOW'
  return isOn.value ? 'ON' : 'OFF'
})

const stateColor = computed(() => {
  if (props.pin.mode === 'input') return isOn.value ? 'blue' : 'default'
  return isOn.value ? 'success' : 'default'
})

const pinIcon = computed(() => {
  const icons = {
    relay: Zap,
    led: Lightbulb,
    button: MousePointerClick,
    dht22: Thermometer,
  }
  return icons[props.pin.type] || Pin
})

const stateIcon = computed(() => {
  if (props.pin.mode === 'input') {
    return isOn.value ? Activity : CircleDot
  }
  return isOn.value ? Power : CircleDot
})

const canToggle = computed(() => props.pin.mode === 'output' && props.pin.type !== 'relay')
const canPulse = computed(() => props.pin.mode === 'output' && props.pin.type === 'relay')
const canSet = computed(() => props.pin.mode === 'output')
const canBlink = computed(() => props.pin.mode === 'output' && props.pin.type === 'led')
const canRead = computed(() => props.pin.mode === 'input')

const timeAgo = computed(() => {
  if (!props.pin.last_update) return ''
  const diff = Math.floor(
    (new Date() - new Date(props.pin.last_update)) / 1000,
  )
  if (diff < 60) return 'Just now'
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
})

async function execute(action, params = {}) {
  isExecuting.value = true
  try {
    emit('action', {
      nodeId: props.nodeId,
      pinId: props.pin.id,
      action,
      params,
    })
  } finally {
    setTimeout(() => {
      isExecuting.value = false
    }, 300)
  }
}

function executePulse() {
  execute('pulse', { duration_ms: pulseDuration.value })
  showPulse.value = false
}

function executePulseDirect() {
  const defaultMs = props.pin.actions?.pulse?.default_ms || 500
  execute('pulse', { duration_ms: defaultMs })
}

function executeBlink() {
  execute('blink', {
    times: blinkTimes.value,
    interval_ms: blinkInterval.value,
  })
  showBlink.value = false
}
</script>

<style scoped>
.pin-card {
  background: var(--light-surface);
  border-color: var(--light-border);
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.pin-card :deep(.ant-card-body) {
  padding: 16px;
}

.pin-card::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: var(--light-border);
  transition: background 0.3s ease;
}

.pin-card.pin-on::before {
  background: var(--success-color);
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.3);
}

.pin-card.pin-active {
  opacity: 0.8;
}

.pin-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}

.pin-icon {
  font-size: 22px;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(13, 148, 136, 0.15);
  border-radius: 8px;
  color: var(--primary-color);
}

.pin-icon.relay {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning-color);
}
.pin-icon.led {
  background: rgba(234, 179, 8, 0.15);
  color: #d97706;
}
.pin-icon.button {
  background: rgba(99, 102, 241, 0.15);
  color: var(--secondary-color);
}
.pin-icon.dht22 {
  background: rgba(239, 68, 68, 0.15);
  color: var(--error-color);
}

.pin-info {
  flex: 1;
}

.pin-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--light-text);
  margin-bottom: 2px;
}

.pin-details {
  font-size: 11px;
  color: var(--light-text-muted);
}

.pin-state-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-size: 11px;
}

.pin-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.btn-toggle.btn-on {
  background: var(--success-color);
  border-color: var(--success-color);
  color: white;
}

.btn-on {
  background: var(--success-color);
  border-color: var(--success-color);
  color: white;
}
.btn-off {
  background: var(--error-color);
  border-color: var(--error-color);
  color: white;
}
.btn-warning {
  background: var(--warning-color);
  border-color: var(--warning-color);
  color: white;
}
.btn-info {
  background: var(--info-color);
  border-color: var(--info-color);
  color: white;
}

.pin-timestamp {
  margin-top: 10px;
  font-size: 11px;
  color: var(--light-text-muted);
  text-align: right;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
}

.pulse-control {
  margin-bottom: 16px;
}

.pulse-value {
  text-align: center;
  font-size: 20px;
  font-weight: 600;
  color: var(--primary-color);
}

.preset-buttons {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.form-group {
  margin-bottom: 12px;
}

.form-group label {
  display: block;
  margin-bottom: 4px;
  color: var(--light-text-muted);
  font-size: 13px;
}
</style>
