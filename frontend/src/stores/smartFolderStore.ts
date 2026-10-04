// ABOUTME: Pinia store for Smart Folders (saved collection views).
// ABOUTME: Persists named view state snapshots via the existing settings API.

import {defineStore} from 'pinia'
import {ref, onScopeDispose} from 'vue'
import * as api from '../api/client'
import type {AttributeFilter} from './collectionStore'

export interface SmartFolder {
  id: string
  name: string
  moduleId: string
  searchQuery: string
  filters: Record<string, AttributeFilter[]>
  tags: string[]
  createdAt: string
}

function generateId(): string {
  const bytes = new Uint8Array(4)
  crypto.getRandomValues(bytes)
  return Array.from(bytes, b => b.toString(16).padStart(2, '0')).join('')
}

export const useSmartFolderStore = defineStore('smartFolders', () => {
  const folders = ref<SmartFolder[]>([])
  const activeSmartFolderId = ref<string | null>(null)

  const baselineReady = ref(false)
  function invalidateBaseline() {baselineReady.value = false}
  function loadFromSettings(settings: any): boolean {
    if (disposed || saving.value) return false
    baselineReady.value = false
    if (!settings || typeof settings !== 'object' || Array.isArray(settings)) return false
    const entries = settings.smartFolders === undefined ? [] : settings.smartFolders
    if (!Array.isArray(entries) || entries.some(folder => !folder || typeof folder.id !== 'string' || !folder.id ||
      typeof folder.name !== 'string' || !folder.name ||
      (folder.tags !== undefined && (!Array.isArray(folder.tags) || folder.tags.some((tag: unknown) => typeof tag !== 'string'))) ||
      (folder.filters !== undefined && (!folder.filters || typeof folder.filters !== 'object' || Array.isArray(folder.filters) || Object.values(folder.filters).some(value => !Array.isArray(value) || value.some(filter => !filter || typeof filter !== 'object')))) ||
      (folder.moduleId !== undefined && typeof folder.moduleId !== 'string') ||
      (folder.searchQuery !== undefined && typeof folder.searchQuery !== 'string')) ||
      new Set(entries.map(folder => folder.id)).size !== entries.length) return false
    folders.value = JSON.parse(JSON.stringify(entries))
    activeSmartFolderId.value = null
    baselineReady.value = true
    return true
  }

  const saveError = ref<string | null>(null)
  const saving = ref(false)
  let pending = Promise.resolve()
  let saveVersion = 0
  let disposed = false
  let removing = false
  const deletionUncertain = ref(false)
  onScopeDispose(() => {disposed = true; baselineReady.value = false; saveVersion++; folders.value = []; activeSmartFolderId.value = null; saveError.value = null; saving.value = false})

  function saveToSettings(): Promise<void> {
    if (disposed || !baselineReady.value || deletionUncertain.value) return Promise.resolve()
    if (removing) return pending
    const version = ++saveVersion
    const snapshot = JSON.parse(JSON.stringify(folders.value))
    saving.value = true
    pending = pending.then(async () => {
      if (disposed) return
      try {
        if (!baselineReady.value) throw new Error('Load settings before saving views')
        await api.put('/api/v1/settings', {smartFolders: snapshot})
        if (version === saveVersion) saveError.value = null
      } catch (error: any) {
        if (version === saveVersion) saveError.value = error?.message || 'Saved views could not be saved'
      } finally {
        if (version === saveVersion) saving.value = false
      }
    })
    return pending
  }

  function create(
    name: string,
    moduleId: string,
    searchQuery: string,
    filters: Record<string, AttributeFilter[]>,
    tags: string[]
  ): SmartFolder | null {
    if (disposed || !baselineReady.value || removing || deletionUncertain.value) return null
    const trimmed = name.trim()
    if (!trimmed) return null

    const folder: SmartFolder = {
      id: generateId(),
      name: trimmed,
      moduleId,
      searchQuery,
      filters: JSON.parse(JSON.stringify(filters)),
      tags: [...tags],
      createdAt: new Date().toISOString(),
    }
    folders.value.push(folder)
    saveToSettings()
    return folder
  }

  function rename(id: string, newName: string): boolean {
    if (disposed || !baselineReady.value || removing || deletionUncertain.value) return false
    const trimmed = newName.trim()
    if (!trimmed) return false
    const folder = folders.value.find(f => f.id === id)
    if (!folder) return false
    folder.name = trimmed
    saveToSettings()
    return true
  }

  async function removePersisted(id: string) {
    if (disposed || !baselineReady.value || saving.value || deletionUncertain.value) throw new Error('Wait for saved views to finish saving, or reload after an uncertain deletion')
    if (!folders.value.some(folder => folder.id === id)) throw new Error('Saved view no longer exists')
    const snapshot = JSON.parse(JSON.stringify(folders.value.filter(folder => folder.id !== id)))
    removing = true
    saving.value = true
    const operation = pending.then(async () => {
      if (disposed) throw new DOMException('Account session changed', 'AbortError')
      await api.put('/api/v1/settings', {smartFolders: snapshot})
      if (disposed) throw new DOMException('Account session changed', 'AbortError')
      folders.value = snapshot
      if (activeSmartFolderId.value === id) activeSmartFolderId.value = null
      saveError.value = null
    })
    pending = operation.catch(() => {})
    try {await operation}
    catch (error: any) {
      if (!disposed) {deletionUncertain.value = true; saveError.value = error?.message || 'Deletion was not confirmed'}
      throw error
    } finally {removing = false; if (!disposed) saving.value = false}
  }

  function setActive(id: string | null) {
    activeSmartFolderId.value = id
  }

  function clearActive() {
    activeSmartFolderId.value = null
  }

  return {
    folders,
    activeSmartFolderId,
    loadFromSettings, baselineReady, invalidateBaseline,
    saveToSettings, saveError, saving,
    create,
    rename,
    remove: removePersisted, removePersisted, deletionUncertain,
    setActive,
    clearActive,
  }
})
