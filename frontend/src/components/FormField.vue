<script lang="ts" setup>
import {computed} from 'vue'
import {main} from '../../wailsjs/go/models'
import MarkdownEditor from './MarkdownEditor.vue'

const props = defineProps<{
  attribute: main.AttributeSchema
  modelValue: any
  errorMessage?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: any]
}>()

const label = computed(() => props.attribute.display?.label || props.attribute.name)
const placeholder = computed(() => props.attribute.display?.placeholder || '')
const widget = computed(() => props.attribute.display?.widget || '')

// Determine which input to render based on type and widget override
const inputType = computed(() => {
  if (widget.value === 'textarea') return 'textarea'
  switch (props.attribute.type) {
    case 'string': return 'text'
    case 'number': return 'number'
    case 'boolean': return 'checkbox'
    case 'date': return 'date'
    case 'enum': return 'select'
    default:
      console.warn(`Unrecognized attribute type "${props.attribute.type}" for "${props.attribute.name}", falling back to text input`)
      return 'text'
  }
})

function onInput(event: Event) {
  const target = event.target as HTMLInputElement
  if (inputType.value === 'checkbox') {
    emit('update:modelValue', target.checked)
  } else if (inputType.value === 'number') {
    emit('update:modelValue', target.value === '' ? null : Number(target.value))
  } else {
    emit('update:modelValue', target.value)
  }
}
</script>

<template>
  <div class="form-field" :class="{'has-error': errorMessage}">
    <label v-if="inputType !== 'checkbox'" class="field-label">
      {{ label }}
      <span v-if="attribute.required" class="required">*</span>
    </label>

    <MarkdownEditor
      v-if="inputType === 'textarea'"
      :modelValue="modelValue ?? ''"
      @update:modelValue="val => emit('update:modelValue', val)"
    />

    <select
      v-else-if="inputType === 'select'"
      :value="modelValue ?? ''"
      @change="onInput"
      class="field-input"
    >
      <option value="" disabled>Select...</option>
      <option v-for="opt in attribute.options" :key="opt" :value="opt">{{ opt }}</option>
    </select>

    <label v-else-if="inputType === 'checkbox'" class="checkbox-label">
      <input
        type="checkbox"
        :checked="!!modelValue"
        @change="onInput"
      />
      {{ label }}
      <span v-if="attribute.required" class="required">*</span>
    </label>

    <input
      v-else
      :type="inputType"
      :value="modelValue ?? ''"
      :placeholder="placeholder"
      @input="onInput"
      class="field-input"
    />

    <div v-if="errorMessage" class="field-error">{{ errorMessage }}</div>
  </div>
</template>

<style scoped>
.form-field {
  margin-bottom: 12px;
}
.field-label {
  display: block;
  font-weight: 600;
  margin-bottom: 4px;
  font-size: 14px;
}
.required {
  color: var(--required-color);
}
.field-input {
  width: 100%;
  padding: 6px 8px;
  border: 1px solid var(--border-input);
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
}
.has-error .field-input {
  border-color: var(--error-text);
}
.field-error {
  color: var(--error-text);
  font-size: 12px;
  margin-top: 2px;
}
.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
}
textarea.field-input {
  min-height: 80px;
  resize: vertical;
}
</style>
