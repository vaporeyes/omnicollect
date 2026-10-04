<!-- ABOUTME: Multi-step import dialog for restoring backup ZIP files. -->
<!-- ABOUTME: States: file picker, analyzing, summary with mode selection, importing, result. -->
<script lang="ts" setup>
import {ref, computed, onMounted, onBeforeUnmount} from 'vue'
import ModalSurface from './ModalSurface.vue'
import {analyzeBackup, executeImport, APIError} from '../api/client'
import type {ImportSummary, ImportResult} from '../api/types'

defineProps<{refreshError?: string; refreshing?: boolean}>()

const emit = defineEmits<{
  close: []
  imported: [result: ImportResult]
  refresh: []
}>()

type Step = 'pick' | 'analyzing' | 'summary' | 'importing' | 'result'

const step = ref<Step>('pick')
const busy = computed(() => step.value === 'analyzing' || step.value === 'importing')
const replaceConfirmed = ref(false)
let alive = true
function guardUnload(event: BeforeUnloadEvent) {
  if (busy.value) {event.preventDefault(); event.returnValue = ''}
}
onMounted(() => window.addEventListener('beforeunload', guardUnload))
onBeforeUnmount(() => {alive = false; window.removeEventListener('beforeunload', guardUnload)})
defineExpose({requestClose: onClose, busy})
const summary = ref<ImportSummary | null>(null)
const result = ref<ImportResult | null>(null)
const mode = ref<'replace' | 'merge'>('merge')
const error = ref('')
const fileInput = ref<HTMLInputElement | null>(null)
const dragOver = ref(false)

function onFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  if (input.files?.length) {
    startAnalyze(input.files[0])
  }
}

function onDrop(event: DragEvent) {
  dragOver.value = false
  const file = event.dataTransfer?.files?.[0]
  if (file && file.name.toLowerCase().endsWith('.zip')) {
    startAnalyze(file)
  } else {
    error.value = 'Please drop a .zip backup file'
  }
}

async function startAnalyze(file: File) {
  if (busy.value) return
  if (!file.name.toLowerCase().endsWith('.zip') || file.size > 256 * 1024 * 1024 || file.size === 0) {
    error.value = 'Choose a non-empty ZIP backup no larger than 256 MB'
    return
  }
  replaceConfirmed.value = false
  step.value = 'analyzing'
  error.value = ''
  try {
    const analyzed = await analyzeBackup(file)
    if (!alive) return
    summary.value = analyzed
    step.value = 'summary'
  } catch (e: any) {
    if (!alive) return
    error.value = e?.message ?? 'Failed to analyze backup'
    step.value = 'pick'
  }
}

async function onConfirm() {
  if (step.value !== 'summary' || !summary.value || (mode.value === 'replace' && !replaceConfirmed.value)) return
  step.value = 'importing'
  error.value = ''
  try {
    const imported = await executeImport(summary.value.tempId, mode.value)
    if (!alive) return
    result.value = imported
    step.value = 'result'
    emit('imported', result.value)
  } catch (e: any) {
    if (!alive) return
    error.value = e?.message ?? 'Import failed'
    if (!(e instanceof APIError)) error.value += ' The connection was lost; the import may have completed. Refresh your collection before retrying.'
    step.value = 'summary'
  }
}

function onClose() {
  if (!busy.value) emit('close')
}
</script>

<template>
  <Teleport to="body">
    <ModalSurface label="Import backup" :busy="busy" class="import-overlay" @close="onClose" @click.self="onClose">
      <div class="import-dialog">
        <h2 class="import-title">Import Backup</h2>

        <!-- Step 1: File picker -->
        <div v-if="step === 'pick'" class="import-step">
          <div
            class="drop-zone"
            :class="{active: dragOver}"
            @dragover.prevent="dragOver = true"
            @dragleave="dragOver = false"
            @drop.prevent="onDrop"
            role="button"
            tabindex="0"
            @keydown.enter.prevent="fileInput?.click()"
            @keydown.space.prevent="fileInput?.click()"
            @click="fileInput?.click()"
          >
            <p class="drop-label">Drop a backup ZIP here or click to browse</p>
            <p class="drop-hint">.zip files exported from OmniCollect</p>
          </div>
          <input
            ref="fileInput"
            type="file"
            accept=".zip"
            style="display: none"
            @change="onFileSelect"
          />
          <p v-if="error" class="import-error">{{ error }}</p>
          <div class="import-actions">
            <button class="import-cancel-btn" @click="onClose">Cancel</button>
          </div>
        </div>

        <!-- Step 2: Analyzing -->
        <div v-if="step === 'analyzing'" class="import-step">
          <div class="import-spinner-wrap">
            <div class="import-spinner"></div>
            <p class="import-spinner-label">Analyzing backup...</p>
          </div>
        </div>

        <!-- Step 3: Summary + mode selection -->
        <div v-if="step === 'summary'" class="import-step">
          <div class="summary-grid">
            <div class="summary-item">
              <span class="summary-value">{{ summary?.format === 'local' ? 'Local (SQLite)' : 'Portable (JSON)' }}</span>
              <span class="summary-label">Format</span>
            </div>
            <div class="summary-item">
              <span class="summary-value">{{ summary?.itemCount }}</span>
              <span class="summary-label">Items</span>
            </div>
            <div class="summary-item">
              <span class="summary-value">{{ summary?.imageCount }}</span>
              <span class="summary-label">Images</span>
            </div>
            <div class="summary-item">
              <span class="summary-value">{{ summary?.moduleCount }}</span>
              <span class="summary-label">Modules</span>
            </div>
          </div>

          <div v-if="summary?.warnings?.length" class="import-warnings">
            <p v-for="w in summary.warnings" :key="w" class="import-warning">{{ w }}</p>
          </div>

          <label v-if="mode === 'replace'"><input type="checkbox" v-model="replaceConfirmed" /> I understand this replaces the current collection and settings.</label>
          <button type="button" class="import-cancel-btn" @click="step = 'pick'; error = ''">Choose another backup</button>
          <div class="mode-selection">
            <label class="mode-option" :class="{selected: mode === 'merge'}">
              <input type="radio" v-model="mode" value="merge" />
              <span class="mode-name">Merge</span>
              <span class="mode-desc">Keep items absent from the backup; matching IDs are updated from the backup.</span>
            </label>
            <label class="mode-option" :class="{selected: mode === 'replace'}">
              <input type="radio" v-model="mode" value="replace" />
              <span class="mode-name">Replace</span>
              <span class="mode-desc">Replace items, schemas, and settings, and permanently clear deletion recovery. Public showcases will be disabled.</span>
            </label>
          </div>

          <p v-if="error" class="import-error">{{ error }}</p>
          <div class="import-actions">
            <button class="import-cancel-btn" @click="onClose">Cancel</button>
            <button class="import-confirm-btn" :disabled="mode === 'replace' && !replaceConfirmed" @click="onConfirm">Import</button>
          </div>
        </div>

        <!-- Step 4: Importing -->
        <div v-if="step === 'importing'" class="import-step">
          <div class="import-spinner-wrap">
            <div class="import-spinner"></div>
            <p class="import-spinner-label">Importing...</p>
          </div>
        </div>

        <!-- Step 5: Result -->
        <div v-if="step === 'result'" class="import-step">
          <p v-if="refreshError" role="alert" class="import-error">{{ refreshError }}</p>
          <button v-if="refreshError" :disabled="refreshing" class="import-cancel-btn" @click="emit('refresh')">Retry view refresh</button>
          <div class="summary-grid">
            <div class="summary-item">
              <span class="summary-value">{{ result?.itemsImported }}</span>
              <span class="summary-label">Items imported</span>
            </div>
            <div class="summary-item">
              <span class="summary-value">{{ result?.imagesRestored }}</span>
              <span class="summary-label">Images restored</span>
            </div>
            <div class="summary-item">
              <span class="summary-value">{{ result?.modulesImported }}</span>
              <span class="summary-label">Modules imported</span>
            </div>
          </div>

          <div v-if="result?.warnings?.length" class="import-warnings">
            <p v-for="w in result.warnings" :key="w" class="import-warning">{{ w }}</p>
          </div>

          <div class="import-actions">
            <button class="import-confirm-btn" @click="onClose">Done</button>
          </div>
        </div>
      </div>
    </ModalSurface>
  </Teleport>
</template>

<style scoped>
.import-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 4000;
}
.import-dialog {
  max-height: calc(100dvh - 32px);
  overflow: auto;
  overflow-wrap: anywhere;
  background: var(--bg-primary, #1e1e2e);
  border: 1px solid var(--border-primary, #333);
  border-radius: var(--radius-md);
  padding: 28px;
  max-width: 440px;
  width: 90%;
  box-shadow: var(--shadow-lg);
}
.import-title {
  margin: 0 0 20px;
  font-family: var(--font-heading);
  font-size: 22px;
  font-weight: 400;
  color: var(--text-primary);
}
.import-step {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.drop-zone {
  border: 2px dashed var(--border-primary, #444);
  border-radius: var(--radius-md);
  padding: 36px 20px;
  text-align: center;
  cursor: pointer;
  transition: border-color var(--transition-fast), background var(--transition-fast);
}
.drop-zone:hover,
.drop-zone.active {
  border-color: var(--accent-blue);
  background: var(--accent-blue-light, rgba(59, 130, 246, 0.06));
}
.drop-label {
  margin: 0 0 4px;
  font-family: var(--font-body);
  font-size: 14px;
  color: var(--text-primary);
}
.drop-hint {
  margin: 0;
  font-size: 12px;
  color: var(--text-muted);
}
.import-spinner-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 24px 0;
}
.import-spinner {
  width: 28px;
  height: 28px;
  border: 3px solid var(--border-primary, #444);
  border-top-color: var(--accent-blue);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
.import-spinner-label {
  margin: 0;
  font-size: 14px;
  color: var(--text-secondary);
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}
.summary-item {
  display: flex;
  flex-direction: column;
  padding: 12px;
  background: var(--bg-secondary, #262637);
  border-radius: var(--radius-sm);
}
.summary-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}
.summary-label {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
}
.import-warnings {
  padding: 8px 12px;
  background: var(--error-bg, rgba(220, 38, 38, 0.08));
  border-radius: var(--radius-sm);
  border-left: 3px solid var(--error-border, #dc2626);
}
.import-warning {
  margin: 0;
  font-size: 12px;
  color: var(--error-text, #fca5a5);
  line-height: 1.5;
}
.mode-selection {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.mode-option {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 8px;
  padding: 12px;
  border: 1px solid var(--border-primary, #444);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: border-color var(--transition-fast), background var(--transition-fast);
}
.mode-option:hover {
  background: var(--bg-hover);
}
.mode-option.selected {
  border-color: var(--accent-blue);
  background: var(--accent-blue-light, rgba(59, 130, 246, 0.06));
}
.mode-option input[type="radio"] {
  accent-color: var(--accent-blue);
}
.mode-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}
.mode-desc {
  flex-basis: 100%;
  font-size: 12px;
  color: var(--text-muted);
  margin-left: 24px;
}
.import-error {
  margin: 0;
  font-size: 13px;
  color: var(--error-text, #fca5a5);
}
.import-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 4px;
}
.import-cancel-btn {
  padding: 8px 18px;
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-primary);
  cursor: pointer;
  font-size: 13px;
}
.import-cancel-btn:hover {
  background: var(--bg-hover);
}
.import-confirm-btn {
  padding: 8px 18px;
  border: none;
  border-radius: var(--radius-sm);
  background: var(--accent-blue);
  color: var(--text-on-accent, #fff);
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
}
.import-confirm-btn:hover {
  filter: brightness(1.1);
}
</style>
