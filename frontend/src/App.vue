<script lang="ts" setup>
import {ref, onMounted, onUnmounted} from 'vue'
import {main} from '../wailsjs/go/models'
import {LoadModuleFile, ExportBackup} from '../wailsjs/go/main/App'
import {WindowSetSystemDefaultTheme} from '../wailsjs/runtime/runtime'
import {useModuleStore} from './stores/moduleStore'
import {useCollectionStore} from './stores/collectionStore'
import ModuleSelector from './components/ModuleSelector.vue'
import DynamicForm from './components/DynamicForm.vue'
import ItemList from './components/ItemList.vue'
import CollectionGrid from './components/CollectionGrid.vue'
import ImageLightbox from './components/ImageLightbox.vue'
import SchemaBuilder from './components/SchemaBuilder.vue'

const moduleStore = useModuleStore()
const collectionStore = useCollectionStore()

// Theme detection via system preference
function applyTheme(dark: boolean) {
  document.documentElement.classList.toggle('dark-theme', dark)
}

const darkMediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
applyTheme(darkMediaQuery.matches)

function onThemeChange(e: MediaQueryListEvent) {
  applyTheme(e.matches)
}
darkMediaQuery.addEventListener('change', onThemeChange)
onUnmounted(() => darkMediaQuery.removeEventListener('change', onThemeChange))

// Tell Wails to follow system theme for window chrome
try { WindowSetSystemDefaultTheme() } catch { /* ok if not available in dev */ }

const selectedSchema = ref<main.ModuleSchema | null>(null)
const editingItem = ref<main.Item | null>(null)
const showForm = ref(false)
const viewMode = ref<'list' | 'grid'>('grid')

// Lightbox state
const lightboxFilename = ref('')
const lightboxVisible = ref(false)

// Schema builder state
const showBuilder = ref(false)
const builderModuleId = ref<string | null>(null)
const builderInitialJSON = ref<string | null>(null)

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
  showBuilder.value = false
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
  showBuilder.value = false
}

function onViewImage(_item: main.Item, filename: string) {
  lightboxFilename.value = filename
  lightboxVisible.value = true
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

// Export backup state
const exporting = ref(false)
const exportMessage = ref<string | null>(null)

async function onExportBackup() {
  exporting.value = true
  exportMessage.value = null
  try {
    const path = await ExportBackup()
    if (path) {
      exportMessage.value = `Backup saved: ${path.split('/').pop()}`
      setTimeout(() => { exportMessage.value = null }, 5000)
    }
  } catch (e: any) {
    exportMessage.value = `Export failed: ${e?.message ?? e}`
  } finally {
    exporting.value = false
  }
}

// Schema builder handlers
function openNewSchemaBuilder() {
  builderModuleId.value = null
  builderInitialJSON.value = null
  showBuilder.value = true
  showForm.value = false
}

async function openEditSchemaBuilder(mod: main.ModuleSchema) {
  try {
    const json = await LoadModuleFile(mod.id)
    builderModuleId.value = mod.id
    builderInitialJSON.value = json
    showBuilder.value = true
    showForm.value = false
  } catch (e: any) {
    alert(`Failed to load schema: ${e?.message ?? e}`)
  }
}

async function onBuilderSaved() {
  showBuilder.value = false
  await moduleStore.fetchModules()
}

function onBuilderClose() {
  showBuilder.value = false
}
</script>

<template>
  <div class="app-layout">
    <aside class="sidebar">
      <h2>OmniCollect</h2>
      <button class="builder-btn" @click="openNewSchemaBuilder">+ New Schema</button>
      <button class="export-btn" :disabled="exporting" @click="onExportBackup">
        {{ exporting ? 'Exporting...' : 'Export Backup' }}
      </button>
      <div v-if="exportMessage" class="export-message">{{ exportMessage }}</div>
      <ModuleSelector
        :modules="moduleStore.modules"
        @select="onModuleSelect"
        @edit="openEditSchemaBuilder"
      />
    </aside>

    <main class="main-content">
      <div v-if="moduleStore.loading || collectionStore.loading" class="loading">
        Loading...
      </div>

      <div v-if="collectionStore.error" class="error-message">
        {{ collectionStore.error }}
      </div>

      <!-- Schema Builder -->
      <SchemaBuilder
        v-if="showBuilder"
        :moduleId="builderModuleId"
        :initialJSON="builderInitialJSON"
        @saved="onBuilderSaved"
        @close="onBuilderClose"
      />

      <!-- Dynamic Form (create/edit item) -->
      <DynamicForm
        v-if="showForm && selectedSchema && !showBuilder"
        :schema="selectedSchema"
        :item="editingItem"
        @save="onSave"
        @cancel="onCancel"
      />

      <!-- Collection views -->
      <template v-if="!showForm && !showBuilder">
        <div class="view-controls">
          <button
            class="view-toggle"
            :class="{active: viewMode === 'list'}"
            @click="viewMode = 'list'"
          >List</button>
          <button
            class="view-toggle"
            :class="{active: viewMode === 'grid'}"
            @click="viewMode = 'grid'"
          >Grid</button>
        </div>

        <ItemList
          v-if="viewMode === 'list'"
          :items="collectionStore.items"
          :modules="moduleStore.modules"
          :activeModuleId="collectionStore.activeModuleId"
          @select="onItemSelect"
          @filterChange="onFilterChange"
          @search="onSearch"
        />

        <CollectionGrid
          v-if="viewMode === 'grid'"
          :items="collectionStore.items"
          :modules="moduleStore.modules"
          @select="onItemSelect"
          @viewImage="onViewImage"
        />
      </template>

      <ImageLightbox
        :filename="lightboxFilename"
        :visible="lightboxVisible"
        @close="lightboxVisible = false"
      />
    </main>
  </div>
</template>

<style>
body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  color: var(--text-primary);
  background: var(--bg-primary);
}
.app-layout {
  display: flex;
  height: 100vh;
}
.sidebar {
  width: 240px;
  padding: 16px;
  border-right: 1px solid var(--border-primary);
  background: var(--bg-secondary);
  overflow-y: auto;
}
.sidebar h2 {
  margin: 0 0 12px 0;
  font-size: 18px;
}
.builder-btn {
  width: 100%;
  padding: 8px;
  margin-bottom: 12px;
  border: 1px dashed var(--accent-blue);
  border-radius: 4px;
  background: var(--bg-primary);
  color: var(--accent-blue);
  cursor: pointer;
  font-size: 13px;
}
.builder-btn:hover {
  background: var(--accent-blue-light);
}
.export-btn {
  width: 100%;
  padding: 8px;
  margin-bottom: 12px;
  border: 1px solid var(--border-primary);
  border-radius: 4px;
  background: var(--bg-primary);
  color: var(--text-primary);
  cursor: pointer;
  font-size: 13px;
}
.export-btn:hover {
  background: var(--bg-tertiary);
}
.export-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.export-message {
  font-size: 12px;
  color: var(--success-text);
  margin-bottom: 8px;
  padding: 4px 0;
}
.main-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  background: var(--bg-primary);
}
.loading {
  color: var(--text-muted);
  padding: 24px;
  text-align: center;
}
.error-message {
  background: var(--error-bg);
  color: var(--error-text);
  padding: 8px 12px;
  border-radius: 4px;
  margin-bottom: 12px;
  font-size: 14px;
}
.view-controls {
  display: flex;
  gap: 4px;
  margin-bottom: 12px;
}
.view-toggle {
  padding: 6px 12px;
  border: 1px solid var(--border-primary);
  background: var(--bg-primary);
  color: var(--text-primary);
  cursor: pointer;
  font-size: 13px;
  border-radius: 4px;
}
.view-toggle.active {
  background: var(--accent-blue);
  color: var(--text-on-accent);
  border-color: var(--accent-blue);
}
</style>
