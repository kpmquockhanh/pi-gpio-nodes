import { 
  Zap, Lightbulb, MousePointerClick, Thermometer, Pin, 
  Plus, Trash2, RotateCw, LayoutDashboard, Workflow, 
  Wifi, WifiOff, ArrowRight, Power, PowerOff, Eye, 
  Clock, Activity, X, Check, ChevronRight, Settings,
  CircleDot, ToggleLeft, ToggleRight, ActivitySquare
} from '@lucide/vue'

// Icon registry mapping
export const iconMap = {
  // Pin types
  relay: Zap,
  led: Lightbulb,
  button: MousePointerClick,
  dht22: Thermometer,
  default: Pin,
  
  // UI actions
  plus: Plus,
  delete: Trash2,
  refresh: RotateCw,
  dashboard: LayoutDashboard,
  automation: Workflow,
  connected: Wifi,
  disconnected: WifiOff,
  polling: Activity,
  arrow: ArrowRight,
  powerOn: Power,
  powerOff: PowerOff,
  eye: Eye,
  clock: Clock,
  activity: ActivitySquare,
  close: X,
  check: Check,
  chevronRight: ChevronRight,
  settings: Settings,
  toggleLeft: ToggleLeft,
  toggleRight: ToggleRight,
  circleDot: CircleDot,
}

export const getIconComponent = (name) => {
  return iconMap[name] || iconMap.default
}

export const getPinIcon = (type) => {
  const mapping = {
    relay: 'relay',
    led: 'led', 
    button: 'button',
    dht22: 'dht22',
  }
  return getIconComponent(mapping[type] || 'default')
}

export default {
  install(app) {
    // Register all icons globally
    Object.entries(iconMap).forEach(([name, component]) => {
      app.component(`Icon${name.charAt(0).toUpperCase() + name.slice(1)}`, component)
    })
  }
}
