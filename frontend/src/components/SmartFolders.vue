<script lang="ts" setup>
// ABOUTME: Sidebar section for Smart Folders (saved collection views).
// ABOUTME: Displays folder list with save, apply, rename, and delete support.

import {ref, nextTick} from 'vue'
import {menuAnchor} from '../composables/menuAnchor'
import {useSmartFolderStore, type SmartFolder} from '../stores/smartFolderStore'
import type {AttributeFilter} from '../stores/collectionStore'

const props = defineProps<{
  activeModuleId: string
  searchQuery: string
  activeFilters: Record<string, AttributeFilter[]>
  activeTags: string[]
}>()

const emit = defineEmits<{
  (e: 'apply', folder: SmartFolder): void
  (e: 'contextMenu', folder: SmartFolder, x: number, y: number): void
}>()

const store = useSmartFolderStore()

// Inline naming state for save
const showSaveInput = ref(false)
const saveInputValue = ref('')
const saveInputRef = ref<HTMLInputElement | null>(null)

// Inline rename state
const renamingId = ref<string | null>(null)
const renameValue = ref('')
const renameInputRef = ref<HTMLInputElement | null>(null)

function startSave() {
  showSaveInput.value = true
  saveInputValue.value = ''
  nextTick(() => saveInputRef.value?.focus())
}

function confirmSave() {
  if (!saveInputValue.value.trim()) return
  store.create(
    saveInputValue.value,
    props.activeModuleId,
    props.searchQuery,
    props.activeFilters,
    props.activeTags
  )
  showSaveInput.value = false
  saveInputValue.value = ''
}

function cancelSave() {
  showSaveInput.value = false
  saveInputValue.value = ''
}

function onFolderClick(folder: SmartFolder) {
  emit('apply', folder)
}

function onFolderContextMenu(e: MouseEvent, folder: SmartFolder) {
  const {x, y} = menuAnchor(e)
  emit('contextMenu', folder, x, y)
}

// Called externally when context menu selects "Rename"
function startRename(id: string) {
  const folder = store.folders.find(f => f.id === id)
  if (!folder) return
  renamingId.value = id
  renameValue.value = folder.name
  nextTick(() => renameInputRef.value?.focus())
}

function confirmRename() {
  if (!renamingId.value) return
  if (renameValue.value.trim()) {
    store.rename(renamingId.value, renameValue.value)
  }
  renamingId.value = null
  renameValue.value = ''
}

function cancelRename() {
  renamingId.value = null
  renameValue.value = ''
}

defineExpose({startRename})
</script>

<template>
  <div class="smart-folders">
    <div class="sf-header">
      <span class="sf-title">Saved Views</span>
    </div>

    <p v-if="!store.baselineReady" role="status" class="sf-empty">Saved views are unavailable until settings load. Creation and edits are disabled.</p>
    <p v-else-if="store.deletionUncertain" role="alert" class="sf-empty">Deletion was not confirmed. Reload to inspect saved views before making more changes.</p>
    <p v-else-if="store.saveError" role="alert" class="sf-empty">
      {{ store.saveError }}. The server may differ. Retry saves the current local views.
      <button :disabled="store.saving" @click="store.saveToSettings()">Retry save</button>
    </p>
    <p v-else-if="store.saving" role="status" class="sf-empty">Saving views...</p>
    <div v-if="store.baselineReady && store.folders.length === 0 && !showSaveInput" class="sf-empty">
      No saved views yet
    </div>

    <div v-if="store.baselineReady" class="sf-list">
      <div
        v-for="folder in store.folders"
        :key="folder.id"
        class="sf-item"
        :class="{active: store.activeSmartFolderId === folder.id}"
        role="group"
        :aria-label="folder.name"
        @contextmenu="onFolderContextMenu($event, folder)"
      >
        <svg class="sf-icon" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M3 2.5a.5.5 0 0 1 .5-.5h5.793a.5.5 0 0 1 .354.146l2.207 2.208a.5.5 0 0 1 .146.353V13.5a.5.5 0 0 1-.5.5h-8a.5.5 0 0 1-.5-.5v-11Z" stroke="currentColor" stroke-width="1.2"/>
          <path d="M5.5 7h5M5.5 9.5h3.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/>
        </svg>
        <input
          v-if="renamingId === folder.id"
          :ref="element => {renameInputRef = element as HTMLInputElement | null}"
          v-model="renameValue"
          class="sf-rename-input"
          aria-label="Saved view name"
          :disabled="!store.baselineReady || store.saving || store.deletionUncertain"
          @keydown.enter.stop="confirmRename"
          @keydown.esc.stop="cancelRename"
          @blur="confirmRename"
          @click.stop
        />
        <button v-else class="sf-name sf-link" :aria-pressed="store.activeSmartFolderId === folder.id" @click="onFolderClick(folder)">{{ folder.name }}</button>
        <button class="sf-link" aria-haspopup="menu" :aria-label="`Actions for saved view ${folder.name}`" @click="onFolderContextMenu($event, folder)">⋯</button>
      </div>
    </div>

    <div v-if="showSaveInput" class="sf-save-input-row">
      <input
        ref="saveInputRef"
        v-model="saveInputValue"
        class="sf-save-input"
        placeholder="View name..."
        @keydown.enter="confirmSave"
        @keydown.esc="cancelSave"
        @blur="cancelSave"
      />
    </div>

    <button v-if="!showSaveInput" class="sf-save-btn" :disabled="!store.baselineReady || store.saving || store.deletionUncertain" @click="startSave">
      Save Current View
    </button>
  </div>
</template>

<style scoped>
.sf-link {border: 0; background: transparent; color: inherit; font: inherit; cursor: pointer; text-align: left;}
.sf-name.sf-link {flex: 1;}
.smart-folders {
  padding: 8px 0;
  border-top: 1px solid var(--border-primary);
  margin-top: 8px;
}

.sf-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 4px;
  margin-bottom: 4px;
}

.sf-title {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide);
  color: var(--text-muted);
}

.sf-empty {
  padding: 8px 4px;
  font-size: 12px;
  color: var(--text-muted);
  font-style: italic;
}

.sf-list {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.sf-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary);
  transition: background var(--transition-fast), color var(--transition-fast);
}

.sf-item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.sf-item.active {
  background: var(--accent-blue-light);
  color: var(--accent-blue);
  font-weight: 500;
}

.sf-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}

.sf-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.sf-rename-input {
  flex: 1;
  min-width: 0;
  padding: 2px 4px;
  border: 1px solid var(--accent-blue);
  border-radius: 3px;
  background: var(--bg-input);
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
}

.sf-save-input-row {
  padding: 4px 0;
}

.sf-save-input {
  width: 100%;
  padding: 6px 8px;
  border: 1px solid var(--accent-blue);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  box-sizing: border-box;
}

.sf-save-input::placeholder {
  color: var(--text-muted);
}

.sf-save-btn {
  width: 100%;
  padding: 7px 12px;
  margin-top: 4px;
  border: 1px dashed var(--border-primary);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 12px;
  transition: background var(--transition-fast), color var(--transition-fast), border-color var(--transition-fast);
}

.sf-save-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--accent-blue);
}
</style>
