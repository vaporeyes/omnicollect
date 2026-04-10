// ABOUTME: Pinia store for Smart Folders (saved collection views).
// ABOUTME: Persists named view state snapshots via the existing settings API.

import {defineStore} from 'pinia'
import {ref} from 'vue'
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

  function loadFromSettings(settings: any) {
    if (settings?.smartFolders && Array.isArray(settings.smartFolders)) {
      folders.value = settings.smartFolders
    }
  }

  async function saveToSettings() {
    try {
      const current = await api.get<any>('/api/v1/settings')
      const updated = {...(current || {}), smartFolders: folders.value}
      await api.put('/api/v1/settings', updated)
    } catch {
      // Settings save failed silently; folders still in memory
    }
  }

  function create(
    name: string,
    moduleId: string,
    searchQuery: string,
    filters: Record<string, AttributeFilter[]>,
    tags: string[]
  ): SmartFolder | null {
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
    const trimmed = newName.trim()
    if (!trimmed) return false
    const folder = folders.value.find(f => f.id === id)
    if (!folder) return false
    folder.name = trimmed
    saveToSettings()
    return true
  }

  function remove(id: string) {
    folders.value = folders.value.filter(f => f.id !== id)
    if (activeSmartFolderId.value === id) {
      activeSmartFolderId.value = null
    }
    saveToSettings()
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
    loadFromSettings,
    create,
    rename,
    remove,
    setActive,
    clearActive,
  }
})
