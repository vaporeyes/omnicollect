<script lang="ts" setup>
import MediaImage from './MediaImage.vue'
import {ref, computed} from 'vue'
import type {Item, ModuleSchema, AttributeSchema} from '../api/types'
import MarkdownRenderer from './MarkdownRenderer.vue'

const props = defineProps<{
  item: Item
  schema: ModuleSchema | null
}>()

const emit = defineEmits<{
  edit: []
  delete: []
  close: []
  viewImage: [filename: string]
}>()

const activeImageIndex = ref(0)
const hasImages = computed(() => props.item.images && props.item.images.length > 0)
const imageCount = computed(() => props.item.images?.length ?? 0)

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  try {
    return new Date(dateStr).toLocaleDateString(undefined, {
      year: 'numeric', month: 'long', day: 'numeric'
    })
  } catch {
    return dateStr
  }
}

function formatPrice(price: number | null | undefined): string {
  if (price === null || price === undefined) return ''
  return `$${price.toFixed(2)}`
}

function formatAttrValue(value: any, type: string): string {
  if (value === null || value === undefined) return ''
  if (type === 'boolean') return value ? 'Yes' : 'No'
  if (type === 'date') return formatDate(String(value))
  return String(value)
}

function isTextarea(attr: AttributeSchema): boolean {
  return attr.display?.widget === 'textarea'
}

function prevImage() {
  if (activeImageIndex.value > 0) activeImageIndex.value--
}

function nextImage() {
  if (activeImageIndex.value < imageCount.value - 1) activeImageIndex.value++
}
</script>

<template>
  <div class="item-detail">
    <!-- Top bar: back + actions -->
    <div class="detail-topbar">
      <button class="back-btn" @click="emit('close')" title="Back to list">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M19 12H5"/><path d="M12 19l-7-7 7-7"/>
        </svg>
        <span>Back</span>
      </button>
      <div class="topbar-actions">
        <button class="action-btn action-edit" @click="emit('edit')" title="Edit item">Edit</button>
        <button class="action-btn action-delete" @click="emit('delete')" title="Delete item">Delete</button>
      </div>
    </div>

    <!-- Split layout -->
    <div class="detail-split">
      <!-- Left: sticky image gallery -->
      <div class="split-left">
        <div class="gallery-sticky">
          <div v-if="hasImages" class="gallery-main">
            <button class="gallery-open" :aria-label="'Open full-size image for ' + item.title" @click="emit('viewImage', item.images[activeImageIndex])">
              <MediaImage :src="'/originals/' + encodeURIComponent(item.images[activeImageIndex])" alt="" class="gallery-image" />
            </button>
            <div class="gallery-inner-shadow"></div>
            <!-- Nav arrows overlaid on image -->
            <button
              v-if="imageCount > 1"
              class="gallery-arrow gallery-arrow-prev" aria-label="Previous image"
              :disabled="activeImageIndex === 0"
              @click="prevImage"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M15 18l-6-6 6-6"/></svg>
            </button>
            <button
              v-if="imageCount > 1"
              class="gallery-arrow gallery-arrow-next" aria-label="Next image"
              :disabled="activeImageIndex === imageCount - 1"
              @click="nextImage"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M9 18l6-6-6-6"/></svg>
            </button>
            <!-- Counter pill -->
            <div v-if="imageCount > 1" class="gallery-counter">
              {{ activeImageIndex + 1 }} / {{ imageCount }}
            </div>
          </div>

          <div v-else class="no-images">
            <svg width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2">
              <rect x="3" y="3" width="18" height="18" rx="2"/>
              <circle cx="8.5" cy="8.5" r="1.5"/>
              <path d="M21 15l-5-5L5 21"/>
            </svg>
            <p>No images</p>
          </div>

          <!-- Thumbnail strip -->
          <div v-if="imageCount > 1" class="gallery-thumbs">
            <button v-for="(filename, idx) in item.images" :key="filename" :class="['thumb', {active: idx === activeImageIndex}]" :aria-label="'Show image ' + (idx + 1)" :aria-pressed="idx === activeImageIndex" @click="activeImageIndex = idx">
              <MediaImage :src="'/thumbnails/' + encodeURIComponent(filename)" alt="" class="thumb-image" />
            </button>
          </div>
        </div>
      </div>

      <!-- Right: scrolling metadata -->
      <div class="split-right">
        <!-- Collection badge -->
        <span class="collection-badge">{{ schema?.displayName ?? item.moduleId }}</span>

        <h2 class="item-title">{{ item.title }}</h2>

        <!-- Price callout -->
        <div v-if="item.purchasePrice !== null && item.purchasePrice !== undefined" class="price-display">
          {{ formatPrice(item.purchasePrice) }}
        </div>

        <!-- Tags -->
        <div v-if="item.tags && item.tags.length > 0" class="tag-display">
          <span v-for="tag in item.tags" :key="tag" class="detail-tag-chip">{{ tag }}</span>
        </div>

        <!-- Divider -->
        <hr class="section-rule" />

        <!-- Details section -->
        <div class="info-section">
          <h3 class="section-heading">Provenance</h3>
          <dl class="info-grid">
            <div class="info-row">
              <dt>Added</dt>
              <dd>{{ formatDate(item.createdAt) }}</dd>
            </div>
            <div class="info-row">
              <dt>Last Updated</dt>
              <dd>{{ formatDate(item.updatedAt) }}</dd>
            </div>
          </dl>
        </div>

        <!-- Custom attributes from schema -->
        <div v-if="schema && schema.attributes.length > 0" class="info-section">
          <h3 class="section-heading">Attributes</h3>
          <!-- Textarea attributes render as Markdown -->
          <div v-for="attr in schema.attributes" :key="attr.name">
            <div v-if="isTextarea(attr) && item.attributes?.[attr.name]" class="attr-markdown">
              <div class="facet-label-sm">{{ attr.display?.label || attr.name }}</div>
              <MarkdownRenderer :content="String(item.attributes[attr.name])" />
            </div>
            <dl v-else class="info-grid">
              <div class="info-row">
                <dt>{{ attr.display?.label || attr.name }}</dt>
                <dd>{{ formatAttrValue(item.attributes?.[attr.name], attr.type) || '--' }}</dd>
              </div>
            </dl>
          </div>
        </div>

        <!-- Raw attributes when no schema available -->
        <div v-else-if="item.attributes && Object.keys(item.attributes).length > 0" class="info-section">
          <h3 class="section-heading">Attributes</h3>
          <dl class="info-grid">
            <div v-for="(value, key) in item.attributes" :key="key" class="info-row">
              <dt>{{ key }}</dt>
              <dd>{{ value ?? '--' }}</dd>
            </div>
          </dl>
        </div>
      </div>
    </div>

  </div>
</template>

<style scoped>
/* --- Top bar --- */
.detail-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-lg);
}
.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 13px;
  font-family: var(--font-body);
  font-weight: 500;
  padding: 6px 2px;
  transition: color var(--transition-fast);
}
.back-btn:hover {
  color: var(--text-primary);
}
.topbar-actions {
  display: flex;
  gap: 8px;
}
.action-btn {
  padding: 7px 16px;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 12px;
  font-family: var(--font-body);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide);
  transition: background var(--transition-fast), color var(--transition-fast);
}
.action-edit {
  background: var(--accent-blue);
  color: var(--text-on-accent);
}
.action-edit:hover {
  background: var(--accent-blue-hover);
}
.action-delete {
  background: var(--error-bg, rgba(239, 68, 68, 0.12));
  color: var(--error-text, #ef4444);
}
.action-delete:hover {
  background: var(--error-border, #dc2626);
  color: #fff;
}

/* --- Split layout --- */
.item-detail {
  max-width: 1100px;
}
.detail-split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-xl);
  align-items: start;
}

/* --- Left column: sticky gallery --- */
.split-left {
  position: relative;
}
.gallery-sticky {
  position: sticky;
  top: 0;
}
.gallery-main {
  position: relative;
  border-radius: var(--radius-md);
  overflow: hidden;
  background: var(--bg-secondary);
  aspect-ratio: 1 / 1;
  display: flex;
  align-items: center;
  justify-content: center;
}
.gallery-open { width: 100%; height: 100%; padding: 0; border: 0; background: transparent; cursor: zoom-in; }
.thumb-image { width: 100%; height: 100%; object-fit: cover; }
.gallery-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  cursor: zoom-in;
  transition: transform 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}
.gallery-image:hover {
  transform: scale(1.02);
}
/* Subtle inner shadow for depth */
.gallery-inner-shadow {
  position: absolute;
  inset: 0;
  pointer-events: none;
  border-radius: var(--radius-md);
  box-shadow: inset 0 2px 20px rgba(0, 0, 0, 0.12), inset 0 0 1px rgba(0, 0, 0, 0.06);
}

/* Nav arrows */
.gallery-arrow {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: var(--overlay-bg);
  color: rgba(255, 255, 255, 0.9);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity var(--transition-fast);
}
.gallery-main:hover .gallery-arrow,
.gallery-main:focus-within .gallery-arrow {
  opacity: 1;
}
.gallery-arrow:disabled {
  opacity: 0 !important;
  cursor: default;
}
.gallery-arrow-prev {
  left: 10px;
}
.gallery-arrow-next {
  right: 10px;
}
.gallery-counter {
  position: absolute;
  bottom: 10px;
  right: 10px;
  background: var(--overlay-bg);
  color: rgba(255, 255, 255, 0.8);
  font-size: 11px;
  font-weight: 500;
  font-family: var(--font-body);
  padding: 3px 10px;
  border-radius: 12px;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.04em;
}

/* Thumbnail strip */
.gallery-thumbs {
  display: flex;
  gap: 8px;
  margin-top: 10px;
  overflow-x: auto;
  padding: 2px 0;
}
.thumb {
  padding: 0;
  background: transparent;
  width: 52px;
  height: 52px;
  object-fit: cover;
  border-radius: var(--radius-sm);
  cursor: pointer;
  border: 2px solid transparent;
  opacity: 0.5;
  transition: opacity var(--transition-fast), border-color var(--transition-fast);
  flex-shrink: 0;
}
.thumb:hover {
  opacity: 0.85;
}
.thumb.active {
  opacity: 1;
  border-color: var(--accent-blue);
}

/* No images placeholder */
.no-images {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  aspect-ratio: 1 / 1;
  color: var(--text-muted);
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
}
.no-images svg {
  opacity: 0.25;
  margin-bottom: 10px;
}
.no-images p {
  margin: 0;
  font-size: 13px;
  font-family: var(--font-body);
}

/* --- Right column: scrolling metadata --- */
.split-right {
  padding-top: var(--space-sm);
  min-height: 0;
}
.collection-badge {
  display: inline-block;
  font-family: var(--font-body);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide);
  color: var(--accent-blue);
  margin-bottom: var(--space-sm);
}
.item-title {
  font-family: var(--font-heading);
  font-size: clamp(28px, 4vw, 42px);
  font-weight: 400;
  line-height: var(--leading-tight);
  letter-spacing: var(--tracking-tight);
  color: var(--text-primary);
  margin: 0 0 var(--space-md);
}
.price-display {
  font-family: var(--font-mono);
  font-size: 20px;
  font-weight: 400;
  color: var(--text-secondary);
  letter-spacing: var(--tracking-tight);
  font-variant-numeric: tabular-nums;
  margin-bottom: var(--space-md);
}

/* Divider */
.section-rule {
  border: none;
  border-top: 1px solid var(--border-primary);
  margin: var(--space-lg) 0;
}

/* Info sections */
.info-section {
  margin-bottom: var(--space-lg);
}
.section-heading {
  font-family: var(--font-body);
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide);
  color: var(--text-muted);
  margin: 0 0 var(--space-sm);
}

/* Key-value grid using definition list */
.info-grid {
  margin: 0;
}
.info-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-primary);
}
.info-row:last-child {
  border-bottom: none;
}
.info-row dt {
  font-family: var(--font-body);
  font-size: 13px;
  font-weight: 400;
  color: var(--text-secondary);
}
.info-row dd {
  margin: 0;
  font-family: var(--font-body);
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  text-align: right;
}

/* Markdown attribute blocks */
.attr-markdown {
  margin-bottom: var(--space-sm);
}
.facet-label-sm {
  font-family: var(--font-body);
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide);
  color: var(--text-muted);
  margin-bottom: 4px;
}

/* Tag display chips */
.tag-display {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: var(--space-md);
}
.detail-tag-chip {
  display: inline-block;
  padding: 3px 10px;
  background: var(--accent-blue-light, rgba(59,130,246,0.12));
  color: var(--accent-blue);
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  font-family: var(--font-body);
}
</style>
