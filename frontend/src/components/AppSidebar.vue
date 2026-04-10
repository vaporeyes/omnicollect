<script lang="ts" setup>
// ABOUTME: Application sidebar with collection navigation, smart folders, and action buttons.
// ABOUTME: Clicking a module filters the view; dedicated buttons for creating items and schemas.

import {ref, computed, onMounted} from 'vue'
import type {ModuleSchema, Showcase, TagCount} from '../api/types'
import {toggleShowcase, listShowcases, getAIStatus} from '../api/client'
import {useModuleStore} from '../stores/moduleStore'
import {useCollectionStore} from '../stores/collectionStore'
import {useSmartFolderStore, type SmartFolder} from '../stores/smartFolderStore'
import SmartFolders from './SmartFolders.vue'

const props = defineProps<{
  exporting: boolean
  authEnabled: boolean
}>()

const emit = defineEmits<{
  (e: 'navigate', moduleId: string): void
  (e: 'newItem', mod: ModuleSchema): void
  (e: 'newSchema'): void
  (e: 'editSchema', mod: ModuleSchema): void
  (e: 'applySmartFolder', folder: SmartFolder): void
  (e: 'smartFolderContextMenu', folder: SmartFolder, x: number, y: number): void
  (e: 'exportBackup'): void
  (e: 'importBackup'): void
  (e: 'openTags'): void
  (e: 'openSettings'): void
  (e: 'signOut'): void
}>()

const moduleStore = useModuleStore()
const collectionStore = useCollectionStore()
const smartFolderStore = useSmartFolderStore()

const smartFoldersRef = ref<InstanceType<typeof SmartFolders> | null>(null)

// Showcase state per module
const showcaseMap = ref<Record<string, Showcase>>({})
const copiedSlug = ref<string | null>(null)
const isCloudMode = ref(false)

onMounted(async () => {
  try {
    const status = await getAIStatus()
    isCloudMode.value = status.cloudMode === true
    if (isCloudMode.value) {
      const showcases = await listShowcases()
      for (const sc of showcases) {
        showcaseMap.value[sc.moduleId] = sc
      }
    }
  } catch {
    isCloudMode.value = false
  }
})

async function onToggleShowcase(mod: ModuleSchema) {
  const current = showcaseMap.value[mod.id]
  const newEnabled = !current?.enabled
  try {
    const result = await toggleShowcase(mod.id, newEnabled)
    showcaseMap.value[mod.id] = result
  } catch (e) {
    console.error('Failed to toggle showcase:', e)
  }
}

function copyShowcaseUrl(mod: ModuleSchema) {
  const sc = showcaseMap.value[mod.id]
  if (!sc?.url) return
  const url = window.location.origin + sc.url
  navigator.clipboard.writeText(url)
  copiedSlug.value = sc.slug
  setTimeout(() => { copiedSlug.value = null }, 2000)
}

// Item counts per module (computed from loaded items)
const moduleItemCounts = computed(() => {
  const counts: Record<string, number> = {}
  for (const item of collectionStore.items) {
    counts[item.moduleId] = (counts[item.moduleId] ?? 0) + 1
  }
  return counts
})

const totalItemCount = computed(() => collectionStore.items.length)

function startRename(id: string) {
  smartFoldersRef.value?.startRename(id)
}

defineExpose({startRename})
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar-brand">
      <h2>OmniCollect</h2>
    </div>

    <div class="sidebar-scroll">
      <!-- Navigation: All Types -->
      <div class="nav-section">
        <div class="nav-section-header">
          <span class="nav-section-title">Browse</span>
        </div>
        <div
          class="nav-item"
          :class="{active: !collectionStore.activeModuleId}"
          @click="emit('navigate', '')"
        >
          <svg class="nav-icon" viewBox="0 0 16 16" fill="none"><rect x="2" y="2" width="5" height="5" rx="1" stroke="currentColor" stroke-width="1.2"/><rect x="9" y="2" width="5" height="5" rx="1" stroke="currentColor" stroke-width="1.2"/><rect x="2" y="9" width="5" height="5" rx="1" stroke="currentColor" stroke-width="1.2"/><rect x="9" y="9" width="5" height="5" rx="1" stroke="currentColor" stroke-width="1.2"/></svg>
          <span class="nav-label">All Types</span>
          <span v-if="totalItemCount > 0" class="nav-count">{{ totalItemCount }}</span>
        </div>

        <!-- Navigation: Per-module -->
        <div
          v-for="mod in moduleStore.modules"
          :key="mod.id"
          class="nav-item"
          :class="{active: collectionStore.activeModuleId === mod.id}"
          @click="emit('navigate', mod.id)"
        >
          <svg class="nav-icon" viewBox="0 0 16 16" fill="none"><path d="M2 4.5A1.5 1.5 0 013.5 3h3.172a1.5 1.5 0 011.06.44l.768.767a1.5 1.5 0 001.06.439H12.5A1.5 1.5 0 0114 6.146V11.5a1.5 1.5 0 01-1.5 1.5h-9A1.5 1.5 0 012 11.5V4.5z" stroke="currentColor" stroke-width="1.2"/></svg>
          <span class="nav-label">{{ mod.displayName }}</span>
          <span class="nav-item-actions">
            <button
              class="nav-action-btn add-btn"
              @click.stop="emit('newItem', mod)"
              title="Add item"
            >+</button>
            <button
              v-if="isCloudMode"
              class="nav-action-btn"
              :class="{'action-active': showcaseMap[mod.id]?.enabled}"
              @click.stop="onToggleShowcase(mod)"
              :title="showcaseMap[mod.id]?.enabled ? 'Make private' : 'Make public'"
            >
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M4 12v8a2 2 0 002 2h12a2 2 0 002-2v-8"/><polyline points="16 6 12 2 8 6"/><line x1="12" y1="2" x2="12" y2="15"/></svg>
            </button>
            <button
              v-if="isCloudMode && showcaseMap[mod.id]?.enabled"
              class="nav-action-btn"
              @click.stop="copyShowcaseUrl(mod)"
              :title="copiedSlug === showcaseMap[mod.id]?.slug ? 'Copied!' : 'Copy link'"
            >{{ copiedSlug === showcaseMap[mod.id]?.slug ? '&#10003;' : '&#128279;' }}</button>
            <button
              class="nav-action-btn"
              @click.stop="emit('editSchema', mod)"
              title="Edit schema"
            >&#9998;</button>
          </span>
          <span v-if="moduleItemCounts[mod.id]" class="nav-count">{{ moduleItemCounts[mod.id] }}</span>
        </div>
      </div>

      <!-- Smart Folders -->
      <SmartFolders
        ref="smartFoldersRef"
        :activeModuleId="collectionStore.activeModuleId"
        :searchQuery="collectionStore.searchQuery"
        :activeFilters="collectionStore.activeFilters"
        :activeTags="collectionStore.activeTags"
        @apply="(f: SmartFolder) => emit('applySmartFolder', f)"
        @contextMenu="(f: SmartFolder, x: number, y: number) => emit('smartFolderContextMenu', f, x, y)"
      />
    </div>

    <div class="sidebar-bottom">
      <button class="sidebar-action-btn primary-action" @click="emit('newSchema')">+ New Schema</button>
      <button class="sidebar-action-btn" :disabled="exporting" @click="emit('exportBackup')">
        {{ exporting ? 'Exporting...' : 'Export Backup' }}
      </button>
      <button class="sidebar-action-btn" @click="emit('importBackup')">Import Backup</button>
      <button class="sidebar-action-btn" @click="emit('openTags')">Tags</button>
      <button class="sidebar-action-btn" @click="emit('openSettings')">Settings</button>
      <button v-if="authEnabled" class="sidebar-action-btn signout" @click="emit('signOut')">Sign Out</button>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 250px;
  padding: 0;
  background: var(--bg-tertiary);
  -webkit-backdrop-filter: blur(var(--glass-blur));
  backdrop-filter: blur(var(--glass-blur));
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-brand {
  padding: 20px 16px 12px;
}

.sidebar-brand h2 {
  margin: 0;
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
  padding: 0 8px;
}

/* Navigation section */
.nav-section {
  margin-bottom: 4px;
}

.nav-section-header {
  padding: 0 8px 4px;
}

.nav-section-title {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide);
  color: var(--text-muted);
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary);
  position: relative;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.nav-item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.nav-item:hover .nav-item-actions {
  opacity: 1;
}

.nav-item.active {
  background: var(--accent-blue-light);
  color: var(--accent-blue);
  font-weight: 500;
}

.nav-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}

.nav-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.nav-count {
  font-size: 11px;
  color: var(--text-muted);
  font-weight: 400;
  min-width: 16px;
  text-align: right;
}

.nav-item-actions {
  display: flex;
  align-items: center;
  gap: 1px;
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.nav-action-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 12px;
  color: var(--text-muted);
  padding: 1px 3px;
  line-height: 1;
  transition: color var(--transition-fast);
  border-radius: 3px;
}

.nav-action-btn:hover {
  color: var(--accent-blue);
  background: var(--accent-blue-light);
}

.nav-action-btn.add-btn {
  font-size: 14px;
  font-weight: 600;
}

.nav-action-btn.action-active {
  color: var(--accent-blue);
}

/* Bottom actions */
.sidebar-bottom {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 8px;
  border-top: 1px solid var(--border-primary);
}

.sidebar-action-btn {
  width: 100%;
  padding: 7px 12px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 12px;
  text-align: left;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.sidebar-action-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.sidebar-action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.sidebar-action-btn.primary-action {
  text-align: center;
  border: 1px dashed var(--accent-blue);
  color: var(--accent-blue);
  font-weight: 500;
  margin-bottom: 4px;
}

.sidebar-action-btn.primary-action:hover {
  background: var(--accent-blue-light);
  border-color: var(--accent-blue-hover);
}

.sidebar-action-btn.signout:hover {
  color: var(--accent-danger, #e74c3c);
}
</style>
