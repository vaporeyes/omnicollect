// ABOUTME: Account-session epochs invalidate tokens, requests and deferred UI work.
// ABOUTME: Local mode starts active; authenticated mode is activated only by the outer guard.
import {ref} from 'vue'
export const sessionVersion = ref(0)
export const sessionReady = ref(true)
export const logoutRequested = ref(false)
export const sessionFailure = ref('')
export function beginSignOut() {logoutRequested.value = true; setSessionIdentity(null)}
let identity: string | null = 'local'
let controller = new AbortController()
export function setSessionIdentity(next: string | null) {
  if (next === identity) return
  identity = next
  sessionFailure.value = ''
  sessionReady.value = false
  controller.abort()
  controller = new AbortController()
  sessionVersion.value++
  sessionReady.value = next !== null
}
export function captureSession() {
  const version = sessionVersion.value
  const signal = controller.signal
  function assertCurrent() {
    if (!sessionReady.value || version !== sessionVersion.value) throw new DOMException('Account session changed', 'AbortError')
  }
  return {signal, assertCurrent}
}
