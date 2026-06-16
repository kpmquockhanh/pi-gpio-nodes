<template>
  <div class="pin-card" :class="{ 'pin-on': isOn, 'pin-active': isExecuting }">
    <div class="pin-header">
      <div class="pin-icon" :class="pin.type">
        <component :is="pinIcon" :size="20" />
      </div>
      <div class="pin-info">
        <h3 class="pin-name">{{ pin.name }}</h3>
        <span class="pin-details">BCM {{ pin.bcm }} | {{ pin.mode }}</span>
      </div>
      <div class="pin-state-badge" :class="stateClass">
        <component :is="stateIcon" :size="12" class="state-icon" />
        {{ stateText }}
      </div>
    </div>

    <div class="pin-actions">
      <button v-if="canToggle" class="btn btn-toggle" :class="{ 'btn-on': isOn }"
        @click="execute('toggle')" :disabled="isExecuting">
        <component :is="isOn ? ToggleRight : ToggleLeft" :size="14" />
        {{ isOn ? 'ON' : 'OFF' }}
      </button>
      <button v-if="canPulse" class="btn btn-secondary" @click="showPulse = true" :disabled="isExecuting">
        <Zap :size="14" />
        Pulse
      </button>
      <button v-if="canSet" class="btn btn-on" @click="execute('set', { state: true })" :disabled="isExecuting || isOn">
        <Power :size="14" />
        ON
      </button>
      <button v-if="canSet" class="btn btn-off" @click="execute('set', { state: false })" :disabled="isExecuting || !isOn">
        <PowerOff :size="14" />
        OFF
      </button>
      <button v-if="canBlink" class="btn btn-warning" @click="showBlink = true" :disabled="isExecuting">
        <Eye :size="14" />
        Blink
      </button>
      <button v-if="canRead" class="btn btn-info" @click="execute('read')" :disabled="isExecuting">
        <Activity :size="14" />
        Read
      </button>
    </div>

    <div v-if="pin.last_update" class="pin-timestamp">
      <Clock :size="12" />
      {{ timeAgo }}
    </div>

    <!-- Pulse Modal -->
    <div v-if="showPulse" class="modal-overlay" @click.self="showPulse = false">
      <div class="modal-content">
        <h4>
          <Zap :size="18" />
          Pulse Duration
        </h4>
        <div class="pulse-control">
          <input v-model.number="pulseDuration" type="range" min="100" max="5000" step="100" class="pulse-slider" />
          <div class="pulse-value">{{ pulseDuration }}ms</div>
        </div>
        <div class="preset-buttons">
          <button @click="pulseDuration = 100">100ms</button>
          <button @click="pulseDuration = 500">500ms</button>
          <button @click="pulseDuration = 1000">1s</button>
          <button @click="pulseDuration = 2000">2s</button>
        </div>
        <div class="modal-actions">
          <button class="btn btn-primary" @click="executePulse">
            <Zap :size="14" />
            Pulse
          </button>
          <button class="btn btn-secondary" @click="showPulse = false">
            <X :size="14" />
            Cancel
          </button>
        </div>
      </div>
    </div>

    <!-- Blink Modal -->
    <div v-if="showBlink" class="modal-overlay" @click.self="showBlink = false">
      <div class="modal-content">
        <h4>
          <Eye :size="18" />
          Blink Settings
        </h4>
        <div class="form-group">
          <label>Times:</label>
          <input v-model.number="blinkTimes" type="number" min="1" max="10" />
        </div>
        <div class="form-group">
          <label>Interval (ms):</label>
          <input v-model.number="blinkInterval" type="number" min="100" max="1000" step="50" />
        </div>
        <div class="modal-actions">
          <button class="btn btn-primary" @click="executeBlink">
            <Eye :size="14" />
            Blink
          </button>
          <button class="btn btn-secondary" @click="showBlink = false">
            <X :size="14" />
            Cancel
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { 
  Zap, Lightbulb, MousePointerClick, Thermometer, Pin,
  ToggleLeft, ToggleRight, Power, PowerOff, Eye, Activity, Clock, X, CircleDot
} from '@lucide/vue'

export default {
  name: 'PinCard',
  components: {
    Zap, Lightbulb, MousePointerClick, Thermometer, Pin,
    ToggleLeft, ToggleRight, Power, PowerOff, Eye, Activity, Clock, X, CircleDot
  },
  props: {
    pin: { type: Object, required: true },
    nodeId: { type: String, required: true },
  },
  emits: ['action'],
  data() {
    return {
      showPulse: false,
      showBlink: false,
      pulseDuration: 500,
      blinkTimes: 3,
      blinkInterval: 200,
      isExecuting: false,
    }
  },
  computed: {
    isOn() {
      return this.pin.state === true || this.pin.state === 'HIGH'
    },
    stateText() {
      if (this.pin.mode === 'input') return this.isOn ? 'HIGH' : 'LOW'
      return this.isOn ? 'ON' : 'OFF'
    },
    stateClass() {
      if (this.pin.mode === 'input') return this.isOn ? 'input-high' : 'input-low'
      return this.isOn ? 'output-on' : 'output-off'
    },
    pinIcon() {
      const icons = {
        relay: Zap,
        led: Lightbulb,
        button: MousePointerClick,
        dht22: Thermometer,
      }
      return icons[this.pin.type] || Pin
    },
    stateIcon() {
      if (this.pin.mode === 'input') {
        return this.isOn ? Activity : CircleDot
      }
      return this.isOn ? Power : CircleDot
    },
    canToggle() {
      return this.pin.mode === 'output' && this.pin.type !== 'relay'
    },
    canPulse() {
      return this.pin.mode === 'output' && this.pin.type === 'relay'
    },
    canSet() {
      return this.pin.mode === 'output'
    },
    canBlink() {
      return this.pin.mode === 'output'
    },
    canRead() {
      return this.pin.mode === 'input'
    },
    timeAgo() {
      if (!this.pin.last_update) return ''
      const diff = Math.floor((new Date() - new Date(this.pin.last_update)) / 1000)
      if (diff < 60) return 'Just now'
      if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
      if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
      return `${Math.floor(diff / 86400)}d ago`
    },
  },
  methods: {
    async execute(action, params = {}) {
      this.isExecuting = true
      try {
        this.$emit('action', { nodeId: this.nodeId, pinId: this.pin.id, action, params })
      } finally {
        setTimeout(() => { this.isExecuting = false }, 300)
      }
    },
    executePulse() {
      this.execute('pulse', { duration_ms: this.pulseDuration })
      this.showPulse = false
    },
    executeBlink() {
      this.execute('blink', { times: this.blinkTimes, interval_ms: this.blinkInterval })
      this.showBlink = false
    },
  },
}
</script>

<style scoped>
.pin-card {
  background: linear-gradient(145deg, #1e293b, #0f172a);
  border: 1px solid #334155;
  border-radius: 12px;
  padding: 16px;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}
.pin-card::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0; height: 3px;
  background: #475569;
  transition: background 0.3s ease;
}
.pin-card.pin-on::before {
  background: #4ade80;
  box-shadow: 0 0 10px rgba(74, 222, 128, 0.5);
}
.pin-card.pin-active { opacity: 0.8; }
.pin-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}
.pin-icon {
  font-size: 22px;
  width: 36px; height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(59, 130, 246, 0.15);
  border-radius: 8px;
  color: #60a5fa;
}
.pin-icon.relay { background: rgba(245, 158, 11, 0.15); color: #f59e0b; }
.pin-icon.led { background: rgba(234, 179, 8, 0.15); color: #facc15; }
.pin-icon.button { background: rgba(99, 102, 241, 0.15); color: #818cf8; }
.pin-icon.dht22 { background: rgba(239, 68, 68, 0.15); color: #f87171; }
.pin-info { flex: 1; }
.pin-name { font-size: 15px; font-weight: 600; color: #f1f5f9; margin-bottom: 2px; }
.pin-details { font-size: 11px; color: #64748b; }
.pin-state-badge {
  padding: 4px 10px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  display: flex;
  align-items: center;
  gap: 4px;
}
.state-icon {
  flex-shrink: 0;
}
.output-on { background: rgba(34, 197, 94, 0.2); color: #4ade80; }
.output-off { background: rgba(100, 116, 139, 0.2); color: #94a3b8; }
.input-high { background: rgba(59, 130, 246, 0.2); color: #60a5fa; }
.input-low { background: rgba(100, 116, 139, 0.2); color: #64748b; }
.pin-actions { display: flex; flex-wrap: wrap; gap: 6px; }
.btn {
  padding: 7px 14px;
  border: none;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  color: white;
  background: #475569;
  display: flex;
  align-items: center;
  gap: 4px;
}
.btn:hover:not(:disabled) { transform: translateY(-1px); opacity: 0.9; }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-toggle { min-width: 60px; }
.btn-toggle.btn-on { background: #22c55e; }
.btn-on { background: #22c55e; }
.btn-off { background: #ef4444; }
.btn-secondary { background: #64748b; }
.btn-warning { background: #f59e0b; }
.btn-info { background: #06b6d4; }
.btn-primary { background: #3b82f6; }
.pin-timestamp { 
  margin-top: 10px; 
  font-size: 11px; 
  color: #475569; 
  text-align: right; 
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
}
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}
.modal-content {
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 12px;
  padding: 24px;
  min-width: 320px;
  max-width: 90vw;
}
.modal-content h4 { 
  margin-bottom: 16px; 
  color: #f1f5f9; 
  display: flex;
  align-items: center;
  gap: 8px;
}
.pulse-control { margin-bottom: 16px; }
.pulse-slider { width: 100%; margin-bottom: 8px; }
.pulse-value { text-align: center; font-size: 20px; font-weight: 600; color: #3b82f6; }
.preset-buttons { display: flex; gap: 8px; margin-bottom: 16px; }
.preset-buttons button {
  flex: 1;
  padding: 6px;
  background: #334155;
  border: none;
  border-radius: 6px;
  color: #e2e8f0;
  font-size: 12px;
  cursor: pointer;
}
.preset-buttons button:hover { background: #475569; }
.form-group { margin-bottom: 12px; }
.form-group label { display: block; margin-bottom: 4px; color: #94a3b8; font-size: 13px; }
.form-group input {
  width: 100%;
  padding: 8px 12px;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #e2e8f0;
  font-size: 14px;
}
.modal-actions { display: flex; gap: 8px; justify-content: flex-end; margin-top: 16px; }
</style>
