<script lang="ts" setup>
import MediaImage from './MediaImage.vue'
import {ref, watch} from 'vue'
import ModalSurface from './ModalSurface.vue'

const props = defineProps<{
  filename: string
  visible: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const zoomed = ref(false)
const originX = ref('50%')
const originY = ref('50%')
const zoomLevel = ref(2.5)

function onMouseMove(event: MouseEvent) {
  if (!zoomed.value) return
  const img = event.currentTarget as HTMLElement
  const rect = img.getBoundingClientRect()
  const x = ((event.clientX - rect.left) / rect.width) * 100
  const y = ((event.clientY - rect.top) / rect.height) * 100
  originX.value = `${x}%`
  originY.value = `${y}%`
}

function toggleZoom() {
  zoomed.value = !zoomed.value
}

function onWheel(event: WheelEvent) {
  if (!zoomed.value) return
  event.preventDefault()
  const delta = event.deltaY > 0 ? -0.3 : 0.3
  zoomLevel.value = Math.min(8, Math.max(1.5, zoomLevel.value + delta))
}

function reset() {
  zoomed.value = false
  zoomLevel.value = 2.5
  originX.value = originY.value = '50%'
}
watch(() => [props.visible, props.filename], reset)
function changeZoom(delta: number) {
  zoomed.value = true
  zoomLevel.value = Math.min(8, Math.max(1.5, zoomLevel.value + delta))
}
function onKeydown(event: KeyboardEvent) {
  event.stopPropagation()
  if (event.key === '+' || event.key === '=') {event.preventDefault(); changeZoom(.3)}
  else if (event.key === '-') {event.preventDefault(); changeZoom(-.3)}
  else if (zoomed.value && ['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'].includes(event.key)) {
    event.preventDefault()
    const clamp = (value: number) => Math.max(0, Math.min(100, value)) + '%'
    if (event.key === 'ArrowLeft') originX.value = clamp(parseFloat(originX.value) - 10)
    if (event.key === 'ArrowRight') originX.value = clamp(parseFloat(originX.value) + 10)
    if (event.key === 'ArrowUp') originY.value = clamp(parseFloat(originY.value) - 10)
    if (event.key === 'ArrowDown') originY.value = clamp(parseFloat(originY.value) + 10)
  }
}
function onClose() {reset(); emit('close')}

</script>

<template>
  <ModalSurface v-if="visible" label="Image viewer" @close="onClose" @keydown="onKeydown">
    <div class="lightbox-overlay" @click.self="onClose">
      <div class="lightbox-content" @click.stop>
        <button autofocus class="lightbox-close" aria-label="Close image viewer" @click="onClose">×</button>
        <div class="zoom-controls">
          <button @click="changeZoom(-.3)" aria-label="Zoom out">−</button>
          <button @click="changeZoom(.3)" aria-label="Zoom in">+</button>
          <button @click="reset">Reset zoom</button>
          <span role="status">{{ zoomed ? zoomLevel.toFixed(1) : '1.0' }}×</span>
        </div>
        <div class="lightbox-hint" v-if="!zoomed">Click or press Enter on the image to inspect</div>
        <div class="lightbox-hint" v-else>Scroll or use +/− to zoom. Arrow keys pan. Click image to reset.</div>
        <div
          class="loupe-container"
          :class="{zoomed}"
          role="button"
          tabindex="0"
          aria-label="Toggle image zoom"
          :aria-pressed="zoomed"
          @keydown.enter.prevent="toggleZoom"
          @keydown.space.prevent="toggleZoom"
          @click="toggleZoom"
          @mousemove="onMouseMove"
          @wheel="onWheel"
        >
          <MediaImage
            :src="'/originals/' + encodeURIComponent(filename)"
            alt="Full resolution"
            :style="zoomed ? {
              transform: `scale(${zoomLevel})`,
              transformOrigin: `${originX} ${originY}`,
            } : {}"
          />
        </div>
      </div>
    </div>
  </ModalSurface>
</template>

<style scoped>
.loupe-container:focus-visible {outline: 2px solid white; outline-offset: 2px;}
.zoom-controls {display: flex; align-items: center; gap: 10px; margin-bottom: 8px; color: white;}
.zoom-controls button {padding: 6px 12px; background: var(--bg-primary); color: var(--text-primary); border: 1px solid var(--border-primary); border-radius: var(--radius-sm);}
.lightbox-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.92);
  background-image: radial-gradient(
    circle, rgba(255, 255, 255, 0.04) 1px, transparent 1px
  );
  background-size: 6px 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.lightbox-content {
  position: relative;
  max-width: 90vw;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.loupe-container {
  overflow: hidden;
  border-radius: 4px;
  cursor: zoom-in;
  max-width: 90vw;
  max-height: 85vh;
}
.loupe-container.zoomed {
  cursor: crosshair;
}
.loupe-container img {
  display: block;
  max-width: 90vw;
  max-height: 85vh;
  object-fit: contain;
  transition: transform 0.1s ease-out;
  will-change: transform;
}
.lightbox-hint {
  color: rgba(255,255,255,0.6);
  font-size: 12px;
  margin-bottom: 6px;
  text-align: center;
  pointer-events: none;
}
.lightbox-close {
  position: absolute;
  top: -12px;
  right: -12px;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: none;
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-md);
  z-index: 1;
}

</style>
