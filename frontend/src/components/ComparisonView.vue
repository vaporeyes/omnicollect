<!-- ABOUTME: Side-by-side comparison view for exactly two collection items. -->
<!-- ABOUTME: Synchronized image galleries and attribute diff table with difference highlighting. -->
<script lang="ts" setup>
import {ref, computed} from 'vue'
import type {Item, ModuleSchema, AttributeSchema} from '../api/types'

const props = defineProps<{
  itemA: Item
  itemB: Item
  schemaA: ModuleSchema | null
  schemaB: ModuleSchema | null
}>()

const emit = defineEmits<{
  close: []
  viewImage: [filename: string]
}>()

// Synchronized gallery state
const activeImageIndex = ref(0)

const maxIndexA = computed(() => Math.max(0, (props.itemA.images?.length ?? 0) - 1))
const maxIndexB = computed(() => Math.max(0, (props.itemB.images?.length ?? 0) - 1))
const effectiveIndexA = computed(() => Math.min(activeImageIndex.value, maxIndexA.value))
const effectiveIndexB = computed(() => Math.min(activeImageIndex.value, maxIndexB.value))
const maxIndex = computed(() => Math.max(maxIndexA.value, maxIndexB.value))
const hasAnyImages = computed(() =>
  (props.itemA.images?.length ?? 0) > 0 || (props.itemB.images?.length ?? 0) > 0
)

function prevImage() {
  if (activeImageIndex.value > 0) activeImageIndex.value--
}
function nextImage() {
  if (activeImageIndex.value < maxIndex.value) activeImageIndex.value++
}

// Diff computation
interface DiffRow {
  label: string
  valueA: string | null
  valueB: string | null
  isDifferent: boolean
}

function formatPrice(val: number | null | undefined): string | null {
  if (val == null) return null
  return '$' + val.toFixed(2)
}

function formatTags(tags: string[] | null | undefined): string | null {
  if (!tags || tags.length === 0) return null
  return [...tags].sort().join(', ')
}

function formatAttrValue(val: any): string | null {
  if (val == null || val === '') return null
  if (typeof val === 'boolean') return val ? 'Yes' : 'No'
  return String(val)
}

const diffRows = computed<DiffRow[]>(() => {
  const rows: DiffRow[] = []

  // Core fields
  const titleA = props.itemA.title || null
  const titleB = props.itemB.title || null
  rows.push({label: 'Title', valueA: titleA, valueB: titleB, isDifferent: titleA !== titleB})

  const priceA = formatPrice(props.itemA.purchasePrice)
  const priceB = formatPrice(props.itemB.purchasePrice)
  rows.push({label: 'Purchase Price', valueA: priceA, valueB: priceB, isDifferent: priceA !== priceB})

  const tagsA = formatTags(props.itemA.tags)
  const tagsB = formatTags(props.itemB.tags)
  rows.push({label: 'Tags', valueA: tagsA, valueB: tagsB, isDifferent: tagsA !== tagsB})

  // Union of schema attributes
  const seenAttrs = new Set<string>()
  const allAttrs: {name: string, label: string}[] = []

  function addAttrs(schema: ModuleSchema | null) {
    if (!schema) return
    for (const attr of schema.attributes) {
      if (!seenAttrs.has(attr.name)) {
        seenAttrs.add(attr.name)
        allAttrs.push({name: attr.name, label: attr.display?.label || attr.name})
      }
    }
  }
  addAttrs(props.schemaA)
  addAttrs(props.schemaB)

  for (const attr of allAttrs) {
    const valA = formatAttrValue(props.itemA.attributes?.[attr.name])
    const valB = formatAttrValue(props.itemB.attributes?.[attr.name])
    rows.push({label: attr.label, valueA: valA, valueB: valB, isDifferent: valA !== valB})
  }

  return rows
})

function imageErrorHandler(event: Event) {
  const img = event.target as HTMLImageElement
  img.style.display = 'none'
  const placeholder = img.nextElementSibling as HTMLElement
  if (placeholder) placeholder.style.display = 'flex'
}
</script>

<template>
  <div class="comparison-view">
    <div class="comparison-header">
      <button class="back-btn" @click="emit('close')">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M19 12H5"/><polyline points="12 19 5 12 12 5"/></svg>
        Back to Collection
      </button>
      <h2 class="comparison-title">Comparing Items</h2>
    </div>

    <!-- Galleries side by side -->
    <div class="galleries">
      <div class="gallery-side">
        <div class="gallery-frame">
          <template v-if="itemA.images && itemA.images.length > 0">
            <img
              :src="'/originals/' + encodeURIComponent(itemA.images[effectiveIndexA])"
              alt=""
              @error="imageErrorHandler"
              @click="emit('viewImage', itemA.images[effectiveIndexA])"
            />
            <div class="gallery-placeholder" style="display: none">
              <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.3">
                <rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="M21 15l-5-5L5 21"/>
              </svg>
            </div>
          </template>
          <div v-else class="gallery-placeholder">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.3">
              <rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="M21 15l-5-5L5 21"/>
            </svg>
            <span>No images</span>
          </div>
          <div v-if="itemA.images && itemA.images.length > 0" class="gallery-counter">
            {{ effectiveIndexA + 1 }} / {{ itemA.images.length }}
          </div>
        </div>
        <div class="gallery-label">{{ itemA.title }}</div>
      </div>

      <div class="gallery-side">
        <div class="gallery-frame">
          <template v-if="itemB.images && itemB.images.length > 0">
            <img
              :src="'/originals/' + encodeURIComponent(itemB.images[effectiveIndexB])"
              alt=""
              @error="imageErrorHandler"
              @click="emit('viewImage', itemB.images[effectiveIndexB])"
            />
            <div class="gallery-placeholder" style="display: none">
              <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.3">
                <rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="M21 15l-5-5L5 21"/>
              </svg>
            </div>
          </template>
          <div v-else class="gallery-placeholder">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.3">
              <rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="M21 15l-5-5L5 21"/>
            </svg>
            <span>No images</span>
          </div>
          <div v-if="itemB.images && itemB.images.length > 0" class="gallery-counter">
            {{ effectiveIndexB + 1 }} / {{ itemB.images.length }}
          </div>
        </div>
        <div class="gallery-label">{{ itemB.title }}</div>
      </div>
    </div>

    <!-- Synchronized navigation controls -->
    <div v-if="hasAnyImages" class="gallery-nav">
      <button class="nav-btn" :disabled="activeImageIndex === 0" @click="prevImage">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="15 18 9 12 15 6"/></svg>
        Prev
      </button>
      <span class="nav-position">{{ activeImageIndex + 1 }}</span>
      <button class="nav-btn" :disabled="activeImageIndex >= maxIndex" @click="nextImage">
        Next
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="9 18 15 12 9 6"/></svg>
      </button>
    </div>

    <!-- Diff table -->
    <div class="diff-section">
      <h3 class="diff-heading">Attributes</h3>
      <table class="diff-table">
        <thead>
          <tr>
            <th class="diff-label-col">Field</th>
            <th>{{ itemA.title }}</th>
            <th>{{ itemB.title }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in diffRows"
            :key="row.label"
            :class="{'diff-row': row.isDifferent}"
          >
            <td class="diff-label">{{ row.label }}</td>
            <td>{{ row.valueA ?? '---' }}</td>
            <td>{{ row.valueB ?? '---' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.comparison-view {
  max-width: 1200px;
  margin: 0 auto;
}
.comparison-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}
.back-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 13px;
  transition: background var(--transition-fast), color var(--transition-fast);
}
.back-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}
.comparison-title {
  margin: 0;
  font-family: 'Instrument Serif', serif;
  font-size: 22px;
  font-weight: 400;
  color: var(--text-primary);
}

/* Galleries */
.galleries {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
  margin-bottom: 16px;
}
.gallery-side {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.gallery-frame {
  position: relative;
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  aspect-ratio: 4/3;
}
.gallery-frame img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
  cursor: pointer;
}
.gallery-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--text-muted);
  font-size: 13px;
}
.gallery-counter {
  position: absolute;
  bottom: 8px;
  right: 8px;
  padding: 2px 8px;
  background: hsla(0, 0%, 0%, 0.5);
  color: #fff;
  font-size: 11px;
  border-radius: var(--radius-sm);
}
.gallery-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Navigation controls */
.gallery-nav {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-bottom: 24px;
}
.nav-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 14px;
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 13px;
  transition: background var(--transition-fast);
}
.nav-btn:hover:not(:disabled) {
  background: var(--bg-hover);
  color: var(--text-primary);
}
.nav-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}
.nav-position {
  font-size: 13px;
  color: var(--text-muted);
  min-width: 20px;
  text-align: center;
}

/* Diff table */
.diff-section {
  margin-top: 8px;
}
.diff-heading {
  font-family: 'Instrument Serif', serif;
  font-size: 18px;
  font-weight: 400;
  color: var(--text-primary);
  margin: 0 0 12px;
}
.diff-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.diff-table th {
  text-align: left;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-muted);
  border-bottom: 2px solid var(--border-primary);
}
.diff-table td {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-primary);
  color: var(--text-primary);
}
.diff-label-col {
  width: 160px;
}
.diff-label {
  font-weight: 500;
  color: var(--text-secondary);
}
.diff-row td {
  background: hsla(40, 70%, 55%, 0.1);
}

/* Responsive stacking */
@media (max-width: 768px) {
  .galleries {
    grid-template-columns: 1fr;
  }
}
</style>
