<script lang="ts" setup>
import {reactive, ref, watch, computed} from 'vue'
import type {Item, ModuleSchema} from '../api/types'
import FormField from './FormField.vue'
import ImageAttach from './ImageAttach.vue'

const props = defineProps<{
  schema: ModuleSchema
  item?: Item | null
}>()

const emit = defineEmits<{
  save: [item: Item]
  cancel: []
}>()

const baseFields = reactive({
  title: '',
  purchasePrice: null as number | null,
})

const attributes = reactive<Record<string, any>>({})
const itemImages = ref<string[]>([])
const validationErrors = reactive<Record<string, string>>({})

const isEditing = computed(() => !!props.item?.id)

// Initialize form when item or schema changes
watch(() => [props.item, props.schema], () => {
  // Reset base fields and images
  baseFields.title = props.item?.title ?? ''
  baseFields.purchasePrice = props.item?.purchasePrice ?? null
  itemImages.value = props.item?.images ? [...props.item.images] : []

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

  const item: Item = {
    id: props.item?.id ?? '',
    moduleId: props.schema.id,
    title: baseFields.title.trim(),
    purchasePrice: baseFields.purchasePrice,
    images: itemImages.value,
    attributes: {...attributes},
    createdAt: props.item?.createdAt ?? '',
    updatedAt: '',
  }

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

      <!-- Image attachment -->
      <ImageAttach
        :images="itemImages"
        @update:images="val => itemImages = val"
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
  background: var(--accent-blue);
  color: var(--text-on-accent);
}
.btn-primary:hover {
  background: var(--accent-blue-hover);
}
.btn-secondary {
  background: var(--btn-secondary-bg);
  color: var(--text-primary);
}
.btn-secondary:hover {
  background: var(--btn-secondary-bg-hover);
}
</style>
