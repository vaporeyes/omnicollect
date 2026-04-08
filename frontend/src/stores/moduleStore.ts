import {defineStore} from 'pinia'
import {ref, computed} from 'vue'
import * as api from '../api/client'
import type {ModuleSchema} from '../api/types'

export const useModuleStore = defineStore('modules', () => {
  const modules = ref<ModuleSchema[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const getModuleById = computed(() => {
    return (id: string) => modules.value.find(m => m.id === id) ?? null
  })

  async function fetchModules() {
    loading.value = true
    error.value = null
    try {
      modules.value = await api.get<ModuleSchema[]>('/api/v1/modules')
    } catch (e: any) {
      error.value = e?.message ?? String(e)
    } finally {
      loading.value = false
    }
  }

  return {modules, loading, error, getModuleById, fetchModules}
})
