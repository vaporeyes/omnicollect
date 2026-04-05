<script lang="ts" setup>
import {reactive, watch, computed} from 'vue'
import {main} from '../../wailsjs/go/models'
import FormField from './FormField.vue'

const props = defineProps<{
  schema: main.ModuleSchema
  item?: main.Item | null
}>()

const emit = defineEmits<{
  save: [item: main.Item]
  cancel: []
}>()

const baseFields = reactive({
  title: '',
  purchasePrice: null as number | null,
})

const attributes = reactive<Record<string, any>>({})
const validationErrors = reactive<Record<string, string>>({})

const isEditing = computed(() => !!props.item?.id)

// Initialize form when item or schema changes
watch(() => [props.item, props.schema], () => {
  // Reset base fields
  baseFields.title = props.item?.title ?? ''
  baseFields.purchasePrice = props.item?.purchasePrice ?? null

  // Reset attributes from item or defaults
  const itemAttrs = props.item?.attributes ?? {}
  for (const attr of props.schema.attributes) {
    if (attr.name in itemAttrs) {
      attributes[attr.name] = itemAttrs[attr.name]
    } else {
      // Set default based on type
      switch (attr.type) {
        case 'boolean': attributes[attr.name] = false; break
        case 'number': attributes[attr.name] = null; break
        default: attributes[attr.name] = ''
      }
    }
  }

  // Clear errors
  Object.keys(validationErrors).forEach(k => delete validationErrors[k])
}, {immediate: true})

function validate(): boolean {
  Object.keys(validationErrors).forEach(k => delete validationErrors[k])
  let valid = true

  if (!baseFields.title.trim()) {
    validationErrors['title'] = 'Title is required'
    valid = false
  }

  for (const attr of props.schema.attributes) {
    if (!attr.required) continue
    const val = attributes[attr.name]
    if (val === null || val === undefined || (typeof val === 'string' && !val.trim())) {
      validationErrors[attr.name] = `${attr.display?.label || attr.name} is required`
      valid = false
    }
  }

  return valid
}

function onSubmit() {
  if (!validate()) return

  const item = new main.Item({
    id: props.item?.id ?? '',
    moduleId: props.schema.id,
    title: baseFields.title.trim(),
    purchasePrice: baseFields.purchasePrice,
    images: props.item?.images ?? [],
    attributes: {...attributes},
    createdAt: props.item?.createdAt ?? '',
    updatedAt: '',
  })

  emit('save', item)
}
</script>

<template>
  <div class="dynamic-form">
    <h3>{{ isEditing ? 'Edit' : 'New' }} {{ schema.displayName }}</h3>

    <form @submit.prevent="onSubmit">
      <!-- Base fields -->
      <div class="form-field" :class="{'has-error': validationErrors['title']}">
        <label class="field-label">Title <span class="required">*</span></label>
        <input
          type="text"
          v-model="baseFields.title"
          placeholder="Item title"
          class="field-input"
        />
        <div v-if="validationErrors['title']" class="field-error">
          {{ validationErrors['title'] }}
        </div>
      </div>

      <div class="form-field">
        <label class="field-label">Purchase Price</label>
        <input
          type="number"
          v-model.number="baseFields.purchasePrice"
          placeholder="0.00"
          step="0.01"
          class="field-input"
        />
      </div>

      <!-- Dynamic attribute fields from schema -->
      <FormField
        v-for="attr in schema.attributes"
        :key="attr.name"
        :attribute="attr"
        :modelValue="attributes[attr.name]"
        :errorMessage="validationErrors[attr.name]"
        @update:modelValue="val => attributes[attr.name] = val"
      />

      <div class="form-actions">
        <button type="submit" class="btn btn-primary">Save</button>
        <button type="button" class="btn btn-secondary" @click="emit('cancel')">Cancel</button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.dynamic-form {
  padding: 16px;
}
.dynamic-form h3 {
  margin: 0 0 16px 0;
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
}
.has-error .field-input {
  border-color: #e53e3e;
}
.field-error {
  color: #e53e3e;
  font-size: 12px;
  margin-top: 2px;
}
.form-actions {
  margin-top: 16px;
  display: flex;
  gap: 8px;
}
.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}
.btn-primary {
  background: #3182ce;
  color: white;
}
.btn-primary:hover {
  background: #2c5282;
}
.btn-secondary {
  background: #e2e8f0;
  color: #333;
}
.btn-secondary:hover {
  background: #cbd5e0;
}
</style>
