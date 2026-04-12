<!-- ABOUTME: Lightweight right-click context menu positioned at cursor coordinates. -->
<!-- ABOUTME: Accepts menu options with labels, actions, and optional destructive styling. -->
<script lang="ts" setup>
import {ref, watch, onMounted, onUnmounted} from 'vue'

export interface MenuOption {
  label: string
  action: string
  destructive?: boolean
}

const props = defineProps<{
  visible: boolean
  x: number
  y: number
  options: MenuOption[]
}>()

const emit = defineEmits<{
  select: [action: string]
  close: []
}>()

const menuEl = ref<HTMLElement | null>(null)

// Adjusted position to keep menu within viewport
const adjustedX = ref(0)
const adjustedY = ref(0)

watch(() => [props.visible, props.x, props.y], () => {
  if (!props.visible) return
  // Start at cursor, will adjust after render in next tick
  adjustedX.value = props.x
  adjustedY.value = props.y
  requestAnimationFrame(() => {
    if (!menuEl.value) return
    const rect = menuEl.value.getBoundingClientRect()
    const vw = window.innerWidth
    const vh = window.innerHeight
    if (rect.right > vw) adjustedX.value = vw - rect.width - 8
    if (rect.bottom > vh) adjustedY.value = vh - rect.height - 8
  })
})

function onClickOutside(e: MouseEvent) {
  if (menuEl.value && !menuEl.value.contains(e.target as Node)) {
    emit('close')
  }
}

onMounted(() => document.addEventListener('mousedown', onClickOutside, true))
onUnmounted(() => document.removeEventListener('mousedown', onClickOutside, true))
</script>

<template>
  <Teleport to="body">
    <Transition name="ctx">
      <div
        v-if="visible"
        ref="menuEl"
        class="context-menu"
        :style="{left: adjustedX + 'px', top: adjustedY + 'px'}"
      >
        <button
          v-for="opt in options"
          :key="opt.action"
          :class="['ctx-item', {destructive: opt.destructive}]"
          @click="emit('select', opt.action); emit('close')"
        >
          {{ opt.label }}
        </button>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.context-menu {
  position: fixed;
  z-index: 2000;
  min-width: 160px;
  padding: 4px;
  background: var(--bg-primary, #1e1e2e);
  border: 1px solid var(--border-primary, #333);
  border-radius: var(--radius-md, 8px);
  box-shadow: var(--shadow-lg, 0 8px 32px rgba(0,0,0,0.25));
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}
.ctx-item {
  display: block;
  width: 100%;
  padding: 7px 12px;
  border: none;
  border-radius: var(--radius-sm, 4px);
  background: transparent;
  color: var(--text-primary);
  font-size: 13px;
  font-family: var(--font-body);
  text-align: left;
  cursor: pointer;
  transition: background 0.1s;
}
.ctx-item:hover {
  background: var(--bg-hover, rgba(255,255,255,0.06));
}
.ctx-item.destructive {
  color: var(--error-text, #ef4444);
}
.ctx-item.destructive:hover {
  background: var(--error-bg, rgba(239, 68, 68, 0.12));
}

/* Entrance/exit animation */
.ctx-enter-active {
  transition: opacity 0.12s ease-out, transform 0.12s ease-out;
}
.ctx-leave-active {
  transition: opacity 0.08s ease-in, transform 0.08s ease-in;
}
.ctx-enter-from {
  opacity: 0;
  transform: scale(0.95);
}
.ctx-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
</style>
