<!-- ABOUTME: Authenticated image loading without putting access tokens in URLs. -->
<!-- ABOUTME: Lazily fetches thumbnails, cancels stale work, and releases blob URLs. -->
<script setup lang="ts">
import {onMounted, onBeforeUnmount, ref, watch} from 'vue'
import {getMedia} from '../api/client'

const props = defineProps<{src: string; loading?: 'lazy' | 'eager'}>()
const emit = defineEmits<{load: [event: Event]; error: [event: Event]}>()
const element = ref<HTMLImageElement | null>(null)
const objectURL = ref<string>()
const pending = ref(false)
const visible = ref(props.loading !== 'lazy' || typeof IntersectionObserver === 'undefined')
let observer: IntersectionObserver | undefined

onMounted(() => {
  if (!visible.value && element.value) {
    observer = new IntersectionObserver(entries => {
      if (entries.some(entry => entry.isIntersecting)) {
        visible.value = true
        observer?.disconnect()
      }
    }, {rootMargin: '200px'})
    observer.observe(element.value)
  }
})
onBeforeUnmount(() => observer?.disconnect())

watch([() => props.src, visible], ([src, shown], _, cleanup) => {
  objectURL.value = undefined
  pending.value = false
  if (!shown || !src) return
  const controller = new AbortController()
  let alive = true
  let ownedURL: string | undefined
  pending.value = true
  cleanup(() => {
    alive = false
    controller.abort()
    if (ownedURL) URL.revokeObjectURL(ownedURL)
  })
  void getMedia(src, controller.signal).then(blob => {
    if (!alive) return
    ownedURL = URL.createObjectURL(blob)
    objectURL.value = ownedURL
  }).catch(() => {
    if (alive && !controller.signal.aborted) element.value?.dispatchEvent(new Event('error'))
  }).finally(() => {
    if (alive) pending.value = false
  })
}, {immediate: true})
</script>

<template>
  <img ref="element" :src="objectURL" :aria-busy="pending" :loading="loading"
       @load="emit('load', $event)" @error="emit('error', $event)" />
</template>
