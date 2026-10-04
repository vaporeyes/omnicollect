<script lang="ts" setup>
import {ref, computed} from 'vue'
import type {Item, ModuleSchema} from '../api/types'
import {useSelectionStore} from '../stores/selectionStore'
import {menuAnchor} from '../composables/menuAnchor'

const selectionStore = useSelectionStore()
function openActions(event: MouseEvent, item: Item) {
  const {x, y} = menuAnchor(event)
  emit('itemContextMenu', item, x, y)
}

const props = defineProps<{
  items: Item[]
  modules: ModuleSchema[]
  unavailable?: boolean
  filtered?: boolean
  activeModuleId?: string
  searchQuery?: string
}>()

const emit = defineEmits<{
  select: [item: Item]
  filterChange: [moduleId: string]
  search: [query: string]
  addItem: []
  itemContextMenu: [item: Item, x: number, y: number]
}>()

const searchText = computed({get: () => props.searchQuery ?? '', set: value => emit('search', value)})
const searchInputEl = ref<HTMLInputElement | null>(null)

function focusSearch() {
  searchInputEl.value?.focus()
}

defineExpose({focusSearch})


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

function sortOrder(key: string): 'none' | 'ascending' | 'descending' {return sortKey.value === key ? sortDir.value === 'asc' ? 'ascending' : 'descending' : 'none'}
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
    } else if (key === 'moduleId') {
      va = moduleName(a.moduleId); vb = moduleName(b.moduleId)
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
      <select class="filter-select" aria-label="Collection type" :value="activeModuleId ?? ''" @change="emit('filterChange', ($event.target as HTMLSelectElement).value)">
        <option value="">All Types</option>
        <option v-for="mod in modules" :key="mod.id" :value="mod.id">
          {{ mod.displayName }}
        </option>
      </select>
      <input
        ref="searchInputEl"
        type="text"
        v-model="searchText"
        placeholder="Search..."
        aria-label="Search collection"
        class="search-input"
      />
    </div>

    <p v-if="unavailable" role="status">Results are unavailable while loading or after an error. Use the retry controls above.</p>
    <div v-else-if="sortedItems.length === 0" class="empty-state">
      <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round">
        <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
        <polyline points="14 2 14 8 20 8"/>
        <line x1="12" y1="12" x2="12" y2="18"/><line x1="9" y1="15" x2="15" y2="15"/>
      </svg>
      <p>{{ filtered ? 'No items match this search or filter.' : 'No items in the loaded collection.' }}</p>
      <p class="empty-hint">{{ filtered ? 'Change the search, collection type, tags or filters to see other items.' : 'Select a collection type to add an item.' }}</p>
      <button v-if="!filtered && modules.length > 0" class="cta-btn" @click="emit('addItem')">Add First Item</button>
    </div>

    <div v-else class="table-wrapper">
      <table class="data-table">
        <thead>
          <tr>
            <th class="col-check">
              <input
                type="checkbox"
                aria-label="Select all displayed items"
                :checked="allSelected"
                :indeterminate="someSelected"
                @change="toggleSelectAll"
                @click.stop
              />
            </th>
            <th class="sortable" :aria-sort="sortOrder('title')">
              <button class="sort-control" @click="toggleSort('title')">Title{{ sortIndicator('title') }}</button>
            </th>
            <th v-if="!activeSchema" class="sortable" :aria-sort="sortOrder('moduleId')">
              <button class="sort-control" @click="toggleSort('moduleId')">Type{{ sortIndicator('moduleId') }}</button>
            </th>
            <th class="sortable col-price" :aria-sort="sortOrder('purchasePrice')">
              <button class="sort-control" @click="toggleSort('purchasePrice')">Price{{ sortIndicator('purchasePrice') }}</button>
            </th>
            <!-- Dynamic columns from active module schema -->
            <th
              v-for="col in dynamicColumns"
              :key="col.key"
              class="sortable"
              :aria-sort="sortOrder(col.key)"
            >
              <button class="sort-control" @click="toggleSort(col.key)">{{ col.label }}{{ sortIndicator(col.key) }}</button>
            </th>
            <th class="sortable col-date" :aria-sort="sortOrder('updatedAt')">
              <button class="sort-control" @click="toggleSort('updatedAt')">Modified{{ sortIndicator('updatedAt') }}</button>
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
            @contextmenu="openActions($event, item)"
          >
            <td class="col-check" @click.stop>
              <input
                type="checkbox"
                :aria-label="`Select ${item.title}`"
                :checked="selectionStore.isSelected(item.id)"
                @click="onCheckboxClick($event as MouseEvent, item, index)"
              />
            </td>
            <td class="col-title"><button class="item-open" @click.stop="emit('select', item)">{{ item.title }}</button><button class="item-actions" :aria-label="`Actions for ${item.title}`" aria-haspopup="menu" @click="openActions($event, item)">⋯</button></td>
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
.sort-control {background: transparent; border: 0; color: inherit; font: inherit; text-align: inherit; cursor: pointer; padding: 0; width: 100%;}
.item-open, .item-actions {border: 0; background: transparent; color: inherit; font: inherit; cursor: pointer; text-align: left;}
.item-actions {margin-left: 12px; padding: 4px 8px;}
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
