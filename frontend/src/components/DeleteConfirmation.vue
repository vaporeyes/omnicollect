<!-- ABOUTME: Explicit, fixed-target confirmation for recoverable item deletion. -->
<!-- ABOUTME: Prevents duplicate writes and dismissal while busy; failed requests remain visible. -->
<script setup lang="ts">
import {ref, onBeforeUnmount} from 'vue'
import ModalSurface from './ModalSurface.vue'
import {useEditorGuard} from '../composables/editorGuard'
import {APIError} from '../api/client'
const props = defineProps<{
  items: {id: string; title: string}[]
  remove: (ids: string[]) => Promise<number>
}>()
const emit = defineEmits<{close: []; deleted: [count: number, ids: string[]]}>()
const targets = props.items.map(item => ({id: item.id, title: item.title}))
const busy = ref(false)
const error = ref('')
const completed = ref(false)
useEditorGuard(ref(false), busy)
let alive = true
onBeforeUnmount(() => {alive = false})
function close() {if (!busy.value) emit('close')}
async function confirm() {
  if (busy.value || completed.value || error.value || !targets.length || targets.length > 500) return
  busy.value = true
  error.value = ''
  try {
    const ids = targets.map(item => item.id)
    const count = await props.remove(ids)
    completed.value = true
    if (alive) emit('deleted', count, ids)
  } catch (e: any) {
    if (alive) error.value = e instanceof APIError
      ? `Deletion was not confirmed: ${e.message}. Inspect Deletion recovery and refresh the collection before retrying.`
      : 'Connection lost. Deletion may have completed. Close and inspect Deletion recovery before retrying.'
  } finally {busy.value = false}
}
</script>
<template>
  <ModalSurface label="Confirm item deletion" :busy="busy" @close="close">
    <div class="delete-overlay" @click.self="close">
      <section class="delete-dialog">
        <h2>Delete {{ targets.length }} item{{ targets.length === 1 ? '' : 's' }}?</h2>
        <ul><li v-for="item in targets.slice(0, 5)" :key="item.id">{{ item.title }}</li></ul>
        <p v-if="targets.length > 5">And {{ targets.length - 5 }} more selected items.</p>
        <p>These records leave the collection but remain in Deletion recovery in the sidebar, where you can recover or permanently discard the batch. Media files are retained. ZIP backups exclude deleted records; Replace imports permanently clear recovery.</p>
        <p v-if="error" role="alert">{{ error }}</p>
        <div class="delete-actions">
          <button autofocus :disabled="busy" @click="close">Cancel</button>
          <button class="danger" :disabled="busy || completed || !!error || !targets.length || targets.length > 500" @click="confirm">
            {{ busy ? 'Deleting…' : `Delete ${targets.length} item${targets.length === 1 ? '' : 's'}` }}
          </button>
        </div>
      </section>
    </div>
  </ModalSurface>
</template>
<style scoped>
.delete-overlay { min-height: 100%; display: grid; place-items: center; background: rgba(0,0,0,.5); padding: 24px; box-sizing: border-box; }
.delete-dialog { width: min(100%, 480px); background: var(--bg-primary); color: var(--text-primary); border: 1px solid var(--border-primary); border-radius: var(--radius-md); padding: 24px; overflow-wrap: anywhere; }
h2 { font-size: 20px; margin-top: 0; }
p, li { font-size: 14px; line-height: 1.5; }
[role=alert] { color: var(--error-text); }
.delete-actions { display: flex; gap: 12px; justify-content: flex-end; }
button { padding: 8px 14px; border: 1px solid var(--border-primary); border-radius: var(--radius-sm); color: var(--text-primary); background: var(--bg-secondary); cursor: pointer; }
button:disabled { opacity: .6; cursor: wait; }
.danger { background: var(--error-bg); color: var(--error-text); }
</style>
