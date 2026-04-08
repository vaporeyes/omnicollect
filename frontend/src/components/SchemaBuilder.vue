<script lang="ts" setup>
import {ref, watch, computed} from 'vue'
import * as api from '../api/client'
import SchemaVisualEditor from './SchemaVisualEditor.vue'
import SchemaCodeEditor from './SchemaCodeEditor.vue'
import SchemaFormPreview from './SchemaFormPreview.vue'

const props = defineProps<{
  moduleId?: string | null
  initialJSON?: string | null
}>()

const emit = defineEmits<{
  saved: [schema: any]
  close: []
}>()

interface DraftAttribute {
  name: string
  type: string
  required: boolean
  options: string[]
  display: { label: string; placeholder: string; widget: string }
}

interface DraftSchema {
  id: string
  displayName: string
  description: string
  attributes: DraftAttribute[]
}

function emptySchema(): DraftSchema {
  return {
    id: '',
    displayName: '',
    description: '',
    attributes: [],
  }
}

const draftSchema = ref<DraftSchema>(emptySchema())
const codeContent = ref('')
const parseError = ref<string | null>(null)
const saveError = ref<string | null>(null)
const hasChanges = ref(false)
const saving = ref(false)

// Tracks whether the code editor or visual editor was the last source of change
let syncSource: 'visual' | 'code' | 'init' = 'init'

// Initialize from props
if (props.initialJSON) {
  try {
    draftSchema.value = JSON.parse(props.initialJSON)
    codeContent.value = JSON.stringify(draftSchema.value, null, 2)
  } catch {
    codeContent.value = props.initialJSON
  }
} else {
  codeContent.value = JSON.stringify(emptySchema(), null, 2)
}

// Visual editor changes -> update code editor
function onVisualChange(schema: DraftSchema) {
  syncSource = 'visual'
  draftSchema.value = schema
  codeContent.value = JSON.stringify(schema, null, 2)
  parseError.value = null
  hasChanges.value = true
}

// Code editor changes -> try to update visual editor
let debounceTimer: ReturnType<typeof setTimeout> | null = null

function onCodeChange(text: string) {
  codeContent.value = text
  hasChanges.value = true

  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    try {
      const parsed = JSON.parse(text)
      syncSource = 'code'
      draftSchema.value = parsed
      parseError.value = null
    } catch (e: any) {
      parseError.value = e.message
    }
  }, 300)
}

const isEditMode = computed(() => !!props.moduleId)

function validate(): string | null {
  if (!draftSchema.value.displayName.trim()) {
    return 'Display name is required'
  }
  if (!draftSchema.value.id.trim()) {
    return 'ID is required'
  }
  const names = new Set<string>()
  for (const attr of draftSchema.value.attributes) {
    if (!attr.name.trim()) return 'All fields must have a name'
    if (names.has(attr.name)) return `Duplicate field name: "${attr.name}"`
    names.add(attr.name)
    if (attr.type === 'enum' && (!attr.options || attr.options.length === 0)) {
      return `Enum field "${attr.name}" must have at least one option`
    }
  }
  return null
}

async function onSave() {
  saveError.value = null
  const validationError = validate()
  if (validationError) {
    saveError.value = validationError
    return
  }

  saving.value = true
  try {
    const schema = await api.post('/api/v1/modules', draftSchema.value)
    hasChanges.value = false
    emit('saved', schema)
  } catch (e: any) {
    saveError.value = e?.message ?? String(e)
  } finally {
    saving.value = false
  }
}

const showDiscardConfirm = ref(false)

function onCancel() {
  if (hasChanges.value) {
    showDiscardConfirm.value = true
    return
  }
  emit('close')
}

function confirmDiscard() {
  showDiscardConfirm.value = false
  emit('close')
}

function cancelDiscard() {
  showDiscardConfirm.value = false
}
</script>

<template>
  <div class="schema-builder">
    <div class="builder-toolbar">
      <h3>{{ isEditMode ? 'Edit Schema' : 'New Schema' }}</h3>
      <div class="toolbar-actions">
        <div v-if="saveError" class="save-error">{{ saveError }}</div>
        <button class="btn btn-primary" :disabled="saving" @click="onSave">
          {{ saving ? 'Saving...' : 'Save' }}
        </button>
        <button class="btn btn-secondary" @click="onCancel">Cancel</button>
      </div>
    </div>

    <div class="builder-panes">
      <div class="left-pane">
        <SchemaVisualEditor
          :schema="draftSchema"
          @update:schema="onVisualChange"
        />
        <SchemaFormPreview :schema="draftSchema" />
      </div>
      <div class="right-pane">
        <SchemaCodeEditor
          :modelValue="codeContent"
          :error="parseError"
          @update:modelValue="onCodeChange"
        />
      </div>
    </div>

    <!-- Discard changes confirmation -->
    <Teleport to="body">
      <div v-if="showDiscardConfirm" class="confirm-overlay" @click.self="cancelDiscard">
        <div class="confirm-dialog">
          <p class="confirm-title">Unsaved Changes</p>
          <p class="confirm-message">You have unsaved changes. Discard them?</p>
          <div class="confirm-actions">
            <button class="confirm-cancel-btn" @click="cancelDiscard">Keep Editing</button>
            <button class="confirm-delete-btn" @click="confirmDiscard">Discard</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.confirm-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 4000;
}
.confirm-dialog {
  background: var(--bg-primary, #1e1e2e);
  border: 1px solid var(--border-primary, #333);
  border-radius: var(--radius-md);
  padding: 28px;
  max-width: 380px;
  width: 90%;
  box-shadow: var(--shadow-lg);
}
.confirm-title {
  margin: 0 0 4px;
  font-family: 'Instrument Serif', serif;
  font-size: 18px;
  font-weight: 400;
  color: var(--text-primary);
}
.confirm-message {
  margin: 0 0 24px;
  font-size: 13px;
  color: var(--text-secondary);
}
.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.confirm-cancel-btn {
  padding: 8px 18px;
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-primary);
  cursor: pointer;
  font-size: 13px;
  font-family: 'Outfit', sans-serif;
}
.confirm-cancel-btn:hover {
  background: var(--bg-hover);
}
.confirm-delete-btn {
  padding: 8px 18px;
  border: none;
  border-radius: var(--radius-sm);
  background: var(--error-border, #dc2626);
  color: #fff;
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
  font-family: 'Outfit', sans-serif;
}
.confirm-delete-btn:hover {
  background: #b91c1c;
}
.schema-builder {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.builder-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-primary);
  margin-bottom: 12px;
}
.builder-toolbar h3 {
  margin: 0;
  font-size: 16px;
}
.toolbar-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}
.save-error {
  color: var(--error-text);
  font-size: 13px;
}
.btn {
  padding: 6px 14px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}
.btn-primary { background: var(--accent-blue); color: var(--text-on-accent); }
.btn-primary:hover { background: var(--accent-blue-hover); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { background: var(--btn-secondary-bg); color: var(--text-primary); }
.btn-secondary:hover { background: var(--btn-secondary-bg-hover); }
.builder-panes {
  display: flex;
  flex: 1;
  gap: 12px;
  overflow: hidden;
}
.left-pane {
  flex: 6;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.right-pane {
  flex: 4;
  overflow-y: auto;
}
</style>
