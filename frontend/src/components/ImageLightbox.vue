<script lang="ts" setup>
import {ref} from 'vue'

defineProps<{
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

function onClose() {
  zoomed.value = false
  zoomLevel.value = 2.5
  emit('close')
}
</script>

<template>
  <div v-if="visible" class="lightbox-overlay" @click="onClose">
    <div class="lightbox-content" @click.stop>
      <button class="lightbox-close" @click="onClose">x</button>
      <div class="lightbox-hint" v-if="!zoomed">Click image to inspect</div>
      <div class="lightbox-hint" v-else>Scroll to adjust zoom. Click to exit.</div>
      <div
        class="loupe-container"
        :class="{zoomed}"
        @click="toggleZoom"
        @mousemove="onMouseMove"
        @mouseleave="zoomed = false"
        @wheel="onWheel"
      >
        <img
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
</template>

<style scoped>
.lightbox-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--overlay-bg);
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
