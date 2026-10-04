// ABOUTME: Pinia store for global toast notifications.
// ABOUTME: Manages a queue of toast messages with auto-dismiss timers.
import {defineStore} from 'pinia'
import {ref, onScopeDispose} from 'vue'

export interface Toast {
  id: number
  message: string
  type: 'success' | 'error' | 'info'
}

let nextId = 0

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<Toast[]>([])
  let disposed = false
  const timers = new Map<number, ReturnType<typeof setTimeout>>()
  onScopeDispose(() => {disposed = true; timers.forEach(clearTimeout); timers.clear(); toasts.value = []})

  function show(message: string, type: Toast['type'] = 'info', durationMs = 4000) {
    if (disposed) return
    const id = nextId++
    toasts.value.push({id, message, type})
    timers.set(id, setTimeout(() => dismiss(id), durationMs))
  }

  function dismiss(id: number) {
    clearTimeout(timers.get(id)); timers.delete(id)
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  return {toasts, show, dismiss}
})
