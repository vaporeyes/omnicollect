<script lang="ts" setup>
import {ref, onMounted} from 'vue'
import {main} from '../wailsjs/go/models'
import {useModuleStore} from './stores/moduleStore'
import {useCollectionStore} from './stores/collectionStore'
import ModuleSelector from './components/ModuleSelector.vue'
import DynamicForm from './components/DynamicForm.vue'
import ItemList from './components/ItemList.vue'

const moduleStore = useModuleStore()
const collectionStore = useCollectionStore()

const selectedSchema = ref<main.ModuleSchema | null>(null)
const editingItem = ref<main.Item | null>(null)
const showForm = ref(false)

onMounted(async () => {
  await Promise.all([
    moduleStore.fetchModules(),
    collectionStore.fetchItems(),
  ])
})

function onModuleSelect(mod: main.ModuleSchema) {
  selectedSchema.value = mod
  editingItem.value = null
  showForm.value = true
}

function onItemSelect(item: main.Item) {
  const schema = moduleStore.getModuleById(item.moduleId)
  if (!schema) {
    alert(`Collection type schema not available for "${item.moduleId}". The schema file may have been removed.`)
    return
  }
  selectedSchema.value = schema
  editingItem.value = item
  showForm.value = true
}

async function onSave(item: main.Item) {
  try {
    await collectionStore.saveItem(item)
    showForm.value = false
    editingItem.value = null
  } catch {
    // Error is already set in collectionStore.error
  }
}

function onCancel() {
  showForm.value = false
  editingItem.value = null
}

function onFilterChange(moduleId: string) {
  collectionStore.setFilter(moduleId)
}

function onSearch(query: string) {
  collectionStore.setSearch(query)
}
</script>

<template>
  <div class="app-layout">
    <aside class="sidebar">
      <h2>OmniCollect</h2>
      <ModuleSelector
        :modules="moduleStore.modules"
        @select="onModuleSelect"
      />
    </aside>

    <main class="main-content">
      <div v-if="moduleStore.loading || collectionStore.loading" class="loading">
        Loading...
      </div>

      <div v-if="collectionStore.error" class="error-message">
        {{ collectionStore.error }}
      </div>

      <DynamicForm
        v-if="showForm && selectedSchema"
        :schema="selectedSchema"
        :item="editingItem"
        @save="onSave"
        @cancel="onCancel"
      />

      <ItemList
        v-if="!showForm"
        :items="collectionStore.items"
        :modules="moduleStore.modules"
        @select="onItemSelect"
        @filterChange="onFilterChange"
        @search="onSearch"
      />
    </main>
  </div>
</template>

<style>
body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  color: #333;
}
.app-layout {
  display: flex;
  height: 100vh;
}
.sidebar {
  width: 240px;
  padding: 16px;
  border-right: 1px solid #e2e8f0;
  background: #f7fafc;
  overflow-y: auto;
}
.sidebar h2 {
  margin: 0 0 16px 0;
  font-size: 18px;
}
.main-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}
.loading {
  color: #888;
  padding: 24px;
  text-align: center;
}
.error-message {
  background: #fed7d7;
  color: #c53030;
  padding: 8px 12px;
  border-radius: 4px;
  margin-bottom: 12px;
  font-size: 14px;
}
</style>
