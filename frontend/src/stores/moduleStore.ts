import {defineStore} from 'pinia'
import {ref, computed} from 'vue'
import {GetActiveModules} from '../../wailsjs/go/main/App'
import {main} from '../../wailsjs/go/models'

export const useModuleStore = defineStore('modules', () => {
  const modules = ref<main.ModuleSchema[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const getModuleById = computed(() => {
    return (id: string) => modules.value.find(m => m.id === id) ?? null
  })

  async function fetchModules() {
    loading.value = true
    error.value = null
    try {
      modules.value = await GetActiveModules()
    } catch (e: any) {
      error.value = e?.message ?? String(e)
    } finally {
      loading.value = false
    }
  }

  return {modules, loading, error, getModuleById, fetchModules}
})
