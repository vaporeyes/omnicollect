// ABOUTME: Unit tests for the Smart Folder store.
// ABOUTME: Covers CRUD operations, persistence loading, and validation.

import {describe, it, expect, beforeEach, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {useSmartFolderStore} from '../smartFolderStore'

// Mock the API client
vi.mock('../../api/client', () => ({
  get: vi.fn().mockResolvedValue({}),
  put: vi.fn().mockResolvedValue(undefined),
}))

describe('smartFolderStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts with empty folders', () => {
    const store = useSmartFolderStore()
    expect(store.folders).toEqual([])
    expect(store.activeSmartFolderId).toBeNull()
  })

  it('creates a folder with state', () => {
    const store = useSmartFolderStore()
    const folder = store.create('Test Folder', 'coins', 'liberty', {}, ['silver'])
    expect(folder).not.toBeNull()
    expect(folder!.name).toBe('Test Folder')
    expect(folder!.moduleId).toBe('coins')
    expect(folder!.searchQuery).toBe('liberty')
    expect(folder!.tags).toEqual(['silver'])
    expect(folder!.id).toHaveLength(8)
    expect(store.folders).toHaveLength(1)
  })

  it('rejects empty name on create', () => {
    const store = useSmartFolderStore()
    expect(store.create('', 'coins', '', {}, [])).toBeNull()
    expect(store.create('   ', 'coins', '', {}, [])).toBeNull()
    expect(store.folders).toHaveLength(0)
  })

  it('trims whitespace from name on create', () => {
    const store = useSmartFolderStore()
    const folder = store.create('  My Folder  ', 'coins', '', {}, [])
    expect(folder!.name).toBe('My Folder')
  })

  it('generates unique IDs', () => {
    const store = useSmartFolderStore()
    store.create('Folder 1', '', '', {}, [])
    store.create('Folder 2', '', '', {}, [])
    const ids = store.folders.map(f => f.id)
    expect(ids[0]).not.toBe(ids[1])
  })

  it('renames a folder', () => {
    const store = useSmartFolderStore()
    const folder = store.create('Original', '', '', {}, [])!
    expect(store.rename(folder.id, 'Updated')).toBe(true)
    expect(store.folders[0].name).toBe('Updated')
  })

  it('rejects empty name on rename', () => {
    const store = useSmartFolderStore()
    const folder = store.create('Original', '', '', {}, [])!
    expect(store.rename(folder.id, '')).toBe(false)
    expect(store.rename(folder.id, '  ')).toBe(false)
    expect(store.folders[0].name).toBe('Original')
  })

  it('returns false when renaming non-existent folder', () => {
    const store = useSmartFolderStore()
    expect(store.rename('nonexistent', 'Name')).toBe(false)
  })

  it('deletes a folder', () => {
    const store = useSmartFolderStore()
    const folder = store.create('To Delete', '', '', {}, [])!
    store.remove(folder.id)
    expect(store.folders).toHaveLength(0)
  })

  it('clears activeSmartFolderId when active folder is deleted', () => {
    const store = useSmartFolderStore()
    const folder = store.create('Active', '', '', {}, [])!
    store.setActive(folder.id)
    expect(store.activeSmartFolderId).toBe(folder.id)
    store.remove(folder.id)
    expect(store.activeSmartFolderId).toBeNull()
  })

  it('does not clear activeSmartFolderId when a different folder is deleted', () => {
    const store = useSmartFolderStore()
    const f1 = store.create('Keep', '', '', {}, [])!
    const f2 = store.create('Delete', '', '', {}, [])!
    store.setActive(f1.id)
    store.remove(f2.id)
    expect(store.activeSmartFolderId).toBe(f1.id)
  })

  it('loads folders from settings object', () => {
    const store = useSmartFolderStore()
    store.loadFromSettings({
      theme: {},
      smartFolders: [
        {id: 'abc', name: 'Loaded', moduleId: 'coins', searchQuery: '', filters: {}, tags: [], createdAt: '2026-01-01T00:00:00Z'},
      ],
    })
    expect(store.folders).toHaveLength(1)
    expect(store.folders[0].name).toBe('Loaded')
  })

  it('handles missing smartFolders key in settings', () => {
    const store = useSmartFolderStore()
    store.loadFromSettings({theme: {}})
    expect(store.folders).toEqual([])
  })

  it('handles null/undefined settings', () => {
    const store = useSmartFolderStore()
    store.loadFromSettings(null)
    expect(store.folders).toEqual([])
    store.loadFromSettings(undefined)
    expect(store.folders).toEqual([])
  })

  it('deep copies filters on create to avoid mutation', () => {
    const store = useSmartFolderStore()
    const filters = {condition: [{field: 'condition', op: 'in' as const, values: ['Mint']}]}
    const folder = store.create('Test', 'coins', '', filters, [])!
    // Mutate original -- should not affect stored folder
    filters.condition[0].values!.push('Fine')
    expect(folder.filters.condition[0].values).toEqual(['Mint'])
  })
})
