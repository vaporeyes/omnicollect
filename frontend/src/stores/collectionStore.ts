import {defineStore} from 'pinia'
import {ref} from 'vue'
import {GetItems, SaveItem, DeleteItem} from '../../wailsjs/go/main/App'
import {main} from '../../wailsjs/go/models'

export const useCollectionStore = defineStore('collection', () => {
  const items = ref<main.Item[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const activeModuleId = ref('')
  const searchQuery = ref('')

  async function fetchItems() {
    loading.value = true
    error.value = null
    try {
      items.value = await GetItems(searchQuery.value, activeModuleId.value)
    } catch (e: any) {
      error.value = e?.message ?? String(e)
    } finally {
      loading.value = false
    }
  }

  async function saveItem(item: main.Item) {
    error.value = null
    try {
      const saved = await SaveItem(item)
      await fetchItems()
      return saved
    } catch (e: any) {
      error.value = e?.message ?? String(e)
      throw e
    }
  }

  async function deleteItem(id: string) {
    error.value = null
    try {
      await DeleteItem(id)
      await fetchItems()
    } catch (e: any) {
      error.value = e?.message ?? String(e)
      throw e
    }
  }

  async function searchAllItems(query: string): Promise<main.Item[]> {
    if (!query) return []
    return await GetItems(query, '')
  }

  function setFilter(moduleId: string) {
    activeModuleId.value = moduleId
    fetchItems()
  }

  function setSearch(query: string) {
    searchQuery.value = query
    fetchItems()
  }

  return {items, loading, error, activeModuleId, searchQuery, fetchItems, saveItem, deleteItem, searchAllItems, setFilter, setSearch}
})
