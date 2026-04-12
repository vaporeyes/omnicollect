<script lang="ts" setup>
import {ref, computed, nextTick, onMounted, onUnmounted} from 'vue'
import * as api from './api/client'
import type {Item, ModuleSchema, BulkDeleteResult, BulkUpdateResult, TagCount} from './api/types'
import {applyPolineTheme, DEFAULT_CONFIG, type ThemeConfig} from './theme'
import {useModuleStore} from './stores/moduleStore'
import {useCollectionStore} from './stores/collectionStore'
import {useToastStore} from './stores/toastStore'
import {useSelectionStore} from './stores/selectionStore'
import {useSmartFolderStore, type SmartFolder} from './stores/smartFolderStore'
import AppSidebar from './components/AppSidebar.vue'
import DynamicForm from './components/DynamicForm.vue'
import ItemList from './components/ItemList.vue'
import CollectionGrid from './components/CollectionGrid.vue'
import ImageLightbox from './components/ImageLightbox.vue'
import SchemaBuilder from './components/SchemaBuilder.vue'
import ItemDetail from './components/ItemDetail.vue'
import SettingsPage from './components/SettingsPage.vue'
import ToastProvider from './components/ToastProvider.vue'
import FilterBar from './components/FilterBar.vue'
import CommandPalette from './components/CommandPalette.vue'
import ContextMenu from './components/ContextMenu.vue'
import type {MenuOption} from './components/ContextMenu.vue'
import BulkActionBar from './components/BulkActionBar.vue'
import TagFilter from './components/TagFilter.vue'
import TagManager from './components/TagManager.vue'
import ImportDialog from './components/ImportDialog.vue'
import DashboardView from './components/DashboardView.vue'
import ComparisonView from './components/ComparisonView.vue'
import {isAuthConfigured} from './auth/plugin'
import {AuthGuard} from './auth/guard'
import {useAuth0} from '@auth0/auth0-vue'

const moduleStore = useModuleStore()
const collectionStore = useCollectionStore()
const toastStore = useToastStore()
const selectionStore = useSelectionStore()
const smartFolderStore = useSmartFolderStore()

// Auth
const authEnabled = isAuthConfigured
const auth0 = authEnabled ? useAuth0() : null
function onSignOut() {
  auth0?.logout({logoutParams: {returnTo: window.location.origin}})
}

// Tag state
const allTags = ref<TagCount[]>([])
const showTagManager = ref(false)

async function refreshTags() {
  try { allTags.value = await api.getAllTags() } catch { /* ignore */ }
}

async function onTagRename(payload: {oldName: string, newName: string}) {
  try {
    await api.renameTag(payload.oldName, payload.newName)
    await refreshTags()
    await collectionStore.fetchItems()
    toastStore.show(`Tag renamed to "${payload.newName}"`, 'success')
  } catch (e: any) {
    toastStore.show(`Rename failed: ${e?.message ?? e}`, 'error')
  }
}

async function onTagDelete(name: string) {
  try {
    await api.deleteTag(name)
    await refreshTags()
    await collectionStore.fetchItems()
    toastStore.show(`Tag "${name}" deleted`, 'success')
  } catch (e: any) {
    toastStore.show(`Delete failed: ${e?.message ?? e}`, 'error')
  }
}

// Bulk action state
const showBulkDeleteConfirm = ref(false)
const showBulkModuleDialog = ref(false)
const bulkTargetModuleId = ref('')

// Theme
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

// View state
const selectedSchema = ref<ModuleSchema | null>(null)
const editingItem = ref<Item | null>(null)
const viewingItem = ref<Item | null>(null)
const viewingSchema = ref<ModuleSchema | null>(null)
const showForm = ref(false)
const showDetail = ref(false)
const viewMode = ref<'list' | 'grid'>('grid')
const showDashboard = ref(true)
const showComparison = ref(false)
const comparisonItems = ref<[Item, Item] | null>(null)

const activeFilterSchema = computed(() => {
  const id = collectionStore.activeModuleId
  if (!id) return null
  return moduleStore.getModuleById(id) ?? null
})

const itemListRef = ref<InstanceType<typeof ItemList> | null>(null)

// Lightbox
const lightboxFilename = ref('')
const lightboxVisible = ref(false)

// Item context menu
const ctxVisible = ref(false)
const ctxX = ref(0)
const ctxY = ref(0)
const ctxItem = ref<Item | null>(null)
const ctxOptions: MenuOption[] = [
  {label: 'View Details', action: 'view'},
  {label: 'Edit', action: 'edit'},
  {label: 'Delete', action: 'delete', destructive: true},
]

// Command palette
const showPalette = ref(false)

// Schema builder
const showBuilder = ref(false)
const builderModuleId = ref<string | null>(null)
const builderInitialJSON = ref<string | null>(null)

// Smart Folder context menu
const sidebarRef = ref<InstanceType<typeof AppSidebar> | null>(null)
const sfCtxFolder = ref<SmartFolder | null>(null)
const sfCtxVisible = ref(false)
const sfCtxX = ref(0)
const sfCtxY = ref(0)
const sfCtxOptions: MenuOption[] = [
  {label: 'Rename', action: 'rename'},
  {label: 'Delete', action: 'delete', destructive: true},
]

// Import/export
const showImportDialog = ref(false)
const exporting = ref(false)

// --- Handlers ---

function onItemContextMenu(item: Item, x: number, y: number) {
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

async function onDeleteItem(item: Item) {
  const title = item.title
  try {
    await collectionStore.deleteItem(item.id)
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

// Keyboard shortcuts
function onGlobalKeydown(e: KeyboardEvent) {
  const mod = e.metaKey || e.ctrlKey

  if (mod && e.key === 'k') {
    e.preventDefault()
    showPalette.value = !showPalette.value
    return
  }

  if (e.key === 'Escape') {
    if (showImportDialog.value) { showImportDialog.value = false; return }
    if (showPalette.value) { showPalette.value = false; return }
    if (showComparison.value) { onCloseComparison(); return }
    if (lightboxVisible.value) { lightboxVisible.value = false; return }
    if (ctxVisible.value) { ctxVisible.value = false; return }
    if (showForm.value) { onCancel(); return }
    if (showDetail.value) { onCloseDetail(); return }
    if (showTagManager.value) { showTagManager.value = false; return }
    if (showBuilder.value) { showBuilder.value = false; return }
    if (showSettings.value) { onSettingsClose(); return }
    return
  }

  if (mod && e.key === 'f') {
    e.preventDefault()
    showForm.value = false
    showDetail.value = false
    showBuilder.value = false
    showSettings.value = false
    showDashboard.value = false
    if (viewMode.value !== 'list') viewMode.value = 'list'
    nextTick(() => itemListRef.value?.focusSearch())
    return
  }

  if (mod && e.key === 'n') {
    e.preventDefault()
    const activeId = collectionStore.activeModuleId
    const m = activeId
      ? moduleStore.getModuleById(activeId)
      : moduleStore.modules[0] ?? null
    if (m) onNewItem(m)
    else toastStore.show('Create a collection schema first', 'info')
    return
  }
}

onMounted(async () => {
  document.addEventListener('keydown', onGlobalKeydown)

  try {
    const settings = await api.get<any>('/api/v1/settings')
    if (settings?.theme) {
      themeConfig.value = {...DEFAULT_CONFIG, ...settings.theme}
      refreshTheme()
    }
    smartFolderStore.loadFromSettings(settings)
  } catch { /* use defaults */ }

  await Promise.all([
    moduleStore.fetchModules(),
    collectionStore.fetchItems(),
    refreshTags(),
  ])
})

// Sidebar navigation: clicking a module filters the view (does NOT open form)
function onNavigate(moduleId: string) {
  selectionStore.clear()
  smartFolderStore.clearActive()
  collectionStore.setFilter(moduleId)
  // Close any open panels when navigating
  showForm.value = false
  showDetail.value = false
  showBuilder.value = false
  showSettings.value = false
  showTagManager.value = false
  showComparison.value = false
  comparisonItems.value = null
  if (!moduleId) showDashboard.value = true
}

// Explicit new item action (from sidebar + button or Cmd+N)
function onNewItem(mod: ModuleSchema) {
  selectionStore.clear()
  selectedSchema.value = mod
  editingItem.value = null
  viewingItem.value = null
  showForm.value = true
  showDetail.value = false
  showBuilder.value = false
}

function onItemSelect(item: Item) {
  selectionStore.clear()
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

function onViewImage(_item: Item, filename: string) {
  lightboxFilename.value = filename
  lightboxVisible.value = true
}

function onDetailViewImage(filename: string) {
  lightboxFilename.value = filename
  lightboxVisible.value = true
}

async function onSave(item: Item) {
  try {
    const saved = await collectionStore.saveItem(item)
    showForm.value = false
    editingItem.value = null
    if (saved) {
      viewingItem.value = saved
      viewingSchema.value = moduleStore.getModuleById(saved.moduleId) ?? null
      showDetail.value = true
      toastStore.show('Item saved', 'success')
      refreshTags()
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
  if (viewingItem.value) showDetail.value = true
}

function onDashboardSelectItem(id: string) {
  const item = collectionStore.items.find(i => i.id === id)
  if (item) onItemSelect(item)
}

function onAddFirstItem() {
  if (moduleStore.modules.length > 0) onNewItem(moduleStore.modules[0])
}

// Smart Folder handlers
function onSmartFolderApply(folder: SmartFolder) {
  if (folder.moduleId && !moduleStore.getModuleById(folder.moduleId)) {
    toastStore.show('Module no longer exists. Showing all items.', 'info')
    collectionStore.activeModuleId = ''
  } else {
    collectionStore.activeModuleId = folder.moduleId
  }
  collectionStore.searchQuery = folder.searchQuery || ''
  collectionStore.activeFilters = folder.filters ? JSON.parse(JSON.stringify(folder.filters)) : {}
  collectionStore.activeTags = folder.tags ? [...folder.tags] : []
  showForm.value = false
  showDetail.value = false
  showBuilder.value = false
  showSettings.value = false
  showTagManager.value = false
  if (!collectionStore.activeModuleId) showDashboard.value = true
  collectionStore.fetchItems()
}

function onSmartFolderContextMenu(folder: SmartFolder, x: number, y: number) {
  sfCtxFolder.value = folder
  sfCtxX.value = x
  sfCtxY.value = y
  sfCtxVisible.value = true
}

function onSfCtxSelect(action: string) {
  const folder = sfCtxFolder.value
  if (!folder) return
  if (action === 'rename') {
    sidebarRef.value?.startRename(folder.id)
  } else if (action === 'delete') {
    smartFolderStore.remove(folder.id)
    toastStore.show(`"${folder.name}" deleted`, 'success')
  }
}

// Search (from ItemList)
function onSearch(query: string) {
  smartFolderStore.clearActive()
  collectionStore.setSearch(query)
}

// Command palette
function onPaletteSelectItem(item: Item) {
  showPalette.value = false
  onItemSelect(item)
}

function onPaletteAction(action: string) {
  showPalette.value = false
  if (action === 'newItem') {
    const activeId = collectionStore.activeModuleId
    const m = activeId
      ? moduleStore.getModuleById(activeId)
      : moduleStore.modules[0] ?? null
    if (m) onNewItem(m)
    else toastStore.show('Create a collection schema first', 'info')
  } else if (action === 'newSchema') {
    openNewSchemaBuilder()
  } else if (action === 'manageTags') {
    openTagManager()
  } else if (action === 'openSettings') {
    openSettings()
  } else if (action === 'exportBackup') {
    onExportBackup()
  } else if (action === 'importBackup') {
    showImportDialog.value = true
  }
}

// Bulk actions
async function onBulkDelete() {
  const ids = selectionStore.selectedIdArray()
  try {
    const result = await api.post<BulkDeleteResult>('/api/v1/items/batch-delete', {ids})
    selectionStore.clear()
    await collectionStore.fetchItems()
    toastStore.show(`${result.deleted} item${result.deleted === 1 ? '' : 's'} deleted`, 'success')
  } catch (e: any) {
    toastStore.show(`Bulk delete failed: ${e?.message ?? e}`, 'error')
  }
  showBulkDeleteConfirm.value = false
}

async function onBulkExportCSV() {
  const ids = selectionStore.selectedIdArray()
  try {
    await api.downloadFile('/api/v1/export/csv', {ids})
    toastStore.show(`Exported ${ids.length} items`, 'success')
  } catch (e: any) {
    toastStore.show(`CSV export failed: ${e?.message ?? e}`, 'error')
  }
}

async function onBulkUpdateModule() {
  if (!bulkTargetModuleId.value) return
  const ids = selectionStore.selectedIdArray()
  try {
    const result = await api.post<BulkUpdateResult>('/api/v1/items/batch-update-module', {ids, newModuleId: bulkTargetModuleId.value})
    selectionStore.clear()
    await collectionStore.fetchItems()
    toastStore.show(`${result.updated} item${result.updated === 1 ? '' : 's'} moved`, 'success')
  } catch (e: any) {
    toastStore.show(`Module update failed: ${e?.message ?? e}`, 'error')
  }
  showBulkModuleDialog.value = false
  bulkTargetModuleId.value = ''
}

function onCompare() {
  const ids = selectionStore.selectedIdArray()
  if (ids.length !== 2) return
  const a = collectionStore.items.find(i => i.id === ids[0])
  const b = collectionStore.items.find(i => i.id === ids[1])
  if (!a || !b) return
  comparisonItems.value = [a, b]
  showComparison.value = true
}

function onCloseComparison() {
  showComparison.value = false
  comparisonItems.value = null
}

async function onImported() {
  showImportDialog.value = false
  await Promise.all([
    moduleStore.fetchModules(),
    collectionStore.fetchItems(),
    refreshTags(),
  ])
  toastStore.show('Backup imported successfully', 'success')
}

async function onExportBackup() {
  exporting.value = true
  try {
    await api.downloadFile('/api/v1/export/backup')
    toastStore.show('Backup downloaded', 'success')
  } catch (e: any) {
    toastStore.show(`Export failed: ${e?.message ?? e}`, 'error')
  } finally {
    exporting.value = false
  }
}

// Schema builder
function openNewSchemaBuilder() {
  builderModuleId.value = null
  builderInitialJSON.value = null
  showBuilder.value = true
  showForm.value = false
}

async function openEditSchemaBuilder(mod: ModuleSchema) {
  try {
    const data = await api.get<any>('/api/v1/modules/' + mod.id + '/file')
    builderModuleId.value = mod.id
    builderInitialJSON.value = typeof data === 'string' ? data : JSON.stringify(data, null, 2)
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

function openTagManager() {
  selectionStore.clear()
  showTagManager.value = true
  showForm.value = false
  showDetail.value = false
  showBuilder.value = false
  showSettings.value = false
  refreshTags()
}

function openSettings() {
  selectionStore.clear()
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
  refreshTheme()
  showSettings.value = false
}
</script>

<template>
  <component :is="authEnabled ? AuthGuard : 'div'">
  <div class="app-layout">
    <AppSidebar
      ref="sidebarRef"
      :exporting="exporting"
      :authEnabled="authEnabled"
      @navigate="onNavigate"
      @newItem="onNewItem"
      @newSchema="openNewSchemaBuilder"
      @editSchema="openEditSchemaBuilder"
      @applySmartFolder="onSmartFolderApply"
      @smartFolderContextMenu="onSmartFolderContextMenu"
      @exportBackup="onExportBackup"
      @importBackup="showImportDialog = true"
      @openTags="openTagManager"
      @openSettings="openSettings"
      @signOut="onSignOut"
    />

    <main class="main-content">
      <div v-if="moduleStore.loading || collectionStore.loading" class="loading">
        Loading...
      </div>

      <div v-if="collectionStore.error" class="error-message">
        {{ collectionStore.error }}
      </div>

      <Transition name="fade-slide" mode="out-in">
        <SettingsPage
          v-if="showSettings"
          :initialConfig="themeConfig"
          :systemDark="systemDark"
          @saved="onSettingsSaved"
          @close="onSettingsClose"
        />
      </Transition>

      <Transition name="fade-slide" mode="out-in">
        <TagManager
          v-if="showTagManager && !showSettings"
          :tags="allTags"
          @rename="onTagRename"
          @delete="onTagDelete"
          @close="showTagManager = false"
        />
      </Transition>

      <Transition name="fade-slide" mode="out-in">
        <SchemaBuilder
          v-if="showBuilder && !showSettings && !showTagManager"
          :moduleId="builderModuleId"
          :initialJSON="builderInitialJSON"
          @saved="onBuilderSaved"
          @close="showBuilder = false"
        />
      </Transition>

      <Transition name="fade-slide" mode="out-in">
        <DynamicForm
          v-if="showForm && selectedSchema && !showBuilder && !showSettings && !showTagManager"
          :schema="selectedSchema"
          :item="editingItem"
          @save="onSave"
          @cancel="onCancel"
        />
      </Transition>

      <Transition name="fade-slide" mode="out-in">
        <ItemDetail
          v-if="showDetail && viewingItem && !showForm && !showBuilder && !showSettings && !showTagManager"
          :item="viewingItem"
          :schema="viewingSchema"
          @edit="onEditFromDetail"
          @delete="onDeleteFromDetail"
          @close="onCloseDetail"
          @viewImage="onDetailViewImage"
        />
      </Transition>

      <Transition name="fade-slide" mode="out-in">
        <ComparisonView
          v-if="showComparison && comparisonItems && !showSettings && !showTagManager"
          :itemA="comparisonItems[0]"
          :itemB="comparisonItems[1]"
          :schemaA="moduleStore.getModuleById(comparisonItems[0].moduleId) ?? null"
          :schemaB="moduleStore.getModuleById(comparisonItems[1].moduleId) ?? null"
          @close="onCloseComparison"
          @viewImage="onDetailViewImage"
        />
      </Transition>

      <!-- Collection views -->
      <template v-if="!showForm && !showDetail && !showBuilder && !showSettings && !showTagManager && !showComparison">
        <div class="view-controls">
          <button
            v-if="!collectionStore.activeModuleId"
            class="view-toggle"
            :class="{active: showDashboard}"
            @click="showDashboard = true"
          >Insights</button>
          <button
            class="view-toggle"
            :class="{active: !showDashboard && viewMode === 'list'}"
            @click="showDashboard = false; viewMode = 'list'"
          >List</button>
          <button
            class="view-toggle"
            :class="{active: !showDashboard && viewMode === 'grid'}"
            @click="showDashboard = false; viewMode = 'grid'"
          >Grid</button>
        </div>

        <FilterBar
          :schema="activeFilterSchema"
          :filters="collectionStore.activeFilters"
          @update="(f: any) => { smartFolderStore.clearActive(); collectionStore.setActiveFilters(f) }"
          @clear="() => { smartFolderStore.clearActive(); collectionStore.clearFilters() }"
        />

        <TagFilter
          :allTags="allTags"
          :selectedTags="collectionStore.activeTags"
          @update="(t: string[]) => { smartFolderStore.clearActive(); collectionStore.setTags(t) }"
        />

        <div v-if="collectionStore.items.length === 0 && Object.keys(collectionStore.activeFilters).length > 0" class="filtered-empty">
          No items match the active filters.
          <button class="filtered-empty-clear" @click="collectionStore.clearFilters">Clear filters</button>
        </div>

        <Transition name="fade-slide" mode="out-in">
          <DashboardView
            v-if="showDashboard && !collectionStore.activeModuleId"
            key="dashboard"
            :items="collectionStore.items"
            :modules="moduleStore.modules"
            :dark="getEffectiveDark()"
            @selectItem="onDashboardSelectItem"
          />

          <ItemList
            v-else-if="viewMode === 'list'"
            key="list"
            ref="itemListRef"
            :items="collectionStore.items"
            :modules="moduleStore.modules"
            :activeModuleId="collectionStore.activeModuleId"
            @select="onItemSelect"
            @filterChange="onNavigate"
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

    <BulkActionBar
      :count="selectionStore.count"
      @delete="showBulkDeleteConfirm = true"
      @export="onBulkExportCSV"
      @editModule="showBulkModuleDialog = true; bulkTargetModuleId = ''"
      @deselectAll="selectionStore.clear()"
      @compare="onCompare"
    />

    <!-- Bulk delete confirmation -->
    <Teleport to="body">
      <div v-if="showBulkDeleteConfirm" class="confirm-overlay" @click.self="showBulkDeleteConfirm = false">
        <div class="confirm-dialog">
          <p class="confirm-title">Delete {{ selectionStore.count }} item{{ selectionStore.count === 1 ? '' : 's' }}?</p>
          <p class="confirm-message">This action cannot be undone.</p>
          <div class="confirm-actions">
            <button class="confirm-cancel-btn" @click="showBulkDeleteConfirm = false">Cancel</button>
            <button class="confirm-delete-btn" @click="onBulkDelete">Delete</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Bulk module edit dialog -->
    <Teleport to="body">
      <div v-if="showBulkModuleDialog" class="confirm-overlay" @click.self="showBulkModuleDialog = false">
        <div class="confirm-dialog">
          <p class="confirm-title">Move {{ selectionStore.count }} item{{ selectionStore.count === 1 ? '' : 's' }} to:</p>
          <select v-model="bulkTargetModuleId" class="bulk-module-select">
            <option value="" disabled>Select module...</option>
            <option v-for="mod in moduleStore.modules" :key="mod.id" :value="mod.id">{{ mod.displayName }}</option>
          </select>
          <div class="confirm-actions">
            <button class="confirm-cancel-btn" @click="showBulkModuleDialog = false">Cancel</button>
            <button class="confirm-delete-btn" :disabled="!bulkTargetModuleId" @click="onBulkUpdateModule" style="background: var(--accent-blue)">Move</button>
          </div>
        </div>
      </div>
    </Teleport>

    <ImportDialog
      v-if="showImportDialog"
      @close="showImportDialog = false"
      @imported="onImported"
    />
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
    <ContextMenu
      :visible="sfCtxVisible"
      :x="sfCtxX"
      :y="sfCtxY"
      :options="sfCtxOptions"
      @select="onSfCtxSelect"
      @close="sfCtxVisible = false"
    />
    <ToastProvider />
  </div>
  </component>
</template>

<style>
body {
  margin: 0;
  font-family: var(--font-body);
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
.confirm-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 4000;
}
.confirm-dialog {
  background: var(--bg-primary, #1e1e2e);
  border: 1px solid var(--border-primary, #333);
  border-radius: var(--radius-md);
  padding: 28px;
  max-width: 360px;
  width: 90%;
  box-shadow: var(--shadow-lg);
}
.confirm-title {
  margin: 0 0 4px;
  font-family: var(--font-heading);
  font-size: 18px;
  font-weight: 400;
  color: var(--text-primary);
}
.confirm-message {
  margin: 0 0 20px;
  font-size: 13px;
  color: var(--text-secondary);
}
.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.confirm-cancel-btn {
  padding: 8px 18px;
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-primary);
  cursor: pointer;
  font-size: 13px;
}
.confirm-cancel-btn:hover {
  background: var(--bg-hover);
}
.confirm-delete-btn {
  padding: 8px 18px;
  border: none;
  border-radius: var(--radius-sm);
  background: var(--error-border, #dc2626);
  color: #fff;
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
}
.confirm-delete-btn:hover {
  background: #b91c1c;
}
.confirm-delete-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.bulk-module-select {
  width: 100%;
  padding: 8px;
  margin: 12px 0 20px;
  border: 1px solid var(--border-input);
  border-radius: var(--radius-sm);
  font-size: 14px;
  background: var(--bg-input);
  color: var(--text-primary);
}
.filtered-empty {
  text-align: center;
  padding: 32px 16px;
  color: var(--text-muted);
  font-size: 14px;
}
.filtered-empty-clear {
  display: inline;
  background: none;
  border: none;
  color: var(--accent-blue);
  cursor: pointer;
  font-size: 14px;
  text-decoration: underline;
  padding: 0;
  margin-left: 4px;
}
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
