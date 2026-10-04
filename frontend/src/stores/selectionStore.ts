// ABOUTME: Pinia store for multi-select state shared between list and grid views.
// ABOUTME: Manages selected item IDs, Shift-click range selection, and select-all.
import {defineStore} from 'pinia'
import {ref, computed} from 'vue'
import type {Item} from '../api/types'

export const useSelectionStore = defineStore('selection', () => {
  const selectedIds = ref<Set<string>>(new Set())
  const lastClickedIndex = ref<number | null>(null)
  const anchorId = ref<string | null>(null)

  const count = computed(() => selectedIds.value.size)
  const hasSelection = computed(() => selectedIds.value.size > 0)

  function isSelected(id: string): boolean {
    return selectedIds.value.has(id)
  }

  function toggle(id: string, index: number) {
    if (!id || !Number.isInteger(index) || index < 0) return
    anchorId.value = id
    const next = new Set(selectedIds.value)
    if (next.has(id)) {
      next.delete(id)
    } else {
      next.add(id)
    }
    selectedIds.value = next
    lastClickedIndex.value = index
  }

  function shiftSelect(targetIndex: number, items: Item[]) {
    if (!Number.isInteger(targetIndex) || !items[targetIndex]) return
    const anchor = items.findIndex(item => item.id === anchorId.value)
    if (anchor < 0) {
      toggle(items[targetIndex]?.id, targetIndex)
      return
    }
    const start = Math.min(anchor, targetIndex)
    const end = Math.max(anchor, targetIndex)
    const next = new Set(selectedIds.value)
    for (let i = start; i <= end; i++) {
      if (items[i]) next.add(items[i].id)
    }
    selectedIds.value = next
  }

  function selectAll(items: Item[]) {
    anchorId.value = null
    lastClickedIndex.value = null
    const next = new Set<string>()
    for (const item of items) {
      next.add(item.id)
    }
    selectedIds.value = next
  }

  function clear() {
    anchorId.value = null
    selectedIds.value = new Set()
    lastClickedIndex.value = null
  }

  function prune(items: Item[]) {
    const visible = new Set(items.map(item => item.id))
    selectedIds.value = new Set([...selectedIds.value].filter(id => visible.has(id)))
    if (anchorId.value && !visible.has(anchorId.value)) {
      anchorId.value = null
      lastClickedIndex.value = null
    }
  }

  function selectedIdArray(): string[] {
    return [...selectedIds.value]
  }

  return {selectedIds, lastClickedIndex, count, hasSelection, isSelected, toggle, shiftSelect, selectAll, clear, prune, selectedIdArray}
})
