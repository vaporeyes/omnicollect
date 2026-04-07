<script lang="ts" setup>
import {ref, nextTick, onMounted, onUnmounted} from 'vue'
import {main} from '../wailsjs/go/models'
import {LoadModuleFile, ExportBackup, LoadSettings} from '../wailsjs/go/main/App'
import {WindowSetSystemDefaultTheme} from '../wailsjs/runtime/runtime'
import {applyPolineTheme, DEFAULT_CONFIG, type ThemeConfig} from './theme'
import {useModuleStore} from './stores/moduleStore'
import {useCollectionStore} from './stores/collectionStore'
import ModuleSelector from './components/ModuleSelector.vue'
import DynamicForm from './components/DynamicForm.vue'
import ItemList from './components/ItemList.vue'
import CollectionGrid from './components/CollectionGrid.vue'
import ImageLightbox from './components/ImageLightbox.vue'
import SchemaBuilder from './components/SchemaBuilder.vue'
import ItemDetail from './components/ItemDetail.vue'
import SettingsPage from './components/SettingsPage.vue'
import ToastProvider from './components/ToastProvider.vue'
import CommandPalette from './components/CommandPalette.vue'
import ContextMenu from './components/ContextMenu.vue'
import type {MenuOption} from './components/ContextMenu.vue'
import {useToastStore} from './stores/toastStore'

const moduleStore = useModuleStore()
const collectionStore = useCollectionStore()
const toastStore = useToastStore()

// Theme configuration, loaded from settings on mount
const themeConfig = ref<ThemeConfig>(JSON.parse(JSON.stringify(DEFAULT_CONFIG)))
const showSettings = ref(false)
const darkMediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
const systemDark = ref(darkMediaQuery.matches)

function getEffectiveDark(): boolean {
  if (themeConfig.value.mode === 'system') return systemDark.value
  return themeConfig.value.mode === 'dark'
}

function refreshTheme() {
  applyPolineTheme(getEffectiveDark(), themeConfig.value)
}

// Apply immediately with defaults, then update when settings load
refreshTheme()

function onThemeChange(e: MediaQueryListEvent) {
  systemDark.value = e.matches
  refreshTheme()
}
darkMediaQuery.addEventListener('change', onThemeChange)
onUnmounted(() => {
  darkMediaQuery.removeEventListener('change', onThemeChange)
  document.removeEventListener('keydown', onGlobalKeydown)
})

// Tell Wails to follow system theme for window chrome
try { WindowSetSystemDefaultTheme() } catch { /* ok if not available in dev */ }

const selectedSchema = ref<main.ModuleSchema | null>(null)
const editingItem = ref<main.Item | null>(null)
const viewingItem = ref<main.Item | null>(null)
const viewingSchema = ref<main.ModuleSchema | null>(null)
const showForm = ref(false)
const showDetail = ref(false)
const viewMode = ref<'list' | 'grid'>('grid')

// Template ref for search focus
const itemListRef = ref<InstanceType<typeof ItemList> | null>(null)

// Lightbox state
const lightboxFilename = ref('')
const lightboxVisible = ref(false)

// Context menu state
const ctxVisible = ref(false)
const ctxX = ref(0)
const ctxY = ref(0)
const ctxItem = ref<main.Item | null>(null)
const ctxOptions: MenuOption[] = [
  {label: 'View Details', action: 'view'},
  {label: 'Edit', action: 'edit'},
  {label: 'Delete', action: 'delete', destructive: true},
]

// Command palette state
const showPalette = ref(false)

// Schema builder state
const showBuilder = ref(false)
const builderModuleId = ref<string | null>(null)
const builderInitialJSON = ref<string | null>(null)

// Context menu handlers
function onItemContextMenu(item: main.Item, x: number, y: number) {
  ctxItem.value = item
  ctxX.value = x
  ctxY.value = y
  ctxVisible.value = true
}

function onCtxSelect(action: string) {
  const item = ctxItem.value
  if (!item) return
  if (action === 'view') {
    onItemSelect(item)
  } else if (action === 'edit') {
    const schema = moduleStore.getModuleById(item.moduleId)
    if (!schema) {
      toastStore.show(`Schema not available for "${item.moduleId}"`, 'error')
      return
    }
    selectedSchema.value = schema
    editingItem.value = item
    viewingItem.value = item
    viewingSchema.value = schema
    showForm.value = true
    showDetail.value = false
    showBuilder.value = false
  } else if (action === 'delete') {
    onDeleteItem(item)
  }
}

// Shared delete logic for both detail view and context menu
async function onDeleteItem(item: main.Item) {
  const title = item.title
  try {
    await collectionStore.deleteItem(item.id)
    // If we were viewing this item, close the detail view
    if (viewingItem.value?.id === item.id) {
      showDetail.value = false
      viewingItem.value = null
      viewingSchema.value = null
    }
    toastStore.show(`"${title}" deleted`, 'success')
  } catch {
    toastStore.show(collectionStore.error ?? 'Failed to delete item', 'error')
  }
}

// Global keyboard shortcuts
function onGlobalKeydown(e: KeyboardEvent) {
  const mod = e.metaKey || e.ctrlKey

  // Cmd/Ctrl+K: toggle command palette
  if (mod && e.key === 'k') {
    e.preventDefault()
    showPalette.value = !showPalette.value
    return
  }

  // Escape: close topmost overlay
  if (e.key === 'Escape') {
    if (showPalette.value) { showPalette.value = false; return }
    if (lightboxVisible.value) { lightboxVisible.value = false; return }
    if (ctxVisible.value) { ctxVisible.value = false; return }
    if (showForm.value) { onCancel(); return }
    if (showDetail.value) { onCloseDetail(); return }
    if (showBuilder.value) { onBuilderClose(); return }
    if (showSettings.value) { onSettingsClose(); return }
    return
  }

  // Cmd/Ctrl+F: focus search (switch to list view if needed)
  if (mod && e.key === 'f') {
    e.preventDefault()
    // Close overlays first
    showForm.value = false
    showDetail.value = false
    showBuilder.value = false
    showSettings.value = false
    if (viewMode.value !== 'list') {
      viewMode.value = 'list'
    }
    nextTick(() => itemListRef.value?.focusSearch())
    return
  }

  // Cmd/Ctrl+N: new item for active module (or first module)
  if (mod && e.key === 'n') {
    e.preventDefault()
    const activeId = collectionStore.activeModuleId
    const mod2 = activeId
      ? moduleStore.getModuleById(activeId)
      : moduleStore.modules[0] ?? null
    if (mod2) {
      onModuleSelect(mod2)
    } else {
      toastStore.show('Create a collection schema first', 'info')
    }
    return
  }
}

onMounted(async () => {
  document.addEventListener('keydown', onGlobalKeydown)

  // Load saved theme settings
  try {
    const json = await LoadSettings()
    const parsed = JSON.parse(json)
    if (parsed.theme) {
      themeConfig.value = {...DEFAULT_CONFIG, ...parsed.theme}
      refreshTheme()
    }
  } catch { /* use defaults */ }

  await Promise.all([
    moduleStore.fetchModules(),
    collectionStore.fetchItems(),
  ])
})

function onModuleSelect(mod: main.ModuleSchema) {
  selectedSchema.value = mod
  editingItem.value = null
  viewingItem.value = null
  showForm.value = true
  showDetail.value = false
  showBuilder.value = false
}

function onItemSelect(item: main.Item) {
  const schema = moduleStore.getModuleById(item.moduleId)
  viewingItem.value = item
  viewingSchema.value = schema ?? null
  showDetail.value = true
  showForm.value = false
  showBuilder.value = false
}

function onEditFromDetail() {
  if (!viewingItem.value) return
  const schema = viewingSchema.value
  if (!schema) {
    toastStore.show(`Schema not available for "${viewingItem.value.moduleId}". The schema file may have been removed.`, 'error')
    return
  }
  selectedSchema.value = schema
  editingItem.value = viewingItem.value
  showForm.value = true
  showDetail.value = false
}

function onCloseDetail() {
  showDetail.value = false
  viewingItem.value = null
  viewingSchema.value = null
}

function onViewImage(_item: main.Item, filename: string) {
  lightboxFilename.value = filename
  lightboxVisible.value = true
}

function onDetailViewImage(filename: string) {
  lightboxFilename.value = filename
  lightboxVisible.value = true
}

async function onSave(item: main.Item) {
  try {
    const saved = await collectionStore.saveItem(item)
    showForm.value = false
    editingItem.value = null
    // Return to detail view of the saved item
    if (saved) {
      viewingItem.value = saved
      viewingSchema.value = moduleStore.getModuleById(saved.moduleId) ?? null
      showDetail.value = true
      toastStore.show('Item saved', 'success')
    }
  } catch {
    toastStore.show(collectionStore.error ?? 'Failed to save item', 'error')
  }
}

function onDeleteFromDetail() {
  if (!viewingItem.value) return
  onDeleteItem(viewingItem.value)
}

function onCancel() {
  showForm.value = false
  editingItem.value = null
  // Return to detail view if we were viewing an item
  if (viewingItem.value) {
    showDetail.value = true
  }
}

function onAddFirstItem() {
  if (moduleStore.modules.length > 0) {
    onModuleSelect(moduleStore.modules[0])
  }
}

function onFilterChange(moduleId: string) {
  collectionStore.setFilter(moduleId)
}

function onSearch(query: string) {
  collectionStore.setSearch(query)
}

// Command palette handlers
function onPaletteSelectItem(item: main.Item) {
  showPalette.value = false
  onItemSelect(item)
}

function onPaletteAction(action: string) {
  showPalette.value = false
  if (action === 'newItem') {
    const activeId = collectionStore.activeModuleId
    const mod = activeId
      ? moduleStore.getModuleById(activeId)
      : moduleStore.modules[0] ?? null
    if (mod) onModuleSelect(mod)
    else toastStore.show('Create a collection schema first', 'info')
  } else if (action === 'newSchema') {
    openNewSchemaBuilder()
  } else if (action === 'openSettings') {
    openSettings()
  } else if (action === 'exportBackup') {
    onExportBackup()
  }
}

// Export backup state
const exporting = ref(false)

async function onExportBackup() {
  exporting.value = true
  try {
    const path = await ExportBackup()
    if (path) {
      toastStore.show(`Backup saved: ${path.split('/').pop()}`, 'success')
    }
  } catch (e: any) {
    toastStore.show(`Export failed: ${e?.message ?? e}`, 'error')
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
    toastStore.show(`Failed to load schema: ${e?.message ?? e}`, 'error')
  }
}

async function onBuilderSaved() {
  showBuilder.value = false
  await moduleStore.fetchModules()
}

function onBuilderClose() {
  showBuilder.value = false
}

function openSettings() {
  showSettings.value = true
  showForm.value = false
  showDetail.value = false
  showBuilder.value = false
}

function onSettingsSaved(config: ThemeConfig) {
  themeConfig.value = config
  refreshTheme()
}

function onSettingsClose() {
  // Re-apply saved config in case user changed things without saving
  refreshTheme()
  showSettings.value = false
}
</script>

<template>
  <div class="app-layout">
    <aside class="sidebar animate-slide-up delay-1">
      <h2 class="animate-fade-in delay-2">OmniCollect</h2>
      <button class="builder-btn animate-fade-in delay-3" @click="openNewSchemaBuilder">+ New Schema</button>
      <div class="sidebar-scroll animate-fade-in delay-4">
        <ModuleSelector
          :modules="moduleStore.modules"
          @select="onModuleSelect"
          @edit="openEditSchemaBuilder"
          @createSchema="openNewSchemaBuilder"
        />
      </div>
      <div class="sidebar-bottom animate-fade-in delay-5">
        <button class="export-btn" :disabled="exporting" @click="onExportBackup">
          {{ exporting ? 'Exporting...' : 'Export Backup' }}
        </button>
        <button class="settings-btn" @click="openSettings">&#9881; Settings</button>
      </div>
    </aside>

    <main class="main-content animate-slide-up delay-2">
      <div v-if="moduleStore.loading || collectionStore.loading" class="loading">
        Loading...
      </div>

      <div v-if="collectionStore.error" class="error-message">
        {{ collectionStore.error }}
      </div>

      <!-- Settings Page -->
      <Transition name="fade-slide" mode="out-in">
        <SettingsPage
          v-if="showSettings"
          :initialConfig="themeConfig"
          :systemDark="systemDark"
          @saved="onSettingsSaved"
          @close="onSettingsClose"
        />
      </Transition>

      <!-- Schema Builder -->
      <Transition name="fade-slide" mode="out-in">
        <SchemaBuilder
          v-if="showBuilder && !showSettings"
          :moduleId="builderModuleId"
          :initialJSON="builderInitialJSON"
          @saved="onBuilderSaved"
          @close="onBuilderClose"
        />
      </Transition>

      <!-- Dynamic Form (create/edit item) -->
      <Transition name="fade-slide" mode="out-in">
        <DynamicForm
          v-if="showForm && selectedSchema && !showBuilder && !showSettings"
          :schema="selectedSchema"
          :item="editingItem"
          @save="onSave"
          @cancel="onCancel"
        />
      </Transition>

      <!-- Item Detail View -->
      <Transition name="fade-slide" mode="out-in">
        <ItemDetail
          v-if="showDetail && viewingItem && !showForm && !showBuilder && !showSettings"
          :item="viewingItem"
          :schema="viewingSchema"
          @edit="onEditFromDetail"
          @delete="onDeleteFromDetail"
          @close="onCloseDetail"
          @viewImage="onDetailViewImage"
        />
      </Transition>

      <!-- Collection views -->
      <template v-if="!showForm && !showDetail && !showBuilder && !showSettings">
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

        <Transition name="fade-slide" mode="out-in">
          <ItemList
            v-if="viewMode === 'list'"
            key="list"
            ref="itemListRef"
            :items="collectionStore.items"
            :modules="moduleStore.modules"
            :activeModuleId="collectionStore.activeModuleId"
            @select="onItemSelect"
            @filterChange="onFilterChange"
            @search="onSearch"
            @addItem="onAddFirstItem"
            @itemContextMenu="onItemContextMenu"
          />

          <CollectionGrid
            v-else-if="viewMode === 'grid'"
            key="grid"
            :items="collectionStore.items"
            :modules="moduleStore.modules"
            @select="onItemSelect"
            @viewImage="onViewImage"
            @addItem="onAddFirstItem"
            @itemContextMenu="onItemContextMenu"
          />
        </Transition>
      </template>

      <ImageLightbox
        :filename="lightboxFilename"
        :visible="lightboxVisible"
        @close="lightboxVisible = false"
      />
    </main>

    <CommandPalette
      :visible="showPalette"
      @close="showPalette = false"
      @selectItem="onPaletteSelectItem"
      @action="onPaletteAction"
    />
    <ContextMenu
      :visible="ctxVisible"
      :x="ctxX"
      :y="ctxY"
      :options="ctxOptions"
      @select="onCtxSelect"
      @close="ctxVisible = false"
    />
    <ToastProvider />
  </div>
</template>

<style>
body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Inter', Roboto, sans-serif;
  color: var(--text-primary);
  background: var(--bg-primary);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  line-height: 1.5;
  font-size: 14px;
}
.app-layout {
  display: flex;
  height: 100vh;
}
.sidebar {
  width: 250px;
  padding: 20px 16px 12px;
  background: var(--bg-tertiary);
  -webkit-backdrop-filter: blur(var(--glass-blur));
  backdrop-filter: blur(var(--glass-blur));
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex-shrink: 0;
}
.sidebar h2 {
  margin: 0 0 12px 0;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide);
  color: var(--text-muted);
}
.sidebar-scroll {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}
.sidebar-bottom {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding-top: 12px;
  border-top: 1px solid var(--border-primary);
  margin-top: 8px;
}
.builder-btn {
  width: 100%;
  padding: 9px 12px;
  margin-bottom: 4px;
  border: 1px dashed var(--accent-blue);
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--accent-blue);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: background var(--transition-fast), border-color var(--transition-fast);
}
.builder-btn:hover {
  background: var(--accent-blue-light);
  border-color: var(--accent-blue-hover);
}
.export-btn {
  width: 100%;
  padding: 7px 12px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 12px;
  transition: background var(--transition-fast), color var(--transition-fast);
}
.export-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}
.export-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.settings-btn {
  width: 100%;
  padding: 7px 12px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 12px;
  transition: background var(--transition-fast), color var(--transition-fast);
}
.settings-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}
.main-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
  background: var(--bg-primary);
  border-top-left-radius: var(--radius-lg);
  box-shadow: -2px 0 16px hsla(220, 10%, 20%, 0.06);
  position: relative;
  z-index: 1;
}
.loading {
  color: var(--text-muted);
  padding: 48px;
  text-align: center;
  font-size: 14px;
}
.error-message {
  background: var(--error-bg);
  color: var(--error-text);
  padding: 10px 14px;
  border-radius: var(--radius-md);
  border-left: 3px solid var(--error-border);
  margin-bottom: 16px;
  font-size: 13px;
}
.view-controls {
  display: flex;
  gap: 2px;
  margin-bottom: 16px;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  padding: 3px;
  width: fit-content;
}
.view-toggle {
  padding: 6px 14px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast), color var(--transition-fast);
}
.view-toggle:hover {
  color: var(--text-primary);
}
.view-toggle.active {
  background: var(--accent-blue);
  color: var(--text-on-accent);
  box-shadow: var(--shadow-sm);
}

/* fade-slide: content panels slide up slightly and fade in */
.fade-slide-enter-active {
  transition: opacity 0.2s ease-out, transform 0.2s ease-out;
}
.fade-slide-leave-active {
  transition: opacity 0.12s ease-in, transform 0.12s ease-in;
}
.fade-slide-enter-from {
  opacity: 0;
  transform: translateY(10px);
}
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
