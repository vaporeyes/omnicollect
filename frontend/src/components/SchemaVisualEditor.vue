<script lang="ts" setup>
import {computed} from 'vue'

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

function moveField(index: number, direction: -1 | 1) {
  const newIndex = index + direction
  if (newIndex < 0 || newIndex >= props.schema.attributes.length) return
  const attrs = [...props.schema.attributes]
  const temp = attrs[index]
  attrs[index] = attrs[newIndex]
  attrs[newIndex] = temp
  emit('update:schema', {...props.schema, attributes: attrs})
}

function updateAttribute(index: number, key: string, value: any) {
  const attrs = props.schema.attributes.map((a, i) => {
    if (i !== index) return a
    return {...a, [key]: value}
  })
  emit('update:schema', {...props.schema, attributes: attrs})
}

function updateDisplay(index: number, key: string, value: string) {
  const attrs = props.schema.attributes.map((a, i) => {
    if (i !== index) return a
    return {...a, display: {...a.display, [key]: value}}
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
  <div class="visual-editor">
    <!-- Schema metadata -->
    <div class="meta-section">
      <div class="form-field">
        <label class="field-label">Display Name <span class="required">*</span></label>
        <input
          type="text"
          :value="schema.displayName"
          @input="updateDisplayName(($event.target as HTMLInputElement).value)"
          placeholder="e.g., Vinyl Records"
          class="field-input"
        />
      </div>
      <div class="form-field">
        <label class="field-label">ID</label>
        <input type="text" :value="schema.id" disabled class="field-input id-field" />
      </div>
      <div class="form-field">
        <label class="field-label">Description</label>
        <input
          type="text"
          :value="schema.description"
          @input="updateField('description', ($event.target as HTMLInputElement).value)"
          placeholder="Optional description"
          class="field-input"
        />
      </div>
    </div>

    <!-- Field list -->
    <div class="fields-section">
      <h4>Fields</h4>

      <div v-for="(attr, idx) in schema.attributes" :key="idx" class="field-row">
        <div class="field-main">
          <input
            type="text"
            :value="attr.name"
            @input="updateAttribute(idx, 'name', ($event.target as HTMLInputElement).value)"
            placeholder="Field name"
            class="field-name-input"
          />
          <select
            :value="attr.type"
            @change="updateAttribute(idx, 'type', ($event.target as HTMLSelectElement).value)"
            class="type-select"
          >
            <option v-for="t in validTypes" :key="t" :value="t">{{ t }}</option>
          </select>
          <label class="req-toggle">
            <input
              type="checkbox"
              :checked="attr.required"
              @change="updateAttribute(idx, 'required', ($event.target as HTMLInputElement).checked)"
            />
            Req
          </label>
          <button class="icon-btn" @click="moveField(idx, -1)" :disabled="idx === 0" title="Move up">^</button>
          <button class="icon-btn" @click="moveField(idx, 1)" :disabled="idx === schema.attributes.length - 1" title="Move down">v</button>
          <button class="icon-btn remove-btn" @click="removeField(idx)" title="Remove">x</button>
        </div>

        <!-- Enum options -->
        <div v-if="attr.type === 'enum'" class="enum-options">
          <div v-for="(opt, oi) in attr.options" :key="oi" class="option-row">
            <input
              type="text"
              :value="opt"
              @input="updateOption(idx, oi, ($event.target as HTMLInputElement).value)"
              placeholder="Option value"
              class="option-input"
            />
            <button class="icon-btn remove-btn" @click="removeOption(idx, oi)">x</button>
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
              placeholder="Label override"
              class="hint-input"
            />
            <input
              type="text"
              :value="attr.display?.placeholder || ''"
              @input="updateDisplay(idx, 'placeholder', ($event.target as HTMLInputElement).value)"
              placeholder="Placeholder text"
              class="hint-input"
            />
            <select
              :value="attr.display?.widget || ''"
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
  border-bottom: 1px solid #e2e8f0;
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
.required { color: #e53e3e; }
.field-input {
  width: 100%;
  padding: 5px 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 13px;
  box-sizing: border-box;
}
.id-field {
  background: #f0f0f0;
  color: #666;
}
.fields-section h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
}
.field-row {
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  padding: 8px;
  margin-bottom: 8px;
}
.field-main {
  display: flex;
  gap: 4px;
  align-items: center;
}
.field-name-input {
  flex: 1;
  padding: 4px 6px;
  border: 1px solid #ccc;
  border-radius: 3px;
  font-size: 13px;
}
.type-select {
  padding: 4px;
  border: 1px solid #ccc;
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
.icon-btn {
  width: 24px;
  height: 24px;
  border: 1px solid #ccc;
  border-radius: 3px;
  background: white;
  cursor: pointer;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.icon-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.icon-btn.remove-btn { color: #e53e3e; border-color: #feb2b2; }
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
  border: 1px solid #ccc;
  border-radius: 3px;
  font-size: 12px;
}
.add-option-btn {
  background: none;
  border: none;
  color: #3182ce;
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
  color: #666;
}
.hints-grid {
  display: flex;
  gap: 4px;
  margin-top: 4px;
}
.hint-input {
  flex: 1;
  padding: 3px 6px;
  border: 1px solid #ccc;
  border-radius: 3px;
  font-size: 12px;
}
.add-field-btn {
  padding: 6px 12px;
  border: 1px dashed #ccc;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  font-size: 13px;
  color: #3182ce;
  width: 100%;
}
.add-field-btn:hover { border-color: #3182ce; }
</style>
