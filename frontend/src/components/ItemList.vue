<script lang="ts" setup>
import {ref} from 'vue'
import {main} from '../../wailsjs/go/models'

const props = defineProps<{
  items: main.Item[]
  modules: main.ModuleSchema[]
}>()

const emit = defineEmits<{
  select: [item: main.Item]
  filterChange: [moduleId: string]
  search: [query: string]
}>()

const searchText = ref('')
let debounceTimer: ReturnType<typeof setTimeout> | null = null

function onSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    emit('search', searchText.value)
  }, 300)
}

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

    <div v-if="items.length === 0" class="empty-state">
      No items found.
    </div>

    <ul v-else class="items">
      <li
        v-for="item in items"
        :key="item.id"
        class="item-row"
        @click="emit('select', item)"
      >
        <div class="item-title">{{ item.title }}</div>
        <div class="item-meta">
          <span class="item-module">{{ moduleName(item.moduleId) }}</span>
          <span class="item-date">{{ formatDate(item.updatedAt) }}</span>
        </div>
      </li>
    </ul>
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
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 14px;
}
.search-input {
  flex: 1;
}
.items {
  list-style: none;
  padding: 0;
  margin: 0;
}
.item-row {
  padding: 10px 12px;
  border-bottom: 1px solid #eee;
  cursor: pointer;
}
.item-row:hover {
  background: rgba(0,0,0,0.03);
}
.item-title {
  font-weight: 600;
  font-size: 14px;
}
.item-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #666;
  margin-top: 2px;
}
.empty-state {
  font-size: 13px;
  color: #888;
  padding: 24px;
  text-align: center;
}
</style>
