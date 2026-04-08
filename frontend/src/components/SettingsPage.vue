<script lang="ts" setup>
import {ref, reactive, watch, computed} from 'vue'
import * as api from '../api/client'
import {
  applyPolineTheme, getPreviewColors,
  DEFAULT_CONFIG, PRESETS,
  type ThemeConfig, type ThemeAnchor,
} from '../theme'

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

// Determine effective dark mode for preview
const effectiveDark = computed(() => {
  if (config.mode === 'system') return props.systemDark
  return config.mode === 'dark'
})

// Live preview: apply theme as config changes
watch(
  () => JSON.stringify(config),
  () => {
    applyPolineTheme(effectiveDark.value, config as ThemeConfig)
  },
  {immediate: false}
)

// Preview palette swatches
const previewColors = computed(() =>
  getPreviewColors(config as ThemeConfig, effectiveDark.value)
)

function applyPreset(name: string) {
  const preset = PRESETS[name]
  if (!preset) return
  config.lightAnchors = JSON.parse(JSON.stringify(preset.lightAnchors))
  config.darkAnchors = JSON.parse(JSON.stringify(preset.darkAnchors))
  config.numPoints = preset.numPoints
  config.hueShift = preset.hueShift
}

function updateAnchor(
  mode: 'lightAnchors' | 'darkAnchors',
  index: number,
  field: keyof ThemeAnchor,
  value: number
) {
  config[mode][index][field] = value
}

function addAnchor(mode: 'lightAnchors' | 'darkAnchors') {
  config[mode].push({hue: 180, saturation: 0.5, lightness: 0.5})
}

function removeAnchor(mode: 'lightAnchors' | 'darkAnchors', index: number) {
  if (config[mode].length <= 2) return
  config[mode].splice(index, 1)
}

function resetToDefault() {
  Object.assign(config, JSON.parse(JSON.stringify(DEFAULT_CONFIG)))
}

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

const presetNames = Object.keys(PRESETS)
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

    <!-- Appearance Mode -->
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
    </section>

    <!-- Presets -->
    <section class="settings-section">
      <h3>Color Presets</h3>
      <div class="preset-grid">
        <button
          v-for="name in presetNames"
          :key="name"
          class="preset-btn"
          @click="applyPreset(name)"
        >{{ name }}</button>
        <button class="preset-btn preset-reset" @click="resetToDefault">Reset Default</button>
      </div>
    </section>

    <!-- Palette Preview -->
    <section class="settings-section">
      <h3>Generated Palette</h3>
      <div class="swatch-row">
        <div
          v-for="(color, i) in previewColors"
          :key="i"
          class="swatch"
          :style="{background: color}"
          :title="color"
        />
      </div>
    </section>

    <!-- Hue Shift -->
    <section class="settings-section">
      <h3>Hue Shift</h3>
      <p class="section-desc">
        Shift all hues across the palette. Useful for creating variations
        of the same color scheme.
      </p>
      <div class="slider-row">
        <input
          type="range"
          min="-180" max="180" step="1"
          v-model.number="config.hueShift"
          class="slider"
        />
        <span class="slider-value">{{ config.hueShift }}&#176;</span>
      </div>
    </section>

    <!-- Num Points -->
    <section class="settings-section">
      <h3>Palette Size</h3>
      <p class="section-desc">
        Number of color points generated between anchors.
      </p>
      <div class="slider-row">
        <input
          type="range"
          min="6" max="20" step="1"
          v-model.number="config.numPoints"
          class="slider"
        />
        <span class="slider-value">{{ config.numPoints }}</span>
      </div>
    </section>

    <!-- Anchor Editor -->
    <section class="settings-section" v-for="mode in (['lightAnchors', 'darkAnchors'] as const)" :key="mode">
      <h3>{{ mode === 'lightAnchors' ? 'Light Mode' : 'Dark Mode' }} Anchors</h3>
      <div class="anchor-list">
        <div v-for="(anchor, idx) in config[mode]" :key="idx" class="anchor-row">
          <div class="anchor-preview" :style="{
            background: `hsl(${(anchor.hue + config.hueShift + 360) % 360}, ${anchor.saturation * 100}%, ${anchor.lightness * 100}%)`
          }" />
          <div class="anchor-controls">
            <label class="anchor-field">
              <span>H</span>
              <input type="range" min="0" max="360" step="1"
                :value="anchor.hue"
                @input="updateAnchor(mode, idx, 'hue', Number(($event.target as HTMLInputElement).value))"
              />
              <span class="anchor-val">{{ Math.round(anchor.hue) }}</span>
            </label>
            <label class="anchor-field">
              <span>S</span>
              <input type="range" min="0" max="1" step="0.01"
                :value="anchor.saturation"
                @input="updateAnchor(mode, idx, 'saturation', Number(($event.target as HTMLInputElement).value))"
              />
              <span class="anchor-val">{{ Math.round(anchor.saturation * 100) }}%</span>
            </label>
            <label class="anchor-field">
              <span>L</span>
              <input type="range" min="0" max="1" step="0.01"
                :value="anchor.lightness"
                @input="updateAnchor(mode, idx, 'lightness', Number(($event.target as HTMLInputElement).value))"
              />
              <span class="anchor-val">{{ Math.round(anchor.lightness * 100) }}%</span>
            </label>
          </div>
          <button
            v-if="config[mode].length > 2"
            class="remove-anchor"
            @click="removeAnchor(mode, idx)"
          >x</button>
        </div>
        <button class="add-anchor" @click="addAnchor(mode)">+ Add Anchor</button>
      </div>
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
.section-desc {
  font-size: 13px;
  color: var(--text-muted);
  margin: 0 0 10px 0;
}

/* Mode selector */
.mode-selector {
  display: flex;
  gap: 2px;
  background: var(--bg-secondary);
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
  box-shadow: var(--shadow-sm);
}

/* Presets */
.preset-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.preset-btn {
  padding: 7px 14px;
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-md);
  background: var(--bg-primary);
  color: var(--text-primary);
  cursor: pointer;
  font-size: 13px;
  transition: background 0.15s, border-color 0.15s;
}
.preset-btn:hover {
  background: var(--bg-hover);
  border-color: var(--accent-blue);
}
.preset-reset {
  color: var(--text-muted);
  border-style: dashed;
}

/* Swatch preview */
.swatch-row {
  display: flex;
  border-radius: var(--radius-md);
  overflow: hidden;
  height: 36px;
}
.swatch {
  flex: 1;
  transition: background 0.2s;
}

/* Sliders */
.slider-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.slider {
  flex: 1;
  height: 6px;
  -webkit-appearance: none;
  appearance: none;
  border-radius: 3px;
  background: var(--bg-tertiary);
  outline: none;
}
.slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--accent-blue);
  cursor: pointer;
  border: 2px solid var(--bg-primary);
  box-shadow: var(--shadow-sm);
}
.slider-value {
  font-size: 13px;
  font-variant-numeric: tabular-nums;
  color: var(--text-secondary);
  min-width: 42px;
  text-align: right;
}

/* Anchor editor */
.anchor-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.anchor-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
}
.anchor-preview {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  flex-shrink: 0;
  border: 1px solid var(--border-primary);
}
.anchor-controls {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.anchor-field {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-muted);
}
.anchor-field span:first-child {
  width: 12px;
  font-weight: 600;
}
.anchor-field input[type="range"] {
  flex: 1;
  height: 4px;
}
.anchor-val {
  min-width: 36px;
  text-align: right;
  font-variant-numeric: tabular-nums;
  color: var(--text-secondary);
}
.remove-anchor {
  width: 24px;
  height: 24px;
  border: 1px solid var(--error-border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--error-text);
  cursor: pointer;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.add-anchor {
  padding: 6px;
  border: 1px dashed var(--border-input);
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--accent-blue);
  cursor: pointer;
  font-size: 13px;
}
.add-anchor:hover { border-color: var(--accent-blue); }
</style>
