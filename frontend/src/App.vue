<script lang="ts" setup>
import {ref, computed, nextTick, onMounted, onUnmounted, watch} from 'vue'
import * as api from './api/client'
import {lazyComponent} from './composables/lazyComponent'
import type {Item, ModuleSchema, BulkDeleteResult, BulkUpdateResult, TagCount} from './api/types'
import {applyTheme, DEFAULT_CONFIG, type ThemeConfig} from './theme'
import {useModuleStore} from './stores/moduleStore'
import {useCollectionStore} from './stores/collectionStore'
import {useToastStore} from './stores/toastStore'
import {useSelectionStore} from './stores/selectionStore'
import {useSmartFolderStore, type SmartFolder} from './stores/smartFolderStore'
import AppSidebar from './components/AppSidebar.vue'
const DynamicForm = lazyComponent(() => import('./components/DynamicForm.vue'))
import {draftAtRisk, mayLeaveDraft} from './composables/draftGuard'
import ItemList from './components/ItemList.vue'
import CollectionGrid from './components/CollectionGrid.vue'
import ImageLightbox from './components/ImageLightbox.vue'
const SchemaBuilder = lazyComponent(() => import('./components/SchemaBuilder.vue'))
const ItemDetail = lazyComponent(() => import('./components/ItemDetail.vue'))
const SettingsPage = lazyComponent(() => import('./components/SettingsPage.vue'))
import ToastProvider from './components/ToastProvider.vue'
import FilterBar from './components/FilterBar.vue'
import CommandPalette from './components/CommandPalette.vue'
import ContextMenu from './components/ContextMenu.vue'
import type {MenuOption} from './components/ContextMenu.vue'
import BulkActionBar from './components/BulkActionBar.vue'
import DeleteConfirmation from './components/DeleteConfirmation.vue'
import RecoveryDialog from './components/RecoveryDialog.vue'
import CollectionPaging from './components/CollectionPaging.vue'
import CollectionSummary from './components/CollectionSummary.vue'
import {useSummaryStore} from './stores/summaryStore'
import TagFilter from './components/TagFilter.vue'
import TagManager from './components/TagManager.vue'
import ActionConfirmation from './components/ActionConfirmation.vue'
import ModalSurface from './components/ModalSurface.vue'
import ImportDialog from './components/ImportDialog.vue'
const DashboardView = lazyComponent(() => import('./components/DashboardView.vue'))
const ComparisonView = lazyComponent(() => import('./components/ComparisonView.vue'))
import {isAuthConfigured} from './auth/plugin'
import {beginSignOut, sessionFailure, captureSession} from './auth/session'
import {useAuth0} from '@auth0/auth0-vue'

const moduleStore = useModuleStore()
const collectionStore = useCollectionStore()
const summaryStore = useSummaryStore()
const toastStore = useToastStore()
const selectionStore = useSelectionStore()
const smartFolderStore = useSmartFolderStore()
const settingsLoading = ref(false)
const settingsError = ref('')
const collectionAvailable = computed(() => collectionStore.loaded && moduleStore.loaded && !collectionStore.loading && !moduleStore.loading && !collectionStore.error && !moduleStore.error)
const resultsFiltered = computed(() => !!(collectionStore.searchQuery.trim() || collectionStore.activeTags.length || Object.keys(collectionStore.activeFilters).length))

// Auth
const authEnabled = isAuthConfigured
const auth0 = authEnabled ? useAuth0() : null
async function onSignOut() {
  if (!auth0 || !canLeaveDraft()) return
  beginSignOut()
  try {await auth0.logout({logoutParams: {returnTo: window.location.origin}})}
  catch {sessionFailure.value = 'Sign out could not complete. Reload before continuing.'}
}
const owner = captureSession()
function currentSession() {try {owner.assertCurrent(); return true} catch {return false}}

// Tag state
const allTags = ref<TagCount[]>([])
const showTagManager = ref(false)
const tagManagerRef = ref<InstanceType<typeof TagManager> | null>(null)
const tagLoadError = ref('')
const mutation = ref<{title: string; description: string; label: string; run: () => Promise<string>; success?: () => void} | null>(null)
const mutationUncertain = ref(false)
function reloadPage() {window.location.reload()}

async function refreshTags() {
  if (!currentSession()) return false
  try {const tags = await api.getAllTags(); if (!currentSession()) return false; allTags.value = tags; tagLoadError.value = ''; return true}
  catch {if (currentSession()) tagLoadError.value = 'Tags could not be loaded. Counts may be out of date.'; return false}
}

function onTagRename(payload: {oldName: string, newName: string}) {
  if (mutation.value || mutationUncertain.value) return
  const {oldName, newName} = payload
  const merge = allTags.value.some(tag => tag.name === newName)
  mutation.value = {
    title: merge ? 'Merge tags?' : 'Rename tag?',
    description: `Change "${oldName}" to "${newName}" on all matching items in every collection. ${merge ? 'The destination already exists: both tags will be merged. This cannot be automatically undone.' : 'Saved-view tag filters are not renamed.'}`,
    label: merge ? 'Merge tags' : 'Rename tag',
    run: async () => {await api.renameTag(oldName, newName); return 'Tag change saved.'},
    success: () => {tagManagerRef.value?.cancelEdit(); collectionStore.activeTags = [...new Set(collectionStore.activeTags.map(tag => tag === oldName ? newName : tag))]},
  }
}

function onTagDelete(name: string) {
  if (mutation.value || mutationUncertain.value) return
  mutation.value = {
    title: `Delete tag "${name}"?`,
    description: 'Remove this tag from every item in every collection. Items will not be deleted. There is no automatic Undo; export a backup first if you may need these tag assignments.',
    label: 'Delete tag',
    run: async () => {await api.deleteTag(name); return 'Tag removed.'},
    success: () => {collectionStore.activeTags = collectionStore.activeTags.filter(tag => tag !== name)},
  }
}

async function onMutationConfirmed(message: string) {
  mutation.value?.success?.()
  mutation.value = null
  selectionStore.clear()
  const [itemsOK, tagsOK, summaryOK] = await Promise.all([collectionStore.fetchItems(), refreshTags(), summaryStore.refresh()])
  toastStore.show(itemsOK && tagsOK && summaryOK ? message : message + ' The view could not fully refresh. Reload; do not repeat the change.', itemsOK && tagsOK && summaryOK ? 'success' : 'error')
}

// Bulk action state
const showBulkDeleteConfirm = ref(false)
const showRecovery = ref(false)
function openRecovery() {
 if (!canLeaveDraft()) return
 showForm.value = false
 showRecovery.value = true
}
async function refreshRecoveryCollection() {
 const [refreshed, tagsRefreshed, summaryOK] = await Promise.all([collectionStore.fetchItems(), refreshTags(), summaryStore.refresh()])
 if (!refreshed || !tagsRefreshed || !summaryOK) throw new Error('Items, tags or totals could not be refreshed')
}
const deleteTargets = ref<{id: string; title: string}[]>([])
const showBulkModuleDialog = ref(false)
const bulkTargetModuleId = ref('')
const bulkTargets = ref<{id: string; title: string}[]>([])

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
  applyTheme(getEffectiveDark())
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
const formRef = ref<InstanceType<typeof DynamicForm> | null>(null)
const editorKey = ref(0)
const savingItem = ref(false)

const builderRef = ref<InstanceType<typeof SchemaBuilder> | null>(null)
const settingsRef = ref<InstanceType<typeof SettingsPage> | null>(null)
const builderKey = ref(0)
const settingsKey = ref(0)
function canLeaveDraft(): boolean {
  if (showRecovery.value || mutation.value || showBulkModuleDialog.value) return false
  if (showTagManager.value && tagManagerRef.value?.canLeave() === false) return false
  if (showSettings.value) {
    if (settingsRef.value?.canLeave() === false) return false
    showSettings.value = false
  }
  if (showBuilder.value) {
    if (builderRef.value?.canLeave() === false) return false
    showBuilder.value = false
  }
  if (!showForm.value) return true
  if (savingItem.value || formRef.value?.submitting) {
    toastStore.show('Wait for the save to finish before leaving.', 'info')
    return false
  }
  return mayLeaveDraft({dirty: !!formRef.value?.dirty, working: !!formRef.value?.working, submitting: savingItem.value},
    () => window.confirm('Discard this unsaved draft and any pending image work?'))
}
function onBeforeUnload(event: BeforeUnloadEvent) {
  if (showForm.value && draftAtRisk({dirty: !!formRef.value?.dirty, working: !!formRef.value?.working, submitting: savingItem.value})) {
    event.preventDefault()
    event.returnValue = ''
  }
}
onMounted(() => window.addEventListener('beforeunload', onBeforeUnload))
onUnmounted(() => window.removeEventListener('beforeunload', onBeforeUnload))
function openImport() {
  if (settingsLoading.value || smartFolderStore.saving) {toastStore.show('Wait for settings to finish before importing.', 'info'); return}
  if (!canLeaveDraft()) return
  importRefreshError.value = ''
  showForm.value = false
  showImportDialog.value = true
}
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
const importRefreshError = ref('')
const importRefreshing = ref(false)
const importRef = ref<InstanceType<typeof ImportDialog> | null>(null)
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
    if (!canLeaveDraft()) return
    editorKey.value++
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

function requestDeletion(items: Item[]) {
  if (showBulkDeleteConfirm.value || !items.length || !canLeaveDraft()) return
  if (items.length > 500) {toastStore.show('Select at most 500 items per deletion.', 'error'); return}
  deleteTargets.value = items.map(({id, title}) => ({id, title}))
  showBulkDeleteConfirm.value = true
}
function onDeleteItem(item: Item) {requestDeletion([item])}

// Keyboard shortcuts
function onGlobalKeydown(e: KeyboardEvent) {
  if (showRecovery.value || showBulkDeleteConfirm.value || mutation.value || showBulkModuleDialog.value) return
  const mod = e.metaKey || e.ctrlKey
  if (showImportDialog.value && e.key !== 'Escape') return

  if (mod && e.key === 'k') {
    e.preventDefault()
    showPalette.value = !showPalette.value
    return
  }

  if (e.key === 'Escape') {
    if (showImportDialog.value) { importRef.value?.requestClose(); return }
    if (showPalette.value) { showPalette.value = false; return }
    if (showComparison.value) { onCloseComparison(); return }
    if (lightboxVisible.value) { lightboxVisible.value = false; return }
    if (ctxVisible.value) { ctxVisible.value = false; return }
    if (showForm.value) { onCancel(); return }
    if (showDetail.value) { onCloseDetail(); return }
    if (showTagManager.value) { if (canLeaveDraft()) showTagManager.value = false; return }
    if (showBuilder.value) { canLeaveDraft(); return }
    if (showSettings.value) { onSettingsClose(); return }
    return
  }

  if (mod && e.key === 'f') {
    e.preventDefault()
    if (!canLeaveDraft()) return
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

onMounted(() => document.addEventListener('keydown', onGlobalKeydown))

async function loadSettings(): Promise<boolean> {
  if (settingsLoading.value || smartFolderStore.saving || !currentSession()) return false
  settingsLoading.value = true
  settingsError.value = ''
  smartFolderStore.invalidateBaseline()
  try {
    const settings = await api.get<any>('/api/v1/settings')
    if (!currentSession()) return false
    if (!smartFolderStore.loadFromSettings(settings)) throw new Error('Stored saved views are invalid')
    if (settings.theme && !['light', 'dark', 'system'].includes(settings.theme.mode)) throw new Error('Stored appearance is invalid')
    themeConfig.value = settings.theme ? {...DEFAULT_CONFIG, ...settings.theme} : {...DEFAULT_CONFIG}
    refreshTheme()
    return true
  } catch (e: any) {
    if (currentSession()) {
      smartFolderStore.invalidateBaseline()
      settingsError.value = 'Settings could not be loaded. Edits are disabled until retry succeeds. ' + (e?.message || '')
    }
    return false
  } finally {settingsLoading.value = false}
}
async function loadInitialData() {
  await loadSettings()
  if (!currentSession()) return
  await Promise.all([
    moduleStore.fetchModules(),
    summaryStore.refresh(),
    collectionStore.fetchItems(),
    refreshTags(),
  ])
}

onMounted(loadInitialData)

// Sidebar navigation: clicking a module filters the view (does NOT open form)
function onNavigate(moduleId: string) {
  if (!canLeaveDraft()) return
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
  if (!canLeaveDraft()) return
  editorKey.value++
  selectionStore.clear()
  selectedSchema.value = mod
  editingItem.value = null
  viewingItem.value = null
  showForm.value = true
  showDetail.value = false
  showBuilder.value = false
}

function onItemSelect(item: Item) {
  if (!canLeaveDraft()) return
  selectionStore.clear()
  const schema = moduleStore.getModuleById(item.moduleId)
  viewingItem.value = item
  viewingSchema.value = schema ?? null
  showDetail.value = true
  showForm.value = false
  showBuilder.value = false
}

function onEditFromDetail() {
  if (!canLeaveDraft()) return
  editorKey.value++
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

async function onSave(item: Item): Promise<boolean> {
  if (savingItem.value) return false
  savingItem.value = true
  try {
    const saved = await collectionStore.saveItem(item)
    showForm.value = false
    editingItem.value = null
    if (saved) {
      viewingItem.value = saved
      viewingSchema.value = moduleStore.getModuleById(saved.moduleId) ?? null
      showDetail.value = true
      toastStore.show('Item saved', 'success')
      void summaryStore.refresh()
      refreshTags()
    }
  } finally {
    savingItem.value = false
  }
  return true
}

function onDeleteFromDetail() {
  if (!viewingItem.value) return
  onDeleteItem(viewingItem.value)
}

function onCancel() {
  if (!canLeaveDraft()) return
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
  if (!canLeaveDraft()) return
  smartFolderStore.setActive(folder.id)
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
    if (mutation.value || mutationUncertain.value || smartFolderStore.saving || !canLeaveDraft()) return
    const id = folder.id
    mutation.value = {
      title: `Delete saved view "${folder.name}"?`,
      description: 'Only this saved view is removed. Your collection items are not deleted. There is no automatic Undo.',
      label: 'Delete saved view',
      run: async () => {await smartFolderStore.removePersisted(id); return 'Saved view deleted.'},
    }
  }
}

// Search (from ItemList)
function clearResultFilters() {
  smartFolderStore.clearActive()
  collectionStore.activeFilters = {}
  collectionStore.activeTags = []
  collectionStore.setSearch('')
}
function onSearch(query: string) {
  smartFolderStore.clearActive()
  collectionStore.setSearch(query, 300)
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
    openImport()
  }
}

// Bulk actions
function onBulkDelete() {
  const ids = new Set(selectionStore.selectedIdArray())
  requestDeletion(collectionStore.items.filter(item => ids.has(item.id)))
}
async function removeConfirmed(ids: string[]): Promise<number> {
  const result = await api.post<BulkDeleteResult>('/api/v1/items/batch-delete', {ids})
  return result.deleted
}
async function onDeleted(count: number, ids: string[]) {
  showBulkDeleteConfirm.value = false
  deleteTargets.value = []
  collectionStore.items = collectionStore.items.filter(item => !ids.includes(item.id))
  selectionStore.prune(collectionStore.items)
  if (viewingItem.value && ids.includes(viewingItem.value.id)) onCloseDetail()
  const outcomes = await Promise.all([collectionStore.fetchItems(), refreshTags(), summaryStore.refresh()])
  const refreshed = outcomes.every(Boolean)
  toastStore.show(refreshed ? `${count} item(s) deleted. Undo from Deletion recovery in the sidebar.` : `${count} item(s) deleted, but the view could not refresh.`,
    refreshed ? 'success' : 'error')
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

function openBulkModule() {
  if (mutationUncertain.value || !canLeaveDraft()) return
  const ids = selectionStore.selectedIdArray()
  if (!ids.length || ids.length > 500) {toastStore.show('Select 1–500 items per move.', 'error'); return}
  const items = collectionStore.items.filter(item => ids.includes(item.id))
  if (items.length !== ids.length) {toastStore.show('Selection is out of date. Refresh before moving.', 'error'); return}
  bulkTargets.value = items.map(({id, title}) => ({id, title}))
  bulkTargetModuleId.value = ''
  showBulkModuleDialog.value = true
}
function onBulkUpdateModule() {
  const target = moduleStore.modules.find(module => module.id === bulkTargetModuleId.value)
  if (!target || mutationUncertain.value) return
  const ids = bulkTargets.value.map(item => item.id)
  const newModuleId = target.id
  showBulkModuleDialog.value = false
  mutation.value = {
    title: `Move ${ids.length} item(s) to ${target.displayName}?`,
    description: `Selected when this dialog opened: ${bulkTargets.value.slice(0, 5).map(item => item.title).join(', ')}${ids.length > 5 ? ', and more' : ''}. Existing attributes are retained. Incompatible items reject the entire move; there is no automatic Undo. Export a backup first if needed.`,
    label: 'Move items',
    run: async () => {const result = await api.post<BulkUpdateResult>('/api/v1/items/batch-update-module', {ids, newModuleId}); return `${result.updated} item(s) moved.`},
  }
}

function onCompare() {
  if (!canLeaveDraft()) return
  showForm.value = false
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
  // Keep committed counts visible; retrying this refresh never repeats restore.
  if (importRefreshing.value) return
  importRefreshing.value = true
  importRefreshError.value = ''
  selectionStore.clear()
  try {
    const [, refreshed, tagsOK, settingsOK, summaryOK] = await Promise.all([
      moduleStore.fetchModules(), collectionStore.fetchItems(), refreshTags(), loadSettings(), summaryStore.refresh(),
    ])
    if (!refreshed || moduleStore.error || !tagsOK || !settingsOK || !summaryOK) {
      importRefreshError.value = 'The backup was imported, but the collection view could not refresh.'
    }
  } catch {
    importRefreshError.value = 'The backup was imported, but the collection view could not refresh.'
  } finally {
    importRefreshing.value = false
  }
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
  if (!canLeaveDraft()) return
  builderKey.value++
  builderModuleId.value = null
  builderInitialJSON.value = null
  showBuilder.value = true
  showForm.value = false
}

async function openEditSchemaBuilder(mod: ModuleSchema) {
  try {
    const data = await api.get<any>('/api/v1/modules/' + mod.id + '/file')
    if (!currentSession()) return
    if (!canLeaveDraft()) return
    builderKey.value++
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
  if (!canLeaveDraft()) return
  selectionStore.clear()
  showTagManager.value = true
  showForm.value = false
  showDetail.value = false
  showBuilder.value = false
  showSettings.value = false
  refreshTags()
}

function openSettings() {
  if (settingsLoading.value || settingsError.value || !smartFolderStore.baselineReady) {toastStore.show('Load settings successfully before editing appearance.', 'info'); return}
  if (!canLeaveDraft()) return
  settingsKey.value++
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
  if (!canLeaveDraft()) return
  refreshTheme()
  showSettings.value = false
}
</script>

<template>
  <div>
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
      @importBackup="openImport"
      @openRecovery="openRecovery"
      @openTags="openTagManager"
      @openSettings="openSettings"
      @signOut="onSignOut"
    />

    <main class="main-content" :class="{'has-selection': collectionAvailable && selectionStore.count > 0}">
      <p v-if="mutationUncertain" role="alert">A change was not confirmed. Reload to inspect the result before more tag, saved-view deletion or move operations. <button @click="reloadPage">Reload</button></p>
      <p v-if="settingsLoading" role="status">Loading settings…</p>
      <p v-else-if="settingsError" role="alert">{{ settingsError }} <button @click="loadSettings">Retry settings</button></p>
      <div v-if="moduleStore.loading || collectionStore.loading || (!collectionStore.loaded && !collectionStore.error)" class="loading" role="status">
        Loading...
      </div>

      <div v-if="collectionStore.error" role="alert" class="error-message">
        Collection results unavailable: {{ collectionStore.error }}
        <button :disabled="collectionStore.loading" @click="collectionStore.fetchItems()">Retry items</button>
      </div>
      <div v-if="moduleStore.error" role="alert" class="error-message">
        Collection types unavailable: {{ moduleStore.error }}
        <button :disabled="moduleStore.loading" @click="moduleStore.fetchModules()">Retry collection types</button>
      </div>

      <Transition name="fade-slide" mode="out-in">
        <SettingsPage
          ref="settingsRef"
          :key="settingsKey"
          v-if="showSettings"
          :initialConfig="themeConfig"
          :systemDark="systemDark"
          @saved="onSettingsSaved"
          @close="showSettings = false"
        />
      </Transition>

      <Transition name="fade-slide" mode="out-in">
        <TagManager
          v-if="showTagManager && !showSettings"
          ref="tagManagerRef"
          :tags="allTags"
          :disabled="mutationUncertain"
          :load-error="tagLoadError"
          @reload="refreshTags"
          @rename="onTagRename"
          @delete="onTagDelete"
          @close="() => {if (canLeaveDraft()) showTagManager = false}"
        />
      </Transition>

      <Transition name="fade-slide" mode="out-in">
        <SchemaBuilder
          ref="builderRef"
          :key="builderKey"
          v-if="showBuilder && !showSettings && !showTagManager"
          :moduleId="builderModuleId"
          :initialJSON="builderInitialJSON"
          @saved="onBuilderSaved"
          @close="showBuilder = false"
        />
      </Transition>

      <Transition name="fade-slide" mode="out-in">
        <DynamicForm
          ref="formRef"
          :key="editorKey"
          v-if="showForm && selectedSchema && !showBuilder && !showSettings && !showTagManager"
          :schema="selectedSchema"
          :item="editingItem"
          :save-item="onSave"
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

        <CollectionSummary />
        <CollectionPaging />
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

        <div v-if="collectionAvailable && resultsFiltered" class="filtered-empty">
          <span v-if="collectionStore.items.length === 0 && collectionStore.offset === 0">No items match the current search or filters.</span>
          <span v-if="collectionStore.searchQuery">Search: “{{ collectionStore.searchQuery }}”</span>
          <button class="filtered-empty-clear" @click="clearResultFilters">Clear search and filters</button>
        </div>

        <Transition v-if="collectionStore.offset === 0 || collectionStore.items.length > 0 || !collectionAvailable" name="fade-slide" mode="out-in">
          <DashboardView
            v-if="showDashboard && !collectionStore.activeModuleId && collectionAvailable"
            key="dashboard"
            :items="collectionStore.items"
            :modules="moduleStore.modules"
            :dark="getEffectiveDark()"
            :filtered="resultsFiltered"
            @selectItem="onDashboardSelectItem"
          />

          <ItemList
            v-else-if="(!showDashboard || !!collectionStore.activeModuleId) && viewMode === 'list'"
            key="list"
            ref="itemListRef"
            :items="collectionStore.items"
            :modules="moduleStore.modules"
            :activeModuleId="collectionStore.activeModuleId"
            :search-query="collectionStore.searchQuery"
            :unavailable="!collectionAvailable"
            :filtered="resultsFiltered"
            @select="onItemSelect"
            @filterChange="onNavigate"
            @search="onSearch"
            @addItem="onAddFirstItem"
            @itemContextMenu="onItemContextMenu"
          />

          <CollectionGrid
            v-else-if="(!showDashboard || !!collectionStore.activeModuleId) && viewMode === 'grid'"
            key="grid"
            :unavailable="!collectionAvailable"
            :filtered="resultsFiltered"
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
      v-if="collectionAvailable"
      :count="selectionStore.count"
      @delete="onBulkDelete"
      @export="onBulkExportCSV"
      @editModule="openBulkModule"
      @deselectAll="selectionStore.clear()"
      @compare="onCompare"
    />

    <RecoveryDialog v-if="showRecovery" :refresh="refreshRecoveryCollection" @close="showRecovery = false" />
    <DeleteConfirmation
      v-if="showBulkDeleteConfirm"
      :items="deleteTargets"
      :remove="removeConfirmed"
      @close="showBulkDeleteConfirm = false; deleteTargets = []"
      @deleted="onDeleted"
    />

    <ActionConfirmation v-if="mutation" :action="mutation" @close="mutation = null" @confirmed="onMutationConfirmed" @uncertain="mutationUncertain = true" />

    <!-- Pick a destination before confirming a fixed-target move. -->
    <ModalSurface v-if="showBulkModuleDialog" label="Choose destination module" @close="showBulkModuleDialog = false">
      <div class="confirm-overlay" @click.self="showBulkModuleDialog = false">
        <div class="confirm-dialog">
          <p class="confirm-title">Move {{ bulkTargets.length }} item(s) to:</p>
          <select v-model="bulkTargetModuleId" class="bulk-module-select" aria-label="Destination module">
            <option value="" disabled>Select module...</option>
            <option v-for="mod in moduleStore.modules" :key="mod.id" :value="mod.id">{{ mod.displayName }}</option>
          </select>
          <div class="confirm-actions">
            <button autofocus class="confirm-cancel-btn" @click="showBulkModuleDialog = false">Cancel</button>
            <button class="confirm-delete-btn" :disabled="!bulkTargetModuleId" @click="onBulkUpdateModule" style="background: var(--accent-blue)">Review move</button>
          </div>
        </div>
      </div>
    </ModalSurface>

    <ImportDialog
      ref="importRef"
      :refresh-error="importRefreshError"
      :refreshing="importRefreshing"
      @refresh="onImported"
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
  </div>
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
  height: 100dvh;
}
@media (max-width: 767px) {
  .app-layout { flex-direction: column; }
  .app-layout > .main-content { padding: 16px; border-radius: 0; }
}
.main-content {
  flex: 1;
  min-width: 0;
  min-height: 0;
  padding: 24px;
  overflow-y: auto;
  background-color: var(--bg-primary);
  background-image:
    linear-gradient(to right, var(--grid-color) 1px, transparent 1px),
    linear-gradient(to bottom, var(--grid-color) 1px, transparent 1px);
  background-size: var(--grid-size) var(--grid-size);
  border-top-left-radius: var(--radius-lg);
  position: relative;
  z-index: 1;
}
.main-content.has-selection { padding-bottom: 224px; }
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
