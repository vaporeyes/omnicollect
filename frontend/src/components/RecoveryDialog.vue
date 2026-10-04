<!-- ABOUTME: Durable deletion recovery with explicit fixed-batch restore/discard consent. -->
<!-- ABOUTME: Blocks duplicate writes, guards unload, and separates commit acknowledgement from refresh. -->
<script setup lang="ts">
import {ref, onMounted, onBeforeUnmount, nextTick} from 'vue'
import * as api from '../api/client'
import ModalSurface from './ModalSurface.vue'
import {useEditorGuard} from '../composables/editorGuard'
import type {DeletionBatch as Batch} from '../api/types'
const props = defineProps<{refresh: () => Promise<void>}>()
const emit = defineEmits<{close: []}>()
const batches = ref<Batch[]>([])
const busy = ref(false), loading = ref(false), loaded = ref(false)
const error = ref(''), message = ref(''), uncertain = ref(false)
const selected = ref<{batch: Batch; discard: boolean} | null>(null)
const consent = ref(false)
const cancelButton = ref<HTMLButtonElement | null>(null)
const refreshButton = ref<HTMLButtonElement | null>(null)
const controller = new AbortController()
let alive = true
useEditorGuard(ref(false), busy)
onBeforeUnmount(() => {alive = false; controller.abort()})
onMounted(() => load(false))
function close() {if (!busy.value) emit('close')}
async function load(refreshCollection: boolean) {
  if (loading.value) return
  loading.value = true
  error.value = ''
  try {
    const result = await api.get<Batch[]>('/api/v1/recovery', controller.signal)
    if (!alive) return
    batches.value = result
    loaded.value = true
    if (refreshCollection) await props.refresh()
    if (alive) uncertain.value = false
  } catch (e: any) {
    if (alive) error.value = `Refresh failed: ${e?.message || 'Connection lost'}. Recovery actions require a successful refresh.`
  } finally {loading.value = false}
}
async function choose(batch: Batch, discard: boolean) {
  if (busy.value || loading.value || error.value || uncertain.value || selected.value) return
  selected.value = {batch: {...batch, titles: [...batch.titles]}, discard}
  consent.value = false
  await nextTick()
  cancelButton.value?.focus()
}
async function confirm() {
  if (busy.value || !selected.value || !consent.value || uncertain.value) return
  const {batch, discard} = selected.value
  busy.value = true
  message.value = ''
  try {
    if (discard) await api.del('/api/v1/recovery/' + encodeURIComponent(batch.recoveryId))
    else await api.post('/api/v1/recovery/' + encodeURIComponent(batch.recoveryId), {})
    if (!alive) return
    message.value = discard ? 'Recovery records permanently discarded. Media files were not erased.' : `Recovered ${batch.deleted} item(s).`
    batches.value = batches.value.filter(b => b.recoveryId !== batch.recoveryId)
    selected.value = null
  } catch (e: any) {
    if (alive) {
      uncertain.value = true
      error.value = `Change was not confirmed: ${e?.message || 'Connection lost'}. It may have completed. Refresh and inspect the collection and recovery list before another change.`
    }
    return
  } finally {busy.value = false}
  // Only refresh after an acknowledged write; failed refresh never resends the mutation.
  await load(true)
  await nextTick()
  if (alive) refreshButton.value?.focus()
}
async function reload() {
  if (busy.value || loading.value) return
  selected.value = null
  await load(true)
}
</script>
<template>
  <ModalSurface label="Deletion recovery" :busy="busy" @close="close">
    <div class="recovery-overlay" @click.self="close">
      <section class="recovery-dialog">
        <h2>Deletion recovery</h2>
        <p>Deleted records stay here until recovered or explicitly discarded. Up to 100 batches / 64 MB are retained; further deletions stop at capacity. A batch can contain at most 500 items / 8 MB.</p>
        <p>ZIP backups exclude these records. Replace imports permanently clear this list. Image files are retained, but recovery cannot recreate files removed outside the app.</p>
        <p v-if="message" role="status">{{ message }}</p>
        <p v-if="error" role="alert">{{ error }}</p>
        <p v-if="loading" role="status">Refreshing…</p>
        <div v-if="selected" class="confirmation" role="group" aria-labelledby="recovery-confirm">
          <h3 id="recovery-confirm">{{ selected.discard ? 'Permanently discard' : 'Recover' }} {{ selected.batch.deleted }} item(s)?</h3>
          <p>{{ selected.batch.titles.join(', ') }}{{ selected.batch.deleted > 5 ? ', and more' : '' }}</p>
          <label><input v-model="consent" type="checkbox" :disabled="busy || uncertain" />
            {{ selected.discard ? 'I understand these saved records cannot be recovered after discard. This does not erase media files or copies in existing backups.' : 'I understand recovered items may reappear in enabled public showcases. Existing IDs are never overwritten; incompatible schemas block the whole batch.' }}
          </label>
          <div class="actions">
            <button ref="cancelButton" :disabled="busy" @click="selected = null">Cancel</button>
            <button class="confirm-action" :disabled="busy || !consent || uncertain" @click="confirm">{{ busy ? 'Working…' : selected.discard ? 'Permanently discard batch' : 'Recover batch' }}</button>
          </div>
        </div>
        <ul v-else-if="loaded && !loading && !error">
          <li v-for="batch in batches" :key="batch.recoveryId">
            <p><strong>{{ batch.deleted }} item(s)</strong> — {{ batch.createdAt }}</p>
            <p>{{ batch.titles.join(', ') }}{{ batch.deleted > 5 ? ', and more' : '' }}</p>
            <div class="actions">
              <button :disabled="uncertain" @click="choose(batch, false)">Recover</button>
              <button :disabled="uncertain" @click="choose(batch, true)">Discard permanently…</button>
            </div>
          </li>
        </ul>
        <p v-if="loaded && !loading && !error && !batches.length">No deleted batches available.</p>
        <div class="actions">
          <button ref="refreshButton" :disabled="busy || loading" @click="reload">Refresh collection and recovery list</button>
          <button :disabled="busy" @click="close">Done</button>
        </div>
      </section>
    </div>
  </ModalSurface>
</template>
<style scoped>
.recovery-overlay { min-height: 100%; display: grid; place-items: center; background: rgba(0,0,0,.5); padding: 24px; box-sizing: border-box; }
.recovery-dialog { width: min(100%, 640px); background: var(--bg-primary); color: var(--text-primary); border: 1px solid var(--border-primary); border-radius: var(--radius-md); padding: 24px; overflow-wrap: anywhere; }
h2 { font-family: var(--font-heading); font-size: 20px; margin-top: 0; }
p, label { font-size: 14px; line-height: 1.5; }
ul { list-style: none; padding: 0; }
li, .confirmation { padding: 12px 0; border-top: 1px solid var(--border-primary); }
[role=alert] { color: var(--error-text); }
.actions { display: flex; gap: 12px; justify-content: flex-end; flex-wrap: wrap; margin-top: 12px; }
button { padding: 8px 14px; border: 1px solid var(--border-primary); border-radius: var(--radius-sm); color: var(--text-primary); background: var(--bg-secondary); cursor: pointer; }
button:disabled { opacity: .6; cursor: not-allowed; }
.confirm-action { background: var(--accent-blue); color: var(--text-on-accent); }
</style>
