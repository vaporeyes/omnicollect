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
  {label: 'Open Settings', keywords: ['settings', 'preferences'], action: 'openSettings'},
  {label: 'Export Backup', keywords: ['backup', 'export'], action: 'exportBackup'},
]

// Filter quick actions by keyword match
const matchedActions = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return []
  return quickActions.filter(a => a.keywords.some(k => k.includes(q) || q.includes(k)))
})

// Combined list for keyboard navigation: quick actions first, then items
const totalCount = computed(() => matchedActions.value.length + results.value.length)

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
  }
}

function scrollHighlightedIntoView() {
  nextTick(() => {
    const el = resultsEl.value?.querySelector('.highlighted')
    if (el) (el as HTMLElement).scrollIntoView({block: 'nearest'})
  })
}

function selectHighlighted() {
  const actionCount = matchedActions.value.length
  const idx = highlightedIndex.value
  if (idx < actionCount) {
    emit('action', matchedActions.value[idx].action)
    emit('close')
  } else {
    const itemIdx = idx - actionCount
    if (itemIdx < results.value.length) {
      emit('selectItem', results.value[itemIdx])
      emit('close')
    }
  }
}

function selectAction(action: string) {
  emit('action', action)
  emit('close')
}

function selectItem(item: Item) {
  emit('selectItem', item)
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
              placeholder="Search items, actions..."
              spellcheck="false"
              autocomplete="off"
            />
            <kbd class="palette-hint">esc</kbd>
          </div>

          <div ref="resultsEl" class="palette-results" v-if="query.trim()">
            <!-- Quick actions -->
            <div v-if="matchedActions.length > 0" class="results-group">
              <div class="group-label">Actions</div>
              <div
                v-for="(act, i) in matchedActions"
                :key="'action-' + act.action"
                :class="['result-row', 'result-action', {highlighted: highlightedIndex === i}]"
                @click="selectAction(act.action)"
                @mouseenter="highlightedIndex = i"
              >
                <svg class="action-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                  <path d="M13 10V3L4 14h7v7l9-11h-7z"/>
                </svg>
                <span class="result-title">{{ act.label }}</span>
              </div>
            </div>

            <!-- Item results -->
            <div v-if="results.length > 0" class="results-group">
              <div class="group-label">Items</div>
              <div
                v-for="(item, i) in results"
                :key="item.id"
                :class="['result-row', 'result-item', {highlighted: highlightedIndex === matchedActions.length + i}]"
                @click="selectItem(item)"
                @mouseenter="highlightedIndex = matchedActions.length + i"
              >
                <img
                  v-if="item.images && item.images.length > 0"
                  :src="'/thumbnails/' + encodeURIComponent(item.images[0])"
                  class="result-thumb"
                  alt=""
                />
                <div v-else class="result-thumb result-thumb-placeholder">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <rect x="3" y="3" width="18" height="18" rx="2"/>
                    <circle cx="8.5" cy="8.5" r="1.5"/>
                    <path d="M21 15l-5-5L5 21"/>
                  </svg>
                </div>
                <span class="result-title">{{ item.title }}</span>
                <span class="result-badge">{{ moduleName(item.moduleId) }}</span>
              </div>
            </div>

            <!-- No results -->
            <div v-if="matchedActions.length === 0 && results.length === 0" class="no-results">
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
  padding-top: 12vh;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
}
.palette-dialog {
  width: 100%;
  max-width: 580px;
  background: var(--bg-secondary, rgba(30, 30, 46, 0.85));
  backdrop-filter: blur(var(--glass-blur, 20px));
  -webkit-backdrop-filter: blur(var(--glass-blur, 20px));
  border: 1px solid var(--border-primary, rgba(255,255,255,0.08));
  border-radius: var(--radius-lg, 16px);
  box-shadow: var(--shadow-lg, 0 16px 48px rgba(0,0,0,0.3));
  overflow: hidden;
}

/* Input area */
.palette-input-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 18px;
  border-bottom: 1px solid var(--border-primary);
}
.search-icon {
  flex-shrink: 0;
  color: var(--text-muted);
  opacity: 0.6;
}
.palette-input {
  flex: 1;
  border: none;
  background: transparent;
  font-family: 'Instrument Serif', serif;
  font-size: 20px;
  color: var(--text-primary);
  outline: none;
  letter-spacing: -0.01em;
}
.palette-input::placeholder {
  color: var(--text-muted);
  opacity: 0.5;
}
.palette-hint {
  flex-shrink: 0;
  font-family: 'Outfit', sans-serif;
  font-size: 10px;
  font-weight: 500;
  color: var(--text-muted);
  background: var(--bg-hover, rgba(255,255,255,0.06));
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid var(--border-primary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* Results area */
.palette-results {
  max-height: 380px;
  overflow-y: auto;
  padding: 6px;
}
.results-group {
  margin-bottom: 4px;
}
.group-label {
  font-family: 'Outfit', sans-serif;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide, 0.08em);
  color: var(--text-muted);
  padding: 6px 10px 4px;
}

/* Result rows */
.result-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  transition: background 0.08s;
}
.result-row:hover,
.result-row.highlighted {
  background: var(--bg-hover, rgba(255,255,255,0.06));
}

/* Action rows */
.action-icon {
  flex-shrink: 0;
  color: var(--accent-blue);
  opacity: 0.8;
}

/* Item rows */
.result-thumb {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm, 4px);
  object-fit: cover;
  flex-shrink: 0;
}
.result-thumb-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary, rgba(255,255,255,0.04));
  color: var(--text-muted);
  opacity: 0.4;
}
.result-title {
  flex: 1;
  font-family: 'Outfit', sans-serif;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.result-badge {
  flex-shrink: 0;
  font-family: 'Outfit', sans-serif;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide, 0.08em);
  color: var(--accent-blue);
  background: var(--accent-blue-light, rgba(59,130,246,0.1));
  padding: 2px 8px;
  border-radius: 10px;
}

.no-results {
  padding: 24px;
  text-align: center;
  font-family: 'Outfit', sans-serif;
  font-size: 13px;
  color: var(--text-muted);
}

/* Transition */
.palette-enter-active {
  transition: opacity 0.15s ease-out;
}
.palette-enter-active .palette-dialog {
  transition: transform 0.15s ease-out, opacity 0.15s ease-out;
}
.palette-leave-active {
  transition: opacity 0.1s ease-in;
}
.palette-leave-active .palette-dialog {
  transition: transform 0.1s ease-in, opacity 0.1s ease-in;
}
.palette-enter-from {
  opacity: 0;
}
.palette-enter-from .palette-dialog {
  opacity: 0;
  transform: scale(0.97) translateY(-8px);
}
.palette-leave-to {
  opacity: 0;
}
.palette-leave-to .palette-dialog {
  opacity: 0;
  transform: scale(0.97) translateY(-8px);
}
</style>
