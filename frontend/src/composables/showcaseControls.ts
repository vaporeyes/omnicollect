// ABOUTME: Shared publication consent and request guards for collection sharing controls.
// ABOUTME: Discloses original-image metadata and reports toggle/clipboard failures visibly.
import {ref, onMounted, onBeforeUnmount} from 'vue'
import {toggleShowcase, listShowcases, getAIStatus} from '../api/client'
import type {ModuleSchema, Showcase} from '../api/types'
export function useShowcaseControls() {
  const showcaseMap = ref<Record<string, Showcase>>({})
  const copiedSlug = ref<string | null>(null)
  const isCloudMode = ref(false)
  const sharingReady = ref(false)
  const sharingError = ref('')
  const sharingBusy = ref<Record<string, boolean>>({})
  let alive = true
  let timer: ReturnType<typeof setTimeout> | undefined
  onBeforeUnmount(() => {alive = false; clearTimeout(timer)})
  onMounted(async () => {
    try {
      const status = await getAIStatus()
      if (!alive) return
      isCloudMode.value = status.cloudMode === true
      if (isCloudMode.value) {
        const entries = await listShowcases()
        if (!alive) return
        for (const sc of entries) showcaseMap.value[sc.moduleId] = sc
      }
      sharingReady.value = true
    } catch {if (alive) sharingError.value = 'Sharing status could not load. Reload before changing visibility.'}
  })
  async function onToggleShowcase(mod: ModuleSchema, event?: Event) {
    event?.stopPropagation()
    if (!sharingReady.value || sharingBusy.value[mod.id]) return
    const enabled = !showcaseMap.value[mod.id]?.enabled
    if (enabled && !window.confirm(`Publish “${mod.displayName}”? Anyone with the link can view the collection name, item titles and cover images, including full-resolution originals. Embedded metadata such as GPS coordinates is NOT removed. Future items and cover changes will also be public. Prices, tags, other attributes and additional images are not included in this gallery. Making it private blocks future requests, but cannot recall copies already downloaded. Continue?`)) return
    sharingBusy.value[mod.id] = true
    sharingError.value = ''
    try {
      const result = await toggleShowcase(mod.id, enabled)
      if (alive) showcaseMap.value[mod.id] = result
    } catch {if (alive) {sharingReady.value = false; sharingError.value = 'Visibility change was not confirmed. Reload to check its current state before retrying.'}}
    finally {sharingBusy.value[mod.id] = false}
  }
  async function copyShowcaseUrl(mod: ModuleSchema, event?: Event) {
    event?.stopPropagation()
    const sc = showcaseMap.value[mod.id]
    if (!sc?.url || !sharingReady.value || sharingBusy.value[mod.id]) return
    try {
      await navigator.clipboard.writeText(new URL(sc.url, window.location.origin).href)
      if (!alive) return
      copiedSlug.value = sc.slug
      sharingError.value = ''
      clearTimeout(timer)
      timer = setTimeout(() => {copiedSlug.value = null}, 2000)
    } catch {if (alive) sharingError.value = 'Could not copy the link. Clipboard access may be blocked.'}
  }
  return {showcaseMap, copiedSlug, isCloudMode, sharingReady, sharingError, sharingBusy, onToggleShowcase, copyShowcaseUrl}
}
