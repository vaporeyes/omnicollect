<script lang="ts" setup>
import {ref, computed} from 'vue'
import {main} from '../../wailsjs/go/models'

const props = defineProps<{
  items: main.Item[]
  modules: main.ModuleSchema[]
  activeModuleId?: string
}>()

const emit = defineEmits<{
  select: [item: main.Item]
  filterChange: [moduleId: string]
  search: [query: string]
  addItem: []
}>()

const searchText = ref('')
let debounceTimer: ReturnType<typeof setTimeout> | null = null

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
            v-for="item in sortedItems"
            :key="item.id"
            class="data-row"
            @click="emit('select', item)"
          >
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
  padding: 6px 8px;
  border: 1px solid var(--border-input);
  border-radius: 4px;
  font-size: 14px;
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
}
.data-table th {
  text-align: left;
  padding: 8px 10px;
  border-bottom: 2px solid var(--border-primary);
  font-weight: 600;
  white-space: nowrap;
  background: var(--bg-secondary);
  position: sticky;
  top: 0;
}
.data-table th.sortable {
  cursor: pointer;
  user-select: none;
}
.data-table th.sortable:hover {
  background: var(--bg-hover);
}
.data-table td {
  padding: 8px 10px;
  border-bottom: 1px solid var(--border-primary);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.data-row {
  cursor: pointer;
}
.data-row:hover {
  background: var(--bg-hover);
}
.col-title {
  font-weight: 500;
}
.col-price {
  text-align: right;
  font-variant-numeric: tabular-nums;
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
