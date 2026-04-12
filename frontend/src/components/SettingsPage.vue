<!-- ABOUTME: Settings page for theme mode selection (light/dark/system). -->
<!-- ABOUTME: Persists choice to backend settings endpoint. -->
<script lang="ts" setup>
import {ref, reactive, watch, computed} from 'vue'
import * as api from '../api/client'
import {applyTheme, DEFAULT_CONFIG, type ThemeConfig} from '../theme'

const props = defineProps<{
  initialConfig: ThemeConfig
  systemDark: boolean
}>()

const emit = defineEmits<{
  close: []
  saved: [config: ThemeConfig]
}>()

const config = reactive<ThemeConfig>(JSON.parse(JSON.stringify(props.initialConfig)))
const saving = ref(false)
const saveMessage = ref<string | null>(null)

const effectiveDark = computed(() => {
  if (config.mode === 'system') return props.systemDark
  return config.mode === 'dark'
})

// Live preview: apply theme as mode changes
watch(
  () => config.mode,
  () => applyTheme(effectiveDark.value),
)

async function onSave() {
  saving.value = true
  saveMessage.value = null
  try {
    await api.put('/api/v1/settings', {theme: config})
    emit('saved', JSON.parse(JSON.stringify(config)))
    saveMessage.value = 'Settings saved'
    setTimeout(() => { saveMessage.value = null }, 3000)
  } catch (e: any) {
    saveMessage.value = `Error: ${e?.message ?? e}`
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="settings-page">
    <div class="settings-header">
      <button class="back-btn" @click="emit('close')">&larr;</button>
      <h2>Settings</h2>
      <div class="header-actions">
        <span v-if="saveMessage" class="save-msg" :class="{error: saveMessage.startsWith('Error')}">
          {{ saveMessage }}
        </span>
        <button class="btn-primary" :disabled="saving" @click="onSave">
          {{ saving ? 'Saving...' : 'Save' }}
        </button>
      </div>
    </div>

    <section class="settings-section">
      <h3>Appearance</h3>
      <div class="mode-selector">
        <button
          v-for="m in (['light', 'system', 'dark'] as const)"
          :key="m"
          :class="['mode-btn', {active: config.mode === m}]"
          @click="config.mode = m"
        >
          {{ m === 'system' ? 'System' : m.charAt(0).toUpperCase() + m.slice(1) }}
        </button>
      </div>
      <p class="mode-desc">
        {{ config.mode === 'light' ? 'The Archive — unbleached paper, heavy ink, Klein Blue.' :
           config.mode === 'dark' ? 'The Vault — deep obsidian, matte charcoal, Cinnabar.' :
           'Follows your system preference.' }}
      </p>
    </section>
  </div>
</template>

<style scoped>
.settings-page {
  max-width: 640px;
}
.settings-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
}
.settings-header h2 {
  flex: 1;
  margin: 0;
  font-size: 20px;
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}
.save-msg {
  font-size: 13px;
  color: var(--success-text);
}
.save-msg.error {
  color: var(--error-text);
}
.back-btn {
  background: none;
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-md);
  padding: 6px 10px;
  cursor: pointer;
  font-size: 16px;
  color: var(--text-primary);
}
.back-btn:hover { background: var(--bg-hover); }
.btn-primary {
  padding: 8px 16px;
  border: none;
  border-radius: var(--radius-md);
  background: var(--accent-blue);
  color: var(--text-on-accent);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
}
.btn-primary:hover { background: var(--accent-blue-hover); }
.btn-primary:disabled { opacity: 0.5; }

.settings-section {
  margin-bottom: 28px;
}
.settings-section h3 {
  font-size: 13px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  margin: 0 0 10px 0;
}

.mode-selector {
  display: flex;
  gap: 2px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-md);
  padding: 3px;
  width: fit-content;
}
.mode-btn {
  padding: 7px 18px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  border-radius: var(--radius-sm);
  transition: background 0.15s, color 0.15s;
}
.mode-btn:hover { color: var(--text-primary); }
.mode-btn.active {
  background: var(--accent-blue);
  color: var(--text-on-accent);
}
.mode-desc {
  margin: 10px 0 0;
  font-size: 13px;
  color: var(--text-muted);
  font-style: italic;
}
</style>
