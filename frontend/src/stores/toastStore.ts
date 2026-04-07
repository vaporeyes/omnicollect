// ABOUTME: Pinia store for global toast notifications.
// ABOUTME: Manages a queue of toast messages with auto-dismiss timers.
import {defineStore} from 'pinia'
import {ref} from 'vue'

export interface Toast {
  id: number
  message: string
  type: 'success' | 'error' | 'info'
}

let nextId = 0

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<Toast[]>([])

  function show(message: string, type: Toast['type'] = 'info', durationMs = 4000) {
    const id = nextId++
    toasts.value.push({id, message, type})
    setTimeout(() => dismiss(id), durationMs)
  }

  function dismiss(id: number) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  return {toasts, show, dismiss}
})
