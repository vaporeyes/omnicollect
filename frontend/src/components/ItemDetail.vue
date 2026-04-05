<script lang="ts" setup>
import {ref, computed} from 'vue'
import {main} from '../../wailsjs/go/models'

const props = defineProps<{
  item: main.Item
  schema: main.ModuleSchema | null
}>()

const emit = defineEmits<{
  edit: []
  close: []
  viewImage: [filename: string]
}>()

const activeImageIndex = ref(0)

const hasImages = computed(() => props.item.images && props.item.images.length > 0)
const imageCount = computed(() => props.item.images?.length ?? 0)

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  try {
    return new Date(dateStr).toLocaleString()
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

function prevImage() {
  if (activeImageIndex.value > 0) activeImageIndex.value--
}

function nextImage() {
  if (activeImageIndex.value < imageCount.value - 1) activeImageIndex.value++
}
</script>

<template>
  <div class="item-detail">
    <div class="detail-header">
      <button class="back-btn" @click="emit('close')" title="Back to list">&larr;</button>
      <h2>{{ item.title }}</h2>
      <button class="edit-btn" @click="emit('edit')" title="Edit item">&#9998; Edit</button>
    </div>

    <!-- Image gallery -->
    <div v-if="hasImages" class="image-gallery">
      <div class="gallery-main">
        <img
          :src="'/originals/' + encodeURIComponent(item.images[activeImageIndex])"
          alt=""
          class="gallery-image"
          @click="emit('viewImage', item.images[activeImageIndex])"
        />
        <div v-if="imageCount > 1" class="gallery-nav">
          <button class="nav-btn" :disabled="activeImageIndex === 0" @click="prevImage">&lsaquo;</button>
          <span class="image-counter">{{ activeImageIndex + 1 }} / {{ imageCount }}</span>
          <button class="nav-btn" :disabled="activeImageIndex === imageCount - 1" @click="nextImage">&rsaquo;</button>
        </div>
      </div>
      <div v-if="imageCount > 1" class="gallery-thumbs">
        <img
          v-for="(filename, idx) in item.images"
          :key="filename"
          :src="'/thumbnails/' + encodeURIComponent(filename)"
          :class="['thumb', {active: idx === activeImageIndex}]"
          @click="activeImageIndex = idx"
          alt=""
        />
      </div>
    </div>

    <div v-else class="no-images">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <rect x="3" y="3" width="18" height="18" rx="2"/>
        <circle cx="8.5" cy="8.5" r="1.5"/>
        <path d="M21 15l-5-5L5 21"/>
      </svg>
      <p>No images attached</p>
    </div>

    <!-- Item info -->
    <div class="detail-info">
      <div class="info-section">
        <h3>Details</h3>
        <div class="info-row">
          <span class="info-label">Collection</span>
          <span class="info-value">{{ schema?.displayName ?? item.moduleId }}</span>
        </div>
        <div v-if="item.purchasePrice !== null && item.purchasePrice !== undefined" class="info-row">
          <span class="info-label">Purchase Price</span>
          <span class="info-value">{{ formatPrice(item.purchasePrice) }}</span>
        </div>
        <div class="info-row">
          <span class="info-label">Created</span>
          <span class="info-value">{{ formatDate(item.createdAt) }}</span>
        </div>
        <div class="info-row">
          <span class="info-label">Last Modified</span>
          <span class="info-value">{{ formatDate(item.updatedAt) }}</span>
        </div>
      </div>

      <!-- Custom attributes from schema -->
      <div v-if="schema && schema.attributes.length > 0" class="info-section">
        <h3>Attributes</h3>
        <div v-for="attr in schema.attributes" :key="attr.name" class="info-row">
          <span class="info-label">{{ attr.display?.label || attr.name }}</span>
          <span class="info-value">
            {{ formatAttrValue(item.attributes?.[attr.name], attr.type) || '--' }}
          </span>
        </div>
      </div>

      <!-- Raw attributes when no schema available -->
      <div v-else-if="item.attributes && Object.keys(item.attributes).length > 0" class="info-section">
        <h3>Attributes</h3>
        <div v-for="(value, key) in item.attributes" :key="key" class="info-row">
          <span class="info-label">{{ key }}</span>
          <span class="info-value">{{ value ?? '--' }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.item-detail {
  max-width: 800px;
}
.detail-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.detail-header h2 {
  flex: 1;
  margin: 0;
  font-size: 20px;
}
.back-btn {
  background: none;
  border: 1px solid var(--border-primary);
  border-radius: 4px;
  padding: 6px 10px;
  cursor: pointer;
  font-size: 16px;
  color: var(--text-primary);
}
.back-btn:hover {
  background: var(--bg-hover);
}
.edit-btn {
  padding: 8px 14px;
  border: none;
  border-radius: 4px;
  background: var(--accent-blue);
  color: var(--text-on-accent);
  cursor: pointer;
  font-size: 13px;
  white-space: nowrap;
}
.edit-btn:hover {
  background: var(--accent-blue-hover);
}

/* Image gallery */
.image-gallery {
  margin-bottom: 20px;
}
.gallery-main {
  position: relative;
  background: var(--bg-secondary);
  border-radius: 8px;
  overflow: hidden;
  aspect-ratio: 4/3;
  display: flex;
  align-items: center;
  justify-content: center;
}
.gallery-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  cursor: zoom-in;
}
.gallery-nav {
  position: absolute;
  bottom: 12px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--overlay-bg);
  padding: 4px 12px;
  border-radius: 20px;
}
.nav-btn {
  background: none;
  border: none;
  color: var(--text-on-accent);
  font-size: 20px;
  cursor: pointer;
  padding: 0 4px;
}
.nav-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}
.image-counter {
  color: rgba(255, 255, 255, 0.8);
  font-size: 13px;
  font-variant-numeric: tabular-nums;
}
.gallery-thumbs {
  display: flex;
  gap: 6px;
  margin-top: 8px;
  overflow-x: auto;
  padding: 4px 0;
}
.thumb {
  width: 56px;
  height: 56px;
  object-fit: cover;
  border-radius: 4px;
  cursor: pointer;
  border: 2px solid transparent;
  opacity: 0.6;
  transition: opacity 0.15s, border-color 0.15s;
}
.thumb:hover {
  opacity: 0.9;
}
.thumb.active {
  opacity: 1;
  border-color: var(--accent-blue);
}

.no-images {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 32px;
  color: var(--text-muted);
  background: var(--bg-secondary);
  border-radius: 8px;
  margin-bottom: 20px;
}
.no-images svg {
  opacity: 0.4;
  margin-bottom: 8px;
}
.no-images p {
  margin: 0;
  font-size: 13px;
}

/* Info sections */
.detail-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.info-section h3 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.info-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  border-bottom: 1px solid var(--border-primary);
  font-size: 14px;
}
.info-label {
  color: var(--text-secondary);
}
.info-value {
  font-weight: 500;
  text-align: right;
}
</style>
