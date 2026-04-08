// ABOUTME: Unit tests for the selection Pinia store.
// ABOUTME: Verifies toggle, shiftSelect, selectAll, clear, and isSelected (pure state, no fetch).
import {describe, it, expect, beforeEach} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {useSelectionStore} from './selectionStore'
import type {Item} from '../api/types'

function makeItem(id: string): Item {
  return {id, moduleId: 'test', title: id, purchasePrice: null, images: [], attributes: {}, createdAt: '', updatedAt: ''}
}

beforeEach(() => {
  setActivePinia(createPinia())
})

describe('toggle', () => {
  it('adds an ID when not selected', () => {
    const store = useSelectionStore()
    store.toggle('a', 0)
    expect(store.isSelected('a')).toBe(true)
    expect(store.count).toBe(1)
  })

  it('removes an ID when already selected', () => {
    const store = useSelectionStore()
    store.toggle('a', 0)
    store.toggle('a', 0)
    expect(store.isSelected('a')).toBe(false)
    expect(store.count).toBe(0)
  })
})

describe('shiftSelect', () => {
  it('selects range from last clicked to target', () => {
    const store = useSelectionStore()
    const items = [makeItem('a'), makeItem('b'), makeItem('c'), makeItem('d')]

    // First click sets anchor
    store.toggle('a', 0)
    // Shift-click on index 2
    store.shiftSelect(2, items)

    expect(store.isSelected('a')).toBe(true)
    expect(store.isSelected('b')).toBe(true)
    expect(store.isSelected('c')).toBe(true)
    expect(store.isSelected('d')).toBe(false)
  })

  it('falls back to toggle when no anchor exists', () => {
    const store = useSelectionStore()
    const items = [makeItem('a'), makeItem('b')]

    store.shiftSelect(1, items)
    expect(store.isSelected('b')).toBe(true)
    expect(store.count).toBe(1)
  })
})

describe('selectAll', () => {
  it('selects all item IDs', () => {
    const store = useSelectionStore()
    const items = [makeItem('a'), makeItem('b'), makeItem('c')]

    store.selectAll(items)

    expect(store.count).toBe(3)
    expect(store.isSelected('a')).toBe(true)
    expect(store.isSelected('b')).toBe(true)
    expect(store.isSelected('c')).toBe(true)
  })
})

describe('clear', () => {
  it('empties the selection', () => {
    const store = useSelectionStore()
    store.toggle('a', 0)
    store.toggle('b', 1)

    store.clear()

    expect(store.count).toBe(0)
    expect(store.hasSelection).toBe(false)
    expect(store.lastClickedIndex).toBeNull()
  })
})

describe('isSelected', () => {
  it('returns false for unselected IDs', () => {
    const store = useSelectionStore()
    expect(store.isSelected('x')).toBe(false)
  })
})

describe('selectedIdArray', () => {
  it('returns array of selected IDs', () => {
    const store = useSelectionStore()
    store.toggle('a', 0)
    store.toggle('b', 1)
    const arr = store.selectedIdArray()
    expect(arr).toHaveLength(2)
    expect(arr).toContain('a')
    expect(arr).toContain('b')
  })
})
