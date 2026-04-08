<script lang="ts" setup>
import {ref, computed} from 'vue'
import type {Item, ModuleSchema} from '../api/types'
import {useSelectionStore} from '../stores/selectionStore'

const selectionStore = useSelectionStore()

const props = defineProps<{
  items: Item[]
  modules: ModuleSchema[]
  activeModuleId?: string
}>()

const emit = defineEmits<{
  select: [item: Item]
  filterChange: [moduleId: string]
  search: [query: string]
  addItem: []
  itemContextMenu: [item: Item, x: number, y: number]
}>()

const searchText = ref('')
const searchInputEl = ref<HTMLInputElement | null>(null)
let debounceTimer: ReturnType<typeof setTimeout> | null = null

function focusSearch() {
  searchInputEl.value?.focus()
}

defineExpose({focusSearch})

function onSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    emit('search', searchText.value)
  }, 300)
}

// Resolve the active module schema for dynamic columns
const activeSchema = computed(() => {
  if (!props.activeModuleId) return null
  return props.modules.find(m => m.id === props.activeModuleId) ?? null
})

// Dynamic columns from the active schema's attributes
const dynamicColumns = computed(() => {
  if (!activeSchema.value) return []
  return activeSchema.value.attributes.map(attr => ({
    key: attr.name,
    label: attr.display?.label || attr.name,
    type: attr.type,
  }))
})

// Sorting state
const sortKey = ref<string>('')
const sortDir = ref<'asc' | 'desc'>('asc')

function toggleSort(key: string) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'asc'
  }
}

// Sort items locally
const sortedItems = computed(() => {
  if (!sortKey.value) return props.items

  const key = sortKey.value
  const dir = sortDir.value === 'asc' ? 1 : -1

  return [...props.items].sort((a, b) => {
    let va: any
    let vb: any

    // Base fields
    if (key === 'title') {
      va = a.title; vb = b.title
    } else if (key === 'purchasePrice') {
      va = a.purchasePrice ?? 0; vb = b.purchasePrice ?? 0
    } else if (key === 'updatedAt') {
      va = a.updatedAt; vb = b.updatedAt
    } else {
      // Dynamic attribute
      va = a.attributes?.[key] ?? ''
      vb = b.attributes?.[key] ?? ''
    }

    if (typeof va === 'number' && typeof vb === 'number') {
      return (va - vb) * dir
    }
    return String(va).localeCompare(String(vb)) * dir
  })
})

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

function formatCell(value: any, type: string): string {
  if (value === null || value === undefined) return ''
  if (type === 'boolean') return value ? 'Yes' : 'No'
  if (type === 'date') return formatDate(String(value))
  return String(value)
}

function formatPrice(price: number | null | undefined): string {
  if (price === null || price === undefined) return ''
  return price.toFixed(2)
}

const allSelected = computed(() =>
  sortedItems.value.length > 0 && sortedItems.value.every(i => selectionStore.isSelected(i.id))
)
const someSelected = computed(() =>
  sortedItems.value.some(i => selectionStore.isSelected(i.id)) && !allSelected.value
)

function onCheckboxClick(event: MouseEvent, item: Item, index: number) {
  event.stopPropagation()
  if (event.shiftKey) {
    selectionStore.shiftSelect(index, sortedItems.value)
  } else {
    selectionStore.toggle(item.id, index)
  }
}

function toggleSelectAll() {
  if (allSelected.value) {
    selectionStore.clear()
  } else {
    selectionStore.selectAll(sortedItems.value)
  }
}

function sortIndicator(key: string): string {
  if (sortKey.value !== key) return ''
  return sortDir.value === 'asc' ? ' \u25B2' : ' \u25BC'
}
</script>

<template>
  <div class="item-list">
    <div class="list-controls">
      <select class="filter-select" @change="emit('filterChange', ($event.target as HTMLSelectElement).value)">
        <option value="">All Types</option>
        <option v-for="mod in modules" :key="mod.id" :value="mod.id">
          {{ mod.displayName }}
        </option>
      </select>
      <input
        ref="searchInputEl"
        type="text"
        v-model="searchText"
        @input="onSearchInput"
        placeholder="Search..."
        class="search-input"
      />
    </div>

    <div v-if="sortedItems.length === 0" class="empty-state">
      <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round">
        <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
        <polyline points="14 2 14 8 20 8"/>
        <line x1="12" y1="12" x2="12" y2="18"/><line x1="9" y1="15" x2="15" y2="15"/>
      </svg>
      <p>No items found</p>
      <p class="empty-hint">Select a collection type from the sidebar to start adding items.</p>
      <button v-if="modules.length > 0" class="cta-btn" @click="emit('addItem')">Add First Item</button>
    </div>

    <div v-else class="table-wrapper">
      <table class="data-table">
        <thead>
          <tr>
            <th class="col-check">
              <input
                type="checkbox"
                :checked="allSelected"
                :indeterminate="someSelected"
                @change="toggleSelectAll"
                @click.stop
              />
            </th>
            <th class="sortable" @click="toggleSort('title')">
              Title{{ sortIndicator('title') }}
            </th>
            <th v-if="!activeSchema" class="sortable" @click="toggleSort('moduleId')">
              Type{{ sortIndicator('moduleId') }}
            </th>
            <th class="sortable col-price" @click="toggleSort('purchasePrice')">
              Price{{ sortIndicator('purchasePrice') }}
            </th>
            <!-- Dynamic columns from active module schema -->
            <th
              v-for="col in dynamicColumns"
              :key="col.key"
              class="sortable"
              @click="toggleSort(col.key)"
            >
              {{ col.label }}{{ sortIndicator(col.key) }}
            </th>
            <th class="sortable col-date" @click="toggleSort('updatedAt')">
              Modified{{ sortIndicator('updatedAt') }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(item, index) in sortedItems"
            :key="item.id"
            :class="['data-row', 'animate-fade-in', {selected: selectionStore.isSelected(item.id)}]"
            :style="{ animationDelay: `${index * 0.03}s` }"
            @click="emit('select', item)"
            @contextmenu.prevent="emit('itemContextMenu', item, $event.clientX, $event.clientY)"
          >
            <td class="col-check" @click.stop>
              <input
                type="checkbox"
                :checked="selectionStore.isSelected(item.id)"
                @click="onCheckboxClick($event as MouseEvent, item, index)"
              />
            </td>
            <td class="col-title">{{ item.title }}</td>
            <td v-if="!activeSchema">{{ moduleName(item.moduleId) }}</td>
            <td class="col-price">{{ formatPrice(item.purchasePrice) }}</td>
            <td v-for="col in dynamicColumns" :key="col.key">
              {{ formatCell(item.attributes?.[col.key], col.type) }}
            </td>
            <td class="col-date">{{ formatDate(item.updatedAt) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.item-list {
  flex: 1;
}
.list-controls {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
.filter-select, .search-input {
  padding: 6px 10px;
  border: 1px solid var(--border-input);
  border-radius: var(--radius-md);
  font-size: 13px;
}
.search-input {
  flex: 1;
}
.table-wrapper {
  overflow-x: auto;
}
.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  font-variant-numeric: tabular-nums;
  line-height: var(--leading-dense);
}
.data-table th {
  text-align: left;
  padding: 6px 10px;
  border-bottom: 1px solid var(--border-primary);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide);
  color: var(--text-muted);
  white-space: nowrap;
  background: var(--bg-secondary);
  -webkit-backdrop-filter: blur(var(--glass-blur));
  backdrop-filter: blur(var(--glass-blur));
  position: sticky;
  top: 0;
  z-index: 1;
}
.data-table th.sortable {
  cursor: pointer;
  user-select: none;
  transition: color var(--transition-fast);
}
.data-table th.sortable:hover {
  color: var(--text-primary);
}
.data-table td {
  padding: 7px 10px;
  border-bottom: 1px solid var(--border-primary);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.data-row {
  cursor: pointer;
  transition: background var(--transition-fast);
}
.data-row:hover {
  background: var(--bg-hover);
}
.col-check {
  width: 32px;
  text-align: center;
  padding: 0 6px !important;
}
.col-check input[type="checkbox"] {
  cursor: pointer;
  width: 15px;
  height: 15px;
}
.data-row.selected {
  background: var(--accent-blue-light, rgba(59,130,246,0.08));
}
.col-title {
  font-weight: 500;
}
.col-price {
  text-align: right;
}
.col-date {
  color: var(--text-secondary);
  white-space: nowrap;
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
