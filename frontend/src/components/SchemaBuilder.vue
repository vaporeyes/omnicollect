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

function onCancel() {
  if (hasChanges.value && !confirm('You have unsaved changes. Discard?')) {
    return
  }
  emit('close')
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
  </div>
</template>

<style scoped>
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
