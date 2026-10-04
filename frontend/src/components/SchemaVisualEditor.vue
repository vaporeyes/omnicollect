<script lang="ts" setup>
import {computed, toRaw, ref, nextTick} from 'vue'
import draggable from 'vuedraggable'

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

const props = defineProps<{
  schema: DraftSchema
}>()

const emit = defineEmits<{
  'update:schema': [schema: DraftSchema]
}>()

const editorElement = ref<HTMLElement | null>(null)
const attributeKeys = new WeakMap<object, number>()
let nextAttributeKey = 0
function fieldKey(attribute: DraftAttribute): number {
  const raw = toRaw(attribute)
  if (!attributeKeys.has(raw)) attributeKeys.set(raw, ++nextAttributeKey)
  return attributeKeys.get(raw)!
}

const validTypes = ['string', 'number', 'boolean', 'date', 'enum']

// Auto-generate slug from displayName
const slugId = computed(() => {
  return props.schema.displayName
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
})

function updateField(key: keyof DraftSchema, value: any) {
  const updated = {...props.schema, [key]: value}
  if (key === 'displayName') {
    updated.id = slugId.value
  }
  emit('update:schema', updated)
}

function updateDisplayName(value: string) {
  const updated = {...props.schema, displayName: value}
  // Auto-generate slug
  updated.id = value.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')
  emit('update:schema', updated)
}

function addField() {
  const attrs = [...props.schema.attributes, {
    name: '',
    type: 'string',
    required: false,
    options: [],
    display: {label: '', placeholder: '', widget: ''},
  }]
  emit('update:schema', {...props.schema, attributes: attrs})
}

function removeField(index: number) {
  const attrs = [...props.schema.attributes]
  attrs.splice(index, 1)
  emit('update:schema', {...props.schema, attributes: attrs})
}

function reorderFields(attributes: DraftAttribute[]) {
  emit('update:schema', {...props.schema, attributes: [...attributes]})
}
async function moveField(index: number, delta: number) {
  const target = index + delta
  if (target < 0 || target >= props.schema.attributes.length) return
  const attrs = [...props.schema.attributes]
  const [moved] = attrs.splice(index, 1)
  attrs.splice(target, 0, moved)
  reorderFields(attrs)
  await nextTick()
  editorElement.value?.querySelector<HTMLInputElement>(`input[aria-label="Field ${target + 1} name"]`)?.focus()
}

function updateAttribute(index: number, key: string, value: any) {
  const attrs = props.schema.attributes.map((a, i) => {
    if (i !== index) return a
    const updated = {...a, [key]: value}
    attributeKeys.set(updated, fieldKey(a))
    return updated
  })
  emit('update:schema', {...props.schema, attributes: attrs})
}

function updateDisplay(index: number, key: string, value: string) {
  const attrs = props.schema.attributes.map((a, i) => {
    if (i !== index) return a
    const updated = {...a, display: {...a.display, [key]: value}}
    attributeKeys.set(updated, fieldKey(a))
    return updated
  })
  emit('update:schema', {...props.schema, attributes: attrs})
}

function addOption(index: number) {
  const attr = props.schema.attributes[index]
  updateAttribute(index, 'options', [...attr.options, ''])
}

function removeOption(attrIndex: number, optIndex: number) {
  const attr = props.schema.attributes[attrIndex]
  const opts = [...attr.options]
  opts.splice(optIndex, 1)
  updateAttribute(attrIndex, 'options', opts)
}

function updateOption(attrIndex: number, optIndex: number, value: string) {
  const attr = props.schema.attributes[attrIndex]
  const opts = [...attr.options]
  opts[optIndex] = value
  updateAttribute(attrIndex, 'options', opts)
}
</script>

<template>
  <div ref="editorElement" class="visual-editor">
    <!-- Schema metadata -->
    <div class="meta-section">
      <div class="form-field">
        <label class="field-label">Display Name <span class="required">*</span></label>
        <input
          type="text"
          :value="schema.displayName"
          @input="updateDisplayName(($event.target as HTMLInputElement).value)"
          aria-label="Schema display name" placeholder="e.g., Vinyl Records"
          class="field-input"
        />
      </div>
      <div class="form-field">
        <label class="field-label">ID</label>
        <input type="text" aria-label="Schema ID" :value="schema.id" disabled class="field-input id-field" />
      </div>
      <div class="form-field">
        <label class="field-label">Description</label>
        <input
          type="text"
          :value="schema.description"
          @input="updateField('description', ($event.target as HTMLInputElement).value)"
          aria-label="Schema description" placeholder="Optional description"
          class="field-input"
        />
      </div>
    </div>

    <!-- Field list -->
    <div class="fields-section">
      <h4>Fields</h4>

      <draggable
        :model-value="schema.attributes"
        :item-key="fieldKey"
        handle=".drag-handle"
        ghost-class="drag-ghost"
        @update:model-value="reorderFields"
      >
        <template #item="{element: attr, index: idx}">
        <div class="field-row">
        <div class="field-main">
          <span class="drag-handle" title="Drag to reorder">&#9776;</span>
          <input
            type="text"
            :value="attr.name"
            @input="updateAttribute(idx, 'name', ($event.target as HTMLInputElement).value)"
            :aria-label="`Field ${idx + 1} name`" placeholder="Field name"
            class="field-name-input"
          />
          <select
            :value="attr.type"
            @change="updateAttribute(idx, 'type', ($event.target as HTMLSelectElement).value)"
            :aria-label="`Field ${idx + 1} type`" class="type-select"
          >
            <option v-for="t in validTypes" :key="t" :value="t">{{ t }}</option>
          </select>
          <label class="req-toggle">
            <input
              type="checkbox"
              :aria-label="`Require field ${idx + 1}`" :checked="attr.required"
              @change="updateAttribute(idx, 'required', ($event.target as HTMLInputElement).checked)"
            />
            Req
          </label>
          <button class="icon-btn" :aria-label="`Move field ${idx + 1} up`" :disabled="idx === 0" @click="moveField(idx, -1)">↑</button>
          <button class="icon-btn" :aria-label="`Move field ${idx + 1} down`" :disabled="idx === schema.attributes.length - 1" @click="moveField(idx, 1)">↓</button>
          <button class="icon-btn remove-btn" @click="removeField(idx)" :aria-label="`Remove field ${idx + 1}`">x</button>
        </div>

        <!-- Enum options -->
        <div v-if="attr.type === 'enum'" class="enum-options">
          <div v-for="(opt, oi) in (attr.options as string[])" :key="oi" class="option-row">
            <input
              type="text"
              :value="opt"
              @input="updateOption(idx, oi, ($event.target as HTMLInputElement).value)"
              :aria-label="`Option ${oi + 1} for field ${idx + 1}`" placeholder="Option value"
              class="option-input"
            />
            <button class="icon-btn remove-btn" :aria-label="`Remove option ${oi + 1} from field ${idx + 1}`" @click="removeOption(idx, oi)">x</button>
          </div>
          <button class="add-option-btn" @click="addOption(idx)">+ Add option</button>
        </div>

        <!-- Display hints (collapsible) -->
        <details class="display-hints">
          <summary>Display hints</summary>
          <div class="hints-grid">
            <input
              type="text"
              :value="attr.display?.label || ''"
              @input="updateDisplay(idx, 'label', ($event.target as HTMLInputElement).value)"
              :aria-label="`Display label for field ${idx + 1}`" placeholder="Label override"
              class="hint-input"
            />
            <input
              type="text"
              :value="attr.display?.placeholder || ''"
              @input="updateDisplay(idx, 'placeholder', ($event.target as HTMLInputElement).value)"
              :aria-label="`Placeholder for field ${idx + 1}`" placeholder="Placeholder text"
              class="hint-input"
            />
            <select
              :aria-label="`Widget for field ${idx + 1}`" :value="attr.display?.widget || ''"
              @change="updateDisplay(idx, 'widget', ($event.target as HTMLSelectElement).value)"
              class="hint-input"
            >
              <option value="">Default widget</option>
              <option value="text">Text</option>
              <option value="textarea">Textarea</option>
              <option value="dropdown">Dropdown</option>
            </select>
          </div>
        </details>
      </div>
        </template>
      </draggable>

      <button class="add-field-btn" @click="addField">+ Add Field</button>
    </div>
  </div>
</template>

<style scoped>
.visual-editor {
  overflow-y: auto;
}
.meta-section {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-primary);
}
.form-field {
  margin-bottom: 8px;
}
.field-label {
  display: block;
  font-weight: 600;
  margin-bottom: 2px;
  font-size: 13px;
}
.required { color: var(--required-color); }
.field-input {
  width: 100%;
  padding: 5px 8px;
  border: 1px solid var(--border-input);
  border-radius: 4px;
  font-size: 13px;
  box-sizing: border-box;
}
.id-field {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}
.fields-section h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
}
.field-row {
  border: 1px solid var(--border-primary);
  border-radius: 4px;
  padding: 8px;
  margin-bottom: 8px;
}
.field-main {
  flex-wrap: wrap;
  display: flex;
  gap: 4px;
  align-items: center;
}
.field-name-input {
  flex: 1;
  padding: 4px 6px;
  border: 1px solid var(--border-input);
  border-radius: 3px;
  font-size: 13px;
}
.type-select {
  padding: 4px;
  border: 1px solid var(--border-input);
  border-radius: 3px;
  font-size: 13px;
}
.req-toggle {
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 2px;
  white-space: nowrap;
}
.drag-handle {
  cursor: grab;
  color: var(--text-muted);
  font-size: 16px;
  padding: 0 2px;
  user-select: none;
}
.drag-handle:active {
  cursor: grabbing;
}
.drag-ghost {
  opacity: 0.4;
  background: var(--accent-blue-light);
  border-color: var(--accent-blue);
}
.icon-btn {
  width: 24px;
  height: 24px;
  border: 1px solid var(--border-input);
  border-radius: 3px;
  background: var(--bg-primary);
  cursor: pointer;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.icon-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.icon-btn.remove-btn { color: var(--error-text); border-color: var(--error-border); }
.enum-options {
  margin-top: 6px;
  padding-left: 12px;
}
.option-row {
  display: flex;
  gap: 4px;
  margin-bottom: 4px;
}
.option-input {
  flex: 1;
  padding: 3px 6px;
  border: 1px solid var(--border-input);
  border-radius: 3px;
  font-size: 12px;
}
.add-option-btn {
  background: none;
  border: none;
  color: var(--accent-blue);
  cursor: pointer;
  font-size: 12px;
  padding: 2px 0;
}
.display-hints {
  margin-top: 6px;
  font-size: 12px;
}
.display-hints summary {
  cursor: pointer;
  color: var(--text-secondary);
}
.hints-grid {
  display: flex;
  gap: 4px;
  margin-top: 4px;
}
.hint-input {
  flex: 1;
  padding: 3px 6px;
  border: 1px solid var(--border-input);
  border-radius: 3px;
  font-size: 12px;
}
.add-field-btn {
  padding: 6px 12px;
  border: 1px dashed var(--border-input);
  border-radius: 4px;
  background: var(--bg-primary);
  cursor: pointer;
  font-size: 13px;
  color: var(--accent-blue);
  width: 100%;
}
.add-field-btn:hover { border-color: var(--accent-blue); }
</style>
