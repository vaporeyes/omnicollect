<script lang="ts" setup>
import {computed} from 'vue'
import FormField from './FormField.vue'
import {main} from '../../wailsjs/go/models'

const props = defineProps<{
  schema: {
    displayName: string
    attributes: Array<{
      name: string
      type: string
      required?: boolean
      options?: string[]
      display?: { label?: string; placeholder?: string; widget?: string }
    }>
  }
}>()

// Convert draft attributes to AttributeSchema format for FormField
const formAttributes = computed(() => {
  return props.schema.attributes.map(attr => {
    return new main.AttributeSchema({
      name: attr.name || '(unnamed)',
      type: attr.type || 'string',
      required: attr.required || false,
      options: attr.options || [],
      display: attr.display ? new main.DisplayHints(attr.display) : undefined,
    })
  })
})
</script>

<template>
  <div class="form-preview">
    <h4>Form Preview</h4>

    <div v-if="schema.attributes.length === 0" class="empty-state">
      Add fields to see preview
    </div>

    <template v-else>
      <!-- Base fields (always shown, disabled) -->
      <div class="form-field">
        <label class="field-label">Title <span class="required">*</span></label>
        <input type="text" disabled placeholder="Item title" class="field-input" />
      </div>
      <div class="form-field">
        <label class="field-label">Purchase Price</label>
        <input type="number" disabled placeholder="0.00" class="field-input" />
      </div>

      <!-- Dynamic fields from schema -->
      <div v-for="attr in formAttributes" :key="attr.name" class="preview-field">
        <FormField
          :attribute="attr"
          :modelValue="null"
          @update:modelValue="() => {}"
        />
      </div>
    </template>
  </div>
</template>

<style scoped>
.form-preview {
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  padding: 12px;
  background: #f7fafc;
}
.form-preview h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #666;
}
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
  color: #e53e3e;
}
.field-input {
  width: 100%;
  padding: 6px 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
  background: #eee;
}
.empty-state {
  font-size: 13px;
  color: #888;
  padding: 24px;
  text-align: center;
}
.preview-field {
  pointer-events: none;
  opacity: 0.8;
}
</style>
