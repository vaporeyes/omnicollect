<script lang="ts" setup>
import {main} from '../../wailsjs/go/models'

const props = defineProps<{
  items: main.Item[]
  modules: main.ModuleSchema[]
}>()

const emit = defineEmits<{
  select: [item: main.Item]
  viewImage: [item: main.Item, filename: string]
}>()

function moduleName(moduleId: string): string {
  const mod = props.modules.find(m => m.id === moduleId)
  return mod?.displayName ?? moduleId
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  try {
    return new Date(dateStr).toLocaleDateString()
  } catch {
    return dateStr
  }
}

function onImageError(event: Event) {
  const img = event.target as HTMLImageElement
  img.style.display = 'none'
  const placeholder = img.nextElementSibling as HTMLElement
  if (placeholder) placeholder.style.display = 'flex'
}
</script>

<template>
  <div class="collection-grid">
    <div v-if="items.length === 0" class="empty-state">
      No items found.
    </div>

    <div v-else class="grid">
      <div
        v-for="item in items"
        :key="item.id"
        class="grid-card"
        @click="emit('select', item)"
      >
        <div class="card-image">
          <img
            v-if="item.images && item.images.length > 0"
            :src="'/thumbnails/' + encodeURIComponent(item.images[0])"
            loading="lazy"
            alt=""
            @error="onImageError"
            @click.stop="emit('viewImage', item, item.images[0])"
          />
          <!-- Placeholder shown when no image or image fails to load -->
          <div
            class="placeholder"
            :style="{display: (!item.images || item.images.length === 0) ? 'flex' : 'none'}"
          >
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#ccc" stroke-width="1.5">
              <rect x="3" y="3" width="18" height="18" rx="2"/>
              <circle cx="8.5" cy="8.5" r="1.5"/>
              <path d="M21 15l-5-5L5 21"/>
            </svg>
          </div>
        </div>
        <div class="card-title">{{ item.title }}</div>
        <div class="card-meta">
          <span class="card-module">{{ moduleName(item.moduleId) }}</span>
          <span class="card-date">{{ formatDate(item.updatedAt) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}
.grid-card {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: box-shadow 0.15s;
}
.grid-card:hover {
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}
.card-image {
  aspect-ratio: 1;
  background: #f7fafc;
  overflow: hidden;
}
.card-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f0f0;
}
.card-title {
  padding: 8px 10px 2px;
  font-weight: 600;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.card-meta {
  display: flex;
  justify-content: space-between;
  padding: 0 10px 8px;
  font-size: 12px;
  color: #666;
}
.empty-state {
  font-size: 13px;
  color: #888;
  padding: 24px;
  text-align: center;
}
</style>
