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
  aiFilledFields.value.clear()
  aiTitleSuggestion.value = ''

  try {
    const result = await analyzeItem(itemImages.value[0], props.schema.id)
    let filledCount = 0

    // Title handling
    if (result.title) {
      if (!baseFields.title.trim()) {
        baseFields.title = result.title
        aiFilledFields.value.add('title')
        filledCount++
      } else {
        aiTitleSuggestion.value = result.title
      }
    }

    // Populate only empty fields
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
    toastStore.show(`Smart Scan complete. Please review highlighted fields.`, 'success')
  } catch (err: any) {
    toastStore.show('AI Scan failed: ' + (err.message || err), 'error')
  } finally {
    aiAnalyzing.value = false
  }
}

function clearHighlight(fieldId: string) {
  if (aiFilledFields.value.has(fieldId)) {
    aiFilledFields.value.delete(fieldId)
  }
}

function acceptTitleSuggestion() {
  baseFields.title = aiTitleSuggestion.value
  aiTitleSuggestion.value = ''
  aiFilledFields.value.add('title')
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
  clearHighlight('title')
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
      <!-- Image Attachment Zone with Smart Scan -->
      <div class="image-attach-zone" :class="{'is-scanning': aiAnalyzing}">
        <ImageAttach
          :images="itemImages"
          @update:images="val => itemImages = val"
        />
        
        <!-- Sleek Loading State -->
        <div v-if="aiAnalyzing" class="scan-overlay">
          <div class="scan-line"></div>
          <div class="scan-pulse"></div>
        </div>

        <!-- AI Control -->
        <div v-if="aiStatus?.enabled" class="ai-controls">
          <button
            type="button"
            class="btn-smart-scan"
            :disabled="!aiButtonEnabled"
            :title="itemImages.length === 0 ? 'Add a photo first' : ''"
            @click="runAIAnalysis"
          >
            <svg v-if="!aiAnalyzing" class="sparkles-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M9.937 15.5A2 2 0 0 0 8.5 14.063l-6.135-1.582a.5.5 0 0 1 0-.962L8.5 9.936A2 2 0 0 0 9.937 8.5l1.582-6.135a.5.5 0 0 1 .963 0L14.063 8.5A2 2 0 0 0 15.5 9.937l6.135 1.581a.5.5 0 0 1 0 .964L15.5 14.063a2 2 0 0 0-1.437 1.437l-1.582 6.135a.5.5 0 0 1-.963 0z"/>
            </svg>
            {{ aiAnalyzing ? 'Scanning...' : 'Smart Scan' }}
          </button>
        </div>
      </div>

      <div class="form-grid">
        <!-- Base fields -->
        <div class="form-field-wrapper" :class="{'ai-highlight': aiFilledFields.has('title')}">
          <label class="field-label">Title <span class="required">*</span></label>
          <input
            type="text"
            v-model="baseFields.title"
            placeholder="Item title"
            class="field-input"
            @input="onTitleInput"
            @focus="clearHighlight('title')"
          />
          <div v-if="validationErrors['title']" class="field-error">
            {{ validationErrors['title'] }}
          </div>
          <div v-if="aiTitleSuggestion" class="ai-title-suggestion" @click="acceptTitleSuggestion">
            AI suggestion: {{ aiTitleSuggestion }}
          </div>
          <!-- Auto-fill Badge -->
          <div v-if="aiFilledFields.has('title')" class="ai-badge">
            <svg class="sparkles-icon" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9.937 15.5A2 2 0 0 0 8.5 14.063l-6.135-1.582a.5.5 0 0 1 0-.962L8.5 9.936A2 2 0 0 0 9.937 8.5l1.582-6.135a.5.5 0 0 1 .963 0L14.063 8.5A2 2 0 0 0 15.5 9.937l6.135 1.581a.5.5 0 0 1 0 .964L15.5 14.063a2 2 0 0 0-1.437 1.437l-1.582 6.135a.5.5 0 0 1-.963 0z"/></svg>
            Auto-filled
          </div>
        </div>

        <div class="form-field-wrapper" :class="{'ai-highlight': aiFilledFields.has('purchasePrice')}">
          <label class="field-label">Purchase Price</label>
          <input
            type="number"
            v-model.number="baseFields.purchasePrice"
            placeholder="0.00"
            step="0.01"
            class="field-input"
            @input="clearHighlight('purchasePrice')"
            @focus="clearHighlight('purchasePrice')"
          />
          <div v-if="aiFilledFields.has('purchasePrice')" class="ai-badge">
            <svg class="sparkles-icon" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9.937 15.5A2 2 0 0 0 8.5 14.063l-6.135-1.582a.5.5 0 0 1 0-.962L8.5 9.936A2 2 0 0 0 9.937 8.5l1.582-6.135a.5.5 0 0 1 .963 0L14.063 8.5A2 2 0 0 0 15.5 9.937l6.135 1.581a.5.5 0 0 1 0 .964L15.5 14.063a2 2 0 0 0-1.437 1.437l-1.582 6.135a.5.5 0 0 1-.963 0z"/></svg>
            Auto-filled
          </div>
        </div>

        <!-- Dynamic attribute fields from schema -->
        <div 
          v-for="attr in schema.attributes" 
          :key="attr.name" 
          class="form-field-wrapper"
          :class="{'ai-highlight': aiFilledFields.has(attr.name)}"
          @focusin="clearHighlight(attr.name)"
        >
          <FormField
            :attribute="attr"
            :modelValue="attributes[attr.name]"
            :errorMessage="validationErrors[attr.name]"
            @update:modelValue="val => { attributes[attr.name] = val; clearHighlight(attr.name); }"
          />
          <div v-if="aiFilledFields.has(attr.name)" class="ai-badge">
            <svg class="sparkles-icon" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9.937 15.5A2 2 0 0 0 8.5 14.063l-6.135-1.582a.5.5 0 0 1 0-.962L8.5 9.936A2 2 0 0 0 9.937 8.5l1.582-6.135a.5.5 0 0 1 .963 0L14.063 8.5A2 2 0 0 0 15.5 9.937l6.135 1.581a.5.5 0 0 1 0 .964L15.5 14.063a2 2 0 0 0-1.437 1.437l-1.582 6.135a.5.5 0 0 1-.963 0z"/></svg>
            Auto-filled
          </div>
        </div>
      </div>

      <!-- Tags -->
      <div class="form-field-wrapper tags-wrapper">
        <TagInput
          v-model="itemTags"
          :allTags="allTags"
        />
      </div>

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
  max-width: 800px;
  margin: 0 auto;
}
.dynamic-form h3 {
  margin: 0 0 24px 0;
  font-family: 'Instrument Serif', serif;
  font-size: 28px;
  letter-spacing: var(--tracking-tight);
}

/* Image Zone & Scan Animation */
.image-attach-zone {
  position: relative;
  background: var(--bg-tertiary);
  border: 1px dashed var(--border-primary);
  border-radius: var(--radius-md);
  padding: 16px;
  margin-bottom: 24px;
  overflow: hidden;
  min-height: 120px;
}
.image-attach-zone.is-scanning {
  border-color: var(--accent-blue);
}

.scan-overlay {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 10;
  overflow: hidden;
  border-radius: var(--radius-md);
}
.scan-line {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--accent-blue);
  box-shadow: 0 0 12px 2px var(--accent-blue);
  animation: scan-line 2s linear infinite;
}
.scan-pulse {
  position: absolute;
  inset: 0;
  background: var(--accent-blue-light);
  animation: pulse 2s infinite ease-in-out;
}

@keyframes scan-line {
  0% { transform: translateY(-10px); }
  100% { transform: translateY(150px); }
}
@keyframes pulse {
  0%, 100% { opacity: 0.2; }
  50% { opacity: 0.6; }
}

.ai-controls {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
.btn-smart-scan {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: 1px solid var(--accent-blue);
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  background: var(--bg-primary);
  color: var(--accent-blue);
  transition: all var(--transition-fast);
  box-shadow: var(--shadow-sm);
}
.btn-smart-scan:hover:not(:disabled) {
  background: var(--accent-blue);
  color: var(--text-on-accent);
}
.btn-smart-scan:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  filter: grayscale(1);
}

/* Form Layout */
.form-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
  margin-bottom: 24px;
}
@media (min-width: 640px) {
  .form-grid {
    grid-template-columns: 1fr 1fr;
  }
}

.form-field-wrapper {
  position: relative;
  transition: background-color var(--transition-normal), border-color var(--transition-normal);
  padding: 12px;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  display: flex;
  flex-direction: column;
}
.form-field-wrapper :deep(.form-field) {
  margin-bottom: 0;
}
.tags-wrapper {
  padding: 0; /* Align with layout */
}

.field-label {
  display: block;
  font-weight: 600;
  margin-bottom: 6px;
  font-size: 13px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.required {
  color: var(--required-color);
}
.field-input {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--border-input);
  border-radius: 6px;
  font-size: 14px;
  box-sizing: border-box;
  background: var(--bg-input);
}
.form-field-wrapper:has(.has-error) .field-input,
.has-error .field-input {
  border-color: var(--error-text);
}
.field-error {
  color: var(--error-text);
  font-size: 12px;
  margin-top: 4px;
}

/* AI Highlight State */
.ai-highlight {
  background-color: var(--accent-blue-light);
  border-left: 3px solid var(--accent-blue);
}
.ai-badge {
  position: absolute;
  top: 8px;
  right: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  text-transform: uppercase;
  font-weight: 700;
  color: var(--accent-blue);
  pointer-events: none;
}
.ai-title-suggestion {
  font-size: 12px;
  color: var(--accent-blue);
  margin-top: 6px;
  cursor: pointer;
}
.ai-title-suggestion:hover {
  text-decoration: underline;
}

/* Buttons */
.form-actions {
  margin-top: 24px;
  display: flex;
  gap: 12px;
  border-top: 1px solid var(--border-primary);
  padding-top: 16px;
}
.btn {
  padding: 10px 20px;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background var(--transition-fast);
}
.btn-primary {
  background: var(--accent-blue);
  color: var(--text-on-accent);
  box-shadow: var(--shadow-sm);
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
