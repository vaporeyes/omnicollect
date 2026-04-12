<script lang="ts" setup>
import type {Item, ModuleSchema} from '../api/types'
import {useSelectionStore} from '../stores/selectionStore'

const selectionStore = useSelectionStore()

const props = defineProps<{
  items: Item[]
  modules: ModuleSchema[]
}>()

const emit = defineEmits<{
  select: [item: Item]
  viewImage: [item: Item, filename: string]
  addItem: []
  itemContextMenu: [item: Item, x: number, y: number]
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

function onSelectClick(event: MouseEvent, item: Item, index: number) {
  event.stopPropagation()
  if (event.shiftKey) {
    selectionStore.shiftSelect(index, props.items)
  } else {
    selectionStore.toggle(item.id, index)
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
        v-for="(item, index) in items"
        :key="item.id"
        :class="['grid-card', {'card-selected': selectionStore.isSelected(item.id)}]"
        :style="{ '--card-index': index }"
        @click="emit('select', item)"
        @contextmenu.prevent="emit('itemContextMenu', item, $event.clientX, $event.clientY)"
      >
        <div class="card-image">
          <div
            :class="['select-badge', {active: selectionStore.isSelected(item.id)}]"
            @click.stop="onSelectClick($event as MouseEvent, item, index)"
          >
            <svg v-if="selectionStore.isSelected(item.id)" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round"><polyline points="20 6 9 17 4 12"/></svg>
          </div>
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
          <div class="card-caption">
            <span class="card-title">{{ item.title }}</span>
            <span class="card-meta">
              <span class="card-module">{{ moduleName(item.moduleId) }}</span>
              <span class="card-date">{{ formatDate(item.updatedAt) }}</span>
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.grid {
  column-width: 240px;
  column-gap: 16px;
}
.grid-card {
  break-inside: avoid;
  margin-bottom: 16px;
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.15s cubic-bezier(0, 0.55, 0.45, 1), box-shadow 0.15s;
  box-shadow: var(--shadow-sm);
  /* Drawer pull entrance */
  clip-path: polygon(0 0, 100% 0, 100% 0, 0 0);
  animation: unroll 0.6s cubic-bezier(0.77, 0, 0.17, 1) forwards;
  animation-delay: calc(var(--card-index, 0) * 40ms);
}
@keyframes unroll {
  to { clip-path: polygon(0 0, 100% 0, 100% 100%, 0 100%); }
}
.grid-card:hover {
  transform: translate(-4px, -4px);
  box-shadow: 4px 4px 0px var(--accent-blue);
}
.card-selected {
  outline: 2px solid var(--accent-blue);
  outline-offset: -2px;
}
.select-badge {
  position: absolute;
  top: 8px;
  left: 8px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 2px solid rgba(255,255,255,0.7);
  background: rgba(0,0,0,0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 2;
  opacity: 0;
  transition: opacity 0.15s, background 0.15s;
}
.grid-card:hover .select-badge,
.select-badge.active {
  opacity: 1;
}
.select-badge.active {
  background: var(--accent-blue);
  border-color: var(--accent-blue);
  color: #fff;
}
.card-image {
  position: relative;
  background: var(--bg-secondary);
  overflow: hidden;
}
/* Scanner sweep reveal on image load */
.card-image::after {
  content: "";
  position: absolute;
  inset: 0;
  background: var(--accent-blue);
  transform: translateX(-100%);
  animation: scanner-sweep 0.5s cubic-bezier(0.77, 0, 0.17, 1) forwards;
  animation-delay: calc(var(--card-index, 0) * 40ms + 0.3s);
  pointer-events: none;
  z-index: 1;
}
@keyframes scanner-sweep {
  0%   { transform: translateX(-100%); }
  50%  { transform: translateX(0); }
  100% { transform: translateX(100%); }
}
.card-image img {
  width: 100%;
  height: auto;
  display: block;
}
.placeholder {
  width: 100%;
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
}
.card-caption {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 8px 10px;
  background: hsla(0, 0%, 10%, 0.45);
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.card-title {
  font-weight: 600;
  font-size: 13px;
  color: #fff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: var(--leading-tight);
}
.card-meta {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: hsla(0, 0%, 100%, 0.7);
  line-height: var(--leading-dense);
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
  border-radius: var(--radius-md);
  background: var(--accent-blue);
  color: var(--text-on-accent);
  cursor: pointer;
  font-size: 14px;
}
.cta-btn:hover {
  background: var(--accent-blue-hover);
}
</style>
