import {defineStore} from 'pinia'
import {ref, computed, onScopeDispose} from 'vue'
import * as api from '../api/client'
import type {ModuleSchema} from '../api/types'

export const useModuleStore = defineStore('modules', () => {
  const modules = ref<ModuleSchema[]>([])
  const loading = ref(false)
  const loaded = ref(false)
  const error = ref<string | null>(null)

  const getModuleById = computed(() => {
    return (id: string) => modules.value.find(m => m.id === id) ?? null
  })

  let disposed = false
  let version = 0
  let request: AbortController | undefined
  onScopeDispose(() => {disposed = true; loaded.value = false; version++; request?.abort(); modules.value = []; error.value = null; loading.value = false})
  async function fetchModules() {
    if (disposed) return
    const current = ++version
    request?.abort()
    request = new AbortController()
    loading.value = true
    error.value = null
    try {
      const result = await api.get<ModuleSchema[]>('/api/v1/modules', request.signal)
      if (current === version) {modules.value = result; loaded.value = true}
    } catch (e: any) {
      if (current === version) error.value = e?.message ?? String(e)
    } finally {
      if (current === version) loading.value = false
    }
  }

  return {modules, loading, loaded, error, getModuleById, fetchModules}
})
