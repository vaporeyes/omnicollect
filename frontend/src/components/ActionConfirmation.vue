<!-- ABOUTME: Fixed-scope confirmation for collection-wide mutations. -->
<!-- ABOUTME: Separates acknowledged writes from refreshes and blocks uncertain repeats. -->
<script setup lang="ts">
import {ref, onBeforeUnmount} from 'vue'
import ModalSurface from './ModalSurface.vue'
import {useEditorGuard} from '../composables/editorGuard'
const props = defineProps<{action: {title: string; description: string; label: string; run: () => Promise<string>}}>()
const emit = defineEmits<{close: []; confirmed: [message: string]; uncertain: []}>()
const action = {...props.action}
const busy = ref(false)
const error = ref('')
let alive = true
let completed = false
onBeforeUnmount(() => {alive = false})
useEditorGuard(ref(false), busy)
function close() {if (!busy.value) emit('close')}
async function confirm() {
  if (busy.value || error.value || completed) return
  busy.value = true
  try {
    const message = await action.run()
    completed = true
    if (alive) emit('confirmed', message)
  } catch (e: any) {
    if (alive) {
      error.value = `Change was not confirmed: ${e?.message || 'Connection lost'}. It may have completed. Reload and inspect the result before another change.`
      emit('uncertain')
    }
  } finally {busy.value = false}
}
function reload() {window.location.reload()}
</script>
<template>
  <ModalSurface :label="action.title" :busy="busy" @close="close">
    <div class="action-overlay" @click.self="close">
      <section class="action-dialog">
        <h2>{{ action.title }}</h2>
        <p>{{ action.description }}</p>
        <p v-if="error" role="alert">{{ error }}</p>
        <div class="actions">
          <button autofocus :disabled="busy" @click="close">{{ error ? 'Close' : 'Cancel' }}</button>
          <button v-if="error" @click="reload">Reload</button>
          <button class="confirm-action" :disabled="busy || !!error || completed" @click="confirm">{{ busy ? 'Saving…' : action.label }}</button>
        </div>
      </section>
    </div>
  </ModalSurface>
</template>
<style scoped>
.action-overlay { min-height: 100%; display: grid; place-items: center; background: rgba(0,0,0,.5); padding: 24px; box-sizing: border-box; }
.action-dialog { width: min(100%, 480px); background: var(--bg-primary); color: var(--text-primary); border: 1px solid var(--border-primary); border-radius: var(--radius-md); padding: 24px; overflow-wrap: anywhere; }
h2 { font-family: var(--font-heading); font-size: 20px; margin-top: 0; }
p { font-size: 14px; line-height: 1.5; }
[role=alert] { color: var(--error-text); }
.actions { display: flex; gap: 12px; justify-content: flex-end; flex-wrap: wrap; }
button { padding: 8px 14px; border: 1px solid var(--border-primary); border-radius: var(--radius-sm); color: var(--text-primary); background: var(--bg-secondary); cursor: pointer; }
button:disabled { opacity: .6; cursor: not-allowed; }
.confirm-action { background: var(--accent-blue); color: var(--text-on-accent); }
</style>
