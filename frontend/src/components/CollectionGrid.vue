<script lang="ts" setup>
import {main} from '../../wailsjs/go/models'

const props = defineProps<{
  items: main.Item[]
  modules: main.ModuleSchema[]
}>()

const emit = defineEmits<{
  select: [item: main.Item]
  viewImage: [item: main.Item, filename: string]
  addItem: []
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
      <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round">
        <rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/>
        <rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/>
      </svg>
      <p>Your collection is empty</p>
      <p class="empty-hint">Select a collection type from the sidebar to add your first item.</p>
      <button v-if="modules.length > 0" class="cta-btn" @click="emit('addItem')">Add First Item</button>
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
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: box-shadow 0.15s;
}
.grid-card:hover {
  box-shadow: var(--shadow-sm);
}
.card-image {
  aspect-ratio: 1;
  background: var(--bg-secondary);
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
  background: var(--bg-tertiary);
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
  color: var(--text-secondary);
}
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 48px 24px;
  color: var(--text-muted);
}
.empty-state svg {
  margin-bottom: 12px;
  opacity: 0.4;
}
.empty-state p {
  margin: 0 0 4px 0;
  font-size: 15px;
}
.empty-hint {
  font-size: 13px !important;
  margin-bottom: 16px !important;
}
.cta-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  background: var(--accent-blue);
  color: var(--text-on-accent);
  cursor: pointer;
  font-size: 14px;
}
.cta-btn:hover {
  background: var(--accent-blue-hover);
}
</style>
