<!-- ABOUTME: Spotlight-style command palette overlay for global search and quick actions. -->
<!-- ABOUTME: Triggered by Cmd/Ctrl+K, searches items across all modules with keyboard navigation. -->
<script lang="ts" setup>
import {ref, computed, watch, nextTick, onMounted, onUnmounted} from 'vue'
import type {Item} from '../api/types'
import {useCollectionStore} from '../stores/collectionStore'
import {useModuleStore} from '../stores/moduleStore'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  close: []
  selectItem: [item: Item]
  action: [action: string]
}>()

const collectionStore = useCollectionStore()
const moduleStore = useModuleStore()

const query = ref('')
const results = ref<Item[]>([])
const highlightedIndex = ref(0)
const inputEl = ref<HTMLInputElement | null>(null)
const resultsEl = ref<HTMLElement | null>(null)
let debounceTimer: ReturnType<typeof setTimeout> | null = null

// Quick actions with keyword triggers
const quickActions = [
  {label: 'Add New Item', keywords: ['new', 'add', 'create'], action: 'newItem'},
  {label: 'Create New Schema', keywords: ['new', 'schema', 'create'], action: 'newSchema'},
  {label: 'Manage Tags', keywords: ['tags', 'tag', 'manage', 'rename', 'delete'], action: 'manageTags'},
  {label: 'Open Settings', keywords: ['settings', 'preferences'], action: 'openSettings'},
  {label: 'Export Backup', keywords: ['backup', 'export'], action: 'exportBackup'},
  {label: 'Import Backup', keywords: ['backup', 'import', 'restore'], action: 'importBackup'},
]

// Filter quick actions by keyword match
const matchedActions = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return []
  return quickActions.filter(a => a.keywords.some(k => k.includes(q) || q.includes(k)))
})

const combinedResults = computed(() => {
  const arr: any[] = []
  for (const act of matchedActions.value) {
    arr.push({ type: 'action', id: act.action, title: act.label, action: act.action })
  }
  for (const item of results.value) {
    arr.push({ 
      type: 'item', 
      id: item.id, 
      title: item.title, 
      moduleName: moduleName(item.moduleId),
      thumbnail: item.images && item.images.length > 0 ? item.images[0] : null,
      rawItem: item
    })
  }
  return arr
})

const totalCount = computed(() => combinedResults.value.length)

// Reset state when palette opens/closes
watch(() => props.visible, (vis) => {
  if (vis) {
    query.value = ''
    results.value = []
    highlightedIndex.value = 0
    nextTick(() => inputEl.value?.focus())
  }
})

// Debounced search
watch(query, (q) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  const trimmed = q.trim()
  if (!trimmed) {
    results.value = []
    highlightedIndex.value = 0
    return
  }
  debounceTimer = setTimeout(async () => {
    const all = await collectionStore.searchAllItems(trimmed)
    results.value = all.slice(0, 25)
    highlightedIndex.value = 0
  }, 200)
})

function moduleName(moduleId: string): string {
  return moduleStore.getModuleById(moduleId)?.displayName ?? moduleId
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    if (highlightedIndex.value < totalCount.value - 1) {
      highlightedIndex.value++
      scrollHighlightedIntoView()
    }
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    if (highlightedIndex.value > 0) {
      highlightedIndex.value--
      scrollHighlightedIntoView()
    }
  } else if (e.key === 'Enter') {
    e.preventDefault()
    selectHighlighted()
  } else if (e.key === 'Escape') {
    emit('close')
  }
}

function scrollHighlightedIntoView() {
  nextTick(() => {
    const el = resultsEl.value?.querySelector('.highlighted')
    if (el) (el as HTMLElement).scrollIntoView({block: 'nearest'})
  })
}

function selectHighlighted() {
  const result = combinedResults.value[highlightedIndex.value]
  if (!result) return
  executeSelection(result)
}

function executeSelection(result: any) {
  if (result.type === 'action') {
    emit('action', result.action)
  } else if (result.type === 'item') {
    emit('selectItem', result.rawItem)
  }
  emit('close')
}

function onBackdropClick() {
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <Transition name="palette">
      <div v-if="visible" class="palette-overlay" @mousedown.self="onBackdropClick">
        <div class="palette-dialog" @keydown="onKeydown">
          <div class="palette-input-wrap">
            <svg class="search-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
            </svg>
            <input
              ref="inputEl"
              v-model="query"
              type="text"
              class="palette-input"
              placeholder="Search items, commands, or jump to..."
              spellcheck="false"
              autocomplete="off"
            />
            <span class="palette-hint">esc</span>
          </div>

          <div ref="resultsEl" class="palette-results" v-if="query.trim() || matchedActions.length > 0">
            <!-- Unified List -->
            <ul class="results-list" role="listbox">
              <li
                v-for="(result, index) in combinedResults"
                :key="result.type + '-' + result.id"
                :class="[
                  'result-row',
                  { 'highlighted': highlightedIndex === index }
                ]"
                @click="executeSelection(result)"
                @mouseenter="highlightedIndex = index"
              >
                <!-- Icon/Thumbnail -->
                <img
                  v-if="result.type === 'item' && result.thumbnail"
                  :src="'/thumbnails/' + encodeURIComponent(result.thumbnail)"
                  class="result-thumb"
                  alt=""
                />
                <div v-else-if="result.type === 'item'" class="result-thumb result-thumb-placeholder">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <rect x="3" y="3" width="18" height="18" rx="2"/>
                    <circle cx="8.5" cy="8.5" r="1.5"/>
                    <path d="M21 15l-5-5L5 21"/>
                  </svg>
                </div>
                <div v-else class="action-icon-wrap">
                  <svg class="action-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                    <path d="M13 10V3L4 14h7v7l9-11h-7z"/>
                  </svg>
                </div>
                
                <!-- Text Content -->
                <div class="result-content">
                  <span class="result-title">{{ result.title }}</span>
                  <span v-if="result.moduleName" class="result-subtitle">{{ result.moduleName }}</span>
                </div>
                
                <!-- Active Indicator -->
                <span v-if="highlightedIndex === index" class="result-active-icon">↵</span>
              </li>
            </ul>

            <!-- No results -->
            <div v-if="combinedResults.length === 0" class="no-results">
              No results for "{{ query.trim() }}"
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.palette-overlay {
  position: fixed;
  inset: 0;
  z-index: 3000;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 15vh;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(var(--glass-blur, 16px));
  -webkit-backdrop-filter: blur(var(--glass-blur, 16px));
}

.palette-dialog {
  width: 100%;
  max-width: 640px;
  background: var(--bg-secondary, rgba(30, 30, 46, 0.85));
  border: 1px solid var(--border-primary, rgba(255,255,255,0.08));
  border-radius: var(--radius-lg, 20px);
  box-shadow: var(--shadow-lg, 0 16px 48px rgba(0,0,0,0.3));
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* Input area */
.palette-input-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-primary);
}

.search-icon {
  flex-shrink: 0;
  color: var(--text-muted);
}

.palette-input {
  flex: 1;
  border: none;
  background: transparent;
  font-family: var(--font-body);
  font-size: 18px;
  color: var(--text-primary);
  outline: none;
}

.palette-input::placeholder {
  color: var(--text-muted);
  opacity: 0.7;
}

.palette-hint {
  flex-shrink: 0;
  font-family: var(--font-body);
  font-size: 11px;
  font-weight: 500;
  color: var(--text-muted);
  background: var(--bg-tertiary);
  padding: 3px 8px;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* Results area */
.palette-results {
  max-height: 60vh;
  overflow-y: auto;
  padding: 8px;
}

.results-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

/* Result rows */
.result-row {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  border-radius: var(--radius-md, 12px);
  cursor: pointer;
  transition: background var(--transition-fast), color var(--transition-fast);
  color: var(--text-primary);
}

.result-row:hover,
.result-row.highlighted {
  background: var(--accent-blue-light);
  color: var(--accent-blue);
}

/* Icons / Thumbs */
.action-icon-wrap {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm, 6px);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  background: transparent;
}
.action-icon {
  color: currentColor;
}

.result-thumb {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm, 6px);
  object-fit: cover;
  flex-shrink: 0;
  margin-right: 16px;
}

.result-thumb-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  color: var(--text-muted);
}

.result-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.result-title {
  font-family: var(--font-body);
  font-size: 15px;
  font-weight: 500;
}

.result-subtitle {
  font-family: var(--font-body);
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 2px;
}

.result-row.highlighted .result-subtitle {
  color: currentColor;
  opacity: 0.7;
}

.result-active-icon {
  font-size: 14px;
  opacity: 0.75;
  color: currentColor;
}

.no-results {
  padding: 24px;
  text-align: center;
  font-family: var(--font-body);
  font-size: 14px;
  color: var(--text-muted);
}

/* Transition */
.palette-enter-active {
  transition: opacity 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}
.palette-enter-active .palette-dialog {
  transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.2s ease-out;
}
.palette-leave-active {
  transition: opacity 0.15s ease-in;
}
.palette-leave-active .palette-dialog {
  transition: transform 0.15s ease-in, opacity 0.15s ease-in;
}
.palette-enter-from,
.palette-leave-to {
  opacity: 0;
}
.palette-enter-from .palette-dialog,
.palette-leave-to .palette-dialog {
  transform: scale(0.95) translateY(10px);
}
</style>
