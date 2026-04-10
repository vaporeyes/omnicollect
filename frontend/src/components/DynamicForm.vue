<script lang="ts" setup>
import {reactive, ref, watch, computed, onMounted, nextTick} from 'vue'
import type {Item, ModuleSchema, TagCount, AIStatus} from '../api/types'
import {getAllTags, getAIStatus, analyzeItem} from '../api/client'
import {useToastStore} from '../stores/toastStore'
import FormField from './FormField.vue'
import ImageAttach from './ImageAttach.vue'
import TagInput from './TagInput.vue'

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
const itemTags = ref<string[]>([])
const allTags = ref<TagCount[]>([])
const validationErrors = reactive<Record<string, string>>({})

const isEditing = computed(() => !!props.item?.id)

// AI analysis state
const toastStore = useToastStore()
const aiStatus = ref<AIStatus | null>(null)
const aiAnalyzing = ref(false)
const aiFilledFields = ref<Set<string>>(new Set())
const aiTitleSuggestion = ref('')

const aiButtonEnabled = computed(() => {
  return aiStatus.value?.enabled && itemImages.value.length > 0 && !aiAnalyzing.value
})

onMounted(async () => {
  try { allTags.value = await getAllTags() } catch { /* ignore */ }
  try { aiStatus.value = await getAIStatus() } catch { /* ignore */ }
})

async function runAIAnalysis() {
  if (!aiButtonEnabled.value) return
  aiAnalyzing.value = true
  aiFilledFields.value = new Set()
  aiTitleSuggestion.value = ''

  try {
    const result = await analyzeItem(itemImages.value[0], props.schema.id)
    let filledCount = 0

    // Title handling (US3): auto-fill if empty, suggest if already has value
    if (result.title) {
      if (!baseFields.title.trim()) {
        baseFields.title = result.title
        aiFilledFields.value.add('title')
        filledCount++
      } else {
        aiTitleSuggestion.value = result.title
      }
    }

    // Populate only empty fields (US2)
    for (const [key, val] of Object.entries(result.attributes)) {
      const current = attributes[key]
      const isEmpty = current === null || current === undefined ||
        (typeof current === 'string' && current.trim() === '') ||
        (typeof current === 'number' && isNaN(current))
      if (isEmpty && key in attributes) {
        attributes[key] = val
        aiFilledFields.value.add(key)
        filledCount++
      }
    }

    if (result.warnings?.length > 0) {
      toastStore.show(result.warnings.join('; '), 'info')
    }
    toastStore.show(`AI filled ${filledCount} field${filledCount !== 1 ? 's' : ''}`, 'success')

    // Clear highlight after a delay (US2 T017)
    setTimeout(() => { aiFilledFields.value = new Set() }, 2000)
  } catch (err: any) {
    toastStore.show('AI analysis failed: ' + (err.message || err), 'error')
  } finally {
    aiAnalyzing.value = false
  }
}

function acceptTitleSuggestion() {
  baseFields.title = aiTitleSuggestion.value
  aiTitleSuggestion.value = ''
}

// Initialize form when item or schema changes
watch(() => [props.item, props.schema], () => {
  // Reset base fields, images, and tags
  baseFields.title = props.item?.title ?? ''
  baseFields.purchasePrice = props.item?.purchasePrice ?? null
  itemImages.value = props.item?.images ? [...props.item.images] : []
  itemTags.value = props.item?.tags ? [...props.item.tags] : []

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

function onTitleInput() {
  aiTitleSuggestion.value = ''
}

function onSubmit() {
  aiTitleSuggestion.value = ''
  if (!validate()) return

  const item: Item = {
    id: props.item?.id ?? '',
    moduleId: props.schema.id,
    title: baseFields.title.trim(),
    purchasePrice: baseFields.purchasePrice,
    images: itemImages.value,
    tags: itemTags.value,
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
      <div class="form-field" :class="{'has-error': validationErrors['title'], 'ai-filled': aiFilledFields.has('title')}">
        <label class="field-label">Title <span class="required">*</span></label>
        <input
          type="text"
          v-model="baseFields.title"
          placeholder="Item title"
          class="field-input"
          @input="onTitleInput"
        />
        <div v-if="validationErrors['title']" class="field-error">
          {{ validationErrors['title'] }}
        </div>
        <div v-if="aiTitleSuggestion" class="ai-title-suggestion" @click="acceptTitleSuggestion">
          AI suggestion: {{ aiTitleSuggestion }}
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
        :class="{'ai-filled': aiFilledFields.has(attr.name)}"
        @update:modelValue="val => attributes[attr.name] = val"
      />

      <!-- Image attachment -->
      <ImageAttach
        :images="itemImages"
        @update:images="val => itemImages = val"
      />

      <!-- AI analysis button -->
      <div v-if="aiStatus?.enabled" class="ai-section">
        <button
          type="button"
          class="btn btn-ai"
          :disabled="!aiButtonEnabled"
          :title="itemImages.length === 0 ? 'Add a photo first' : ''"
          @click="runAIAnalysis"
        >
          {{ aiAnalyzing ? 'Analyzing...' : 'Analyze with AI' }}
        </button>
      </div>

      <!-- Tags -->
      <TagInput
        v-model="itemTags"
        :allTags="allTags"
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
.ai-section {
  margin: 12px 0;
}
.btn-ai {
  padding: 8px 16px;
  border: 1px solid var(--accent-blue);
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  background: transparent;
  color: var(--accent-blue);
}
.btn-ai:hover:not(:disabled) {
  background: var(--accent-blue);
  color: var(--text-on-accent);
}
.btn-ai:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.ai-filled {
  animation: ai-highlight 2s ease-out;
}
@keyframes ai-highlight {
  0% { background-color: rgba(59, 130, 246, 0.15); }
  100% { background-color: transparent; }
}
.ai-title-suggestion {
  font-size: 12px;
  color: var(--accent-blue);
  margin-top: 4px;
  cursor: pointer;
}
.ai-title-suggestion:hover {
  text-decoration: underline;
}
</style>
