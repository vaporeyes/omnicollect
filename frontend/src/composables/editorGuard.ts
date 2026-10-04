// ABOUTME: Navigation and unload protection shared by schema and settings editors.
// ABOUTME: Pending writes block exit; dirty drafts require explicit discard consent.
import {onMounted, onBeforeUnmount, type Ref} from 'vue'
export function useEditorGuard(dirty: Ref<boolean>, saving: Ref<boolean>) {
  function canLeave() {
    if (saving.value) return false
    return !dirty.value || window.confirm('Discard unsaved changes?')
  }
  function beforeUnload(event: BeforeUnloadEvent) {
    if (!dirty.value && !saving.value) return
    event.preventDefault()
    event.returnValue = ''
  }
  onMounted(() => window.addEventListener('beforeunload', beforeUnload))
  onBeforeUnmount(() => window.removeEventListener('beforeunload', beforeUnload))
  return {canLeave}
}
