<script lang="ts" setup>
import {ref, computed, onBeforeUnmount} from 'vue'
import {useEditorGuard} from '../composables/editorGuard'
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
const {canLeave} = useEditorGuard(hasChanges, saving)
defineExpose({canLeave})
let alive = true
onBeforeUnmount(() => {alive = false; if (debounceTimer) clearTimeout(debounceTimer)})

// Initialize from props
if (props.initialJSON) {
  try {
    draftSchema.value = parseDraft(props.initialJSON)
    codeContent.value = JSON.stringify(draftSchema.value, null, 2)
  } catch (e: any) {
    parseError.value = e.message
    codeContent.value = props.initialJSON
  }
} else {
  codeContent.value = JSON.stringify(emptySchema(), null, 2)
}

function parseDraft(text: string): DraftSchema {
  const value = JSON.parse(text)
  if (!value || typeof value !== 'object' || Array.isArray(value) ||
      typeof value.id !== 'string' || typeof value.displayName !== 'string' ||
      (value.description !== undefined && typeof value.description !== 'string') ||
      !Array.isArray(value.attributes)) throw new Error('Schema needs an ID, display name and fields array.')
  for (const attr of value.attributes) {
    if (!attr || typeof attr.name !== 'string' || typeof attr.type !== 'string' ||
        (attr.required !== undefined && typeof attr.required !== 'boolean') ||
        (attr.options !== undefined && (!Array.isArray(attr.options) || attr.options.some((v: unknown) => typeof v !== 'string'))) ||
        (attr.display !== undefined && (!attr.display || typeof attr.display !== 'object' || Array.isArray(attr.display)))) {
      throw new Error('Each field needs a name, type and valid options/display settings.')
    }
  }
  return value
}

// Visual editor changes -> update code editor
function onVisualChange(schema: DraftSchema) {
  if (saving.value) return
  if (debounceTimer) clearTimeout(debounceTimer)
  draftSchema.value = schema
  codeContent.value = JSON.stringify(schema, null, 2)
  parseError.value = null
  hasChanges.value = true
}

// Code editor changes -> try to update visual editor
let debounceTimer: ReturnType<typeof setTimeout> | null = null

function onCodeChange(text: string) {
  if (saving.value) return
  codeContent.value = text
  hasChanges.value = true

  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    try {
      const parsed = parseDraft(text)
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
  if (saving.value) return
  saveError.value = null
  if (debounceTimer) clearTimeout(debounceTimer)
  try {draftSchema.value = parseDraft(codeContent.value); parseError.value = null}
  catch (e: any) {parseError.value = e.message; saveError.value = e.message; return}
  const validationError = validate()
  if (validationError) {
    saveError.value = validationError
    return
  }

  saving.value = true
  try {
    const schema = await api.post('/api/v1/modules', JSON.parse(JSON.stringify(draftSchema.value)))
    if (!alive) return
    hasChanges.value = false
    emit('saved', schema)
  } catch (e: any) {
    saveError.value = e?.message ?? String(e)
  } finally {
    saving.value = false
  }
}

function onCancel() {
  if (canLeave()) emit('close')
}

</script>

<template>
  <div class="schema-builder">
    <div class="builder-toolbar">
      <h3>{{ isEditMode ? 'Edit Schema' : 'New Schema' }}</h3>
      <div class="toolbar-actions">
        <div v-if="saveError" role="alert" class="save-error">{{ saveError }}</div>
        <button class="btn btn-primary" :disabled="saving" @click="onSave">
          {{ saving ? 'Saving...' : 'Save' }}
        </button>
        <button class="btn btn-secondary" :disabled="saving" @click="onCancel">Cancel</button>
      </div>
    </div>

    <fieldset class="builder-panes" :disabled="saving" style="border: 0; padding: 0; margin: 0;">
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
          :disabled="saving"
          :error="parseError"
          @update:modelValue="onCodeChange"
        />
      </div>
    </fieldset>
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
