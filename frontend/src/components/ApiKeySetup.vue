<template>
  <div class="api-key-setup">
    <a-card class="setup-card" :bordered="true">
      <div class="setup-header">
        <Shield :size="40" class="setup-icon" />
        <h2>API Key Required</h2>
        <p class="setup-desc">
          Enter your master node's API key to connect. This is set in your server's config file.
        </p>
      </div>

      <div class="setup-form">
        <a-input
          v-model:value="apiKeyInput"
          placeholder="Enter API key (e.g., dev-secret-key)"
          size="large"
          @pressEnter="saveApiKey"
        />
        <a-button
          type="primary"
          size="large"
          :loading="testing"
          :disabled="!apiKeyInput.trim()"
          @click="saveApiKey"
        >
          <Lock :size="16" />
          Connect
        </a-button>
      </div>

      <a-alert
        v-if="error"
        type="error"
        :message="error"
        class="setup-error"
        show-icon
      />

      <div class="setup-hint">
        <Info :size="14" />
        <span>
          The API key is stored in your browser's localStorage. You can find it in your server's
          <code>config.toml</code> under <code>[security]</code> → <code>api_key</code>.
        </span>
      </div>
    </a-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Shield, Lock, Info } from '@lucide/vue'
import { api } from '../services/api.js'

const emit = defineEmits(['connected'])

const apiKeyInput = ref('')
const testing = ref(false)
const error = ref('')

async function saveApiKey() {
  const key = apiKeyInput.value.trim()
  if (!key) return

  testing.value = true
  error.value = ''

  try {
    // Temporarily set the key for testing
    api.defaults.headers.common['X-API-Key'] = key

    // Test with a GET request (no auth required, but validates connection)
    await api.get('/api/nodes')

    // If we get here, connection works — save to localStorage
    localStorage.setItem('api_key', key)
    emit('connected')
  } catch (err) {
    if (err.response?.status === 401) {
      error.value = 'Invalid API key. Please check your config file.'
    } else if (err.response?.status === 403) {
      error.value = 'Access denied. Check your API key.'
    } else {
      error.value = `Connection failed: ${err.message}`
    }
    // Clear the failed key
    delete api.defaults.headers.common['X-API-Key']
  } finally {
    testing.value = false
  }
}
</script>

<style scoped>
.api-key-setup {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: var(--light-bg);
}

.setup-card {
  max-width: 480px;
  width: 100%;
  background: var(--light-surface);
  border-color: var(--light-border);
}

.setup-card :deep(.ant-card-body) {
  padding: 40px;
}

.setup-header {
  text-align: center;
  margin-bottom: 32px;
}

.setup-icon {
  color: var(--primary-color);
  margin-bottom: 16px;
}

.setup-header h2 {
  font-size: 24px;
  font-weight: 700;
  color: var(--light-text);
  margin-bottom: 8px;
}

.setup-desc {
  color: var(--light-text-muted);
  font-size: 14px;
  line-height: 1.5;
}

.setup-form {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.setup-form .ant-input {
  flex: 1;
}

.setup-error {
  margin-bottom: 16px;
}

.setup-hint {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px;
  background: var(--light-bg);
  border-radius: var(--radius-md);
  font-size: 13px;
  color: var(--light-text-muted);
  line-height: 1.5;
}

.setup-hint svg {
  flex-shrink: 0;
  margin-top: 2px;
}

.setup-hint code {
  background: rgba(13, 148, 136, 0.1);
  color: var(--primary-color);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 12px;
}

@media (max-width: 480px) {
  .setup-card :deep(.ant-card-body) {
    padding: 24px;
  }

  .setup-form {
    flex-direction: column;
  }
}
</style>
