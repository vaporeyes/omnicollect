<!-- ABOUTME: Lightweight right-click context menu positioned at cursor coordinates. -->
<!-- ABOUTME: Accepts menu options with labels, actions, and optional destructive styling. -->
<script lang="ts" setup>
import {ref, watch, nextTick, onMounted, onBeforeUnmount} from 'vue'

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

const adjustedX = ref(0)
const adjustedY = ref(0)
let opener: HTMLElement | null = null
let generation = 0
let closing = false
function restoreFocus() {if (opener?.isConnected) opener.focus()}
function dismiss(restore = true) {
  if (!props.visible || closing) return
  closing = true
  if (restore) restoreFocus()
  emit('close')
}
function choose(action: string) {
  if (closing) return
  closing = true
  restoreFocus()
  emit('select', action)
  emit('close')
}
watch(() => [props.visible, props.x, props.y], async () => {
  const current = ++generation
  if (!props.visible) {
    if (menuEl.value?.contains(document.activeElement)) restoreFocus()
    return
  }
  closing = false
  if (!menuEl.value?.contains(document.activeElement)) opener = document.activeElement instanceof HTMLElement ? document.activeElement : null
  adjustedX.value = props.x
  adjustedY.value = props.y
  await nextTick()
  if (current !== generation || !props.visible || !menuEl.value) return
  const rect = menuEl.value.getBoundingClientRect()
  adjustedX.value = Math.max(8, Math.min(props.x, window.innerWidth - rect.width - 8))
  adjustedY.value = Math.max(8, Math.min(props.y, window.innerHeight - rect.height - 8))
  menuEl.value.querySelector<HTMLButtonElement>('button')?.focus()
}, {immediate: true})

function onKeydown(event: KeyboardEvent) {
  event.stopPropagation()
  if (event.key === 'Escape') {event.preventDefault(); dismiss(); return}
  if (event.key === 'Tab') {dismiss(); return}
  const buttons = Array.from(menuEl.value?.querySelectorAll<HTMLButtonElement>('button') || [])
  if (!buttons.length || !['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key)) return
  event.preventDefault()
  const current = buttons.indexOf(document.activeElement as HTMLButtonElement)
  const next = event.key === 'Home' ? 0 : event.key === 'End' ? buttons.length - 1 :
    (current + (event.key === 'ArrowDown' ? 1 : -1) + buttons.length) % buttons.length
  buttons[next]?.focus()
}
function onOutside(event: Event) {
  if (menuEl.value && !menuEl.value.contains(event.target as Node)) dismiss(false)
}
onMounted(() => {
  document.addEventListener('mousedown', onOutside, true)
  document.addEventListener('focusin', onOutside)
})
onBeforeUnmount(() => {
  generation++
  if (menuEl.value?.contains(document.activeElement)) restoreFocus()
  document.removeEventListener('mousedown', onOutside, true)
  document.removeEventListener('focusin', onOutside)
})
</script>

<template>
  <Teleport to="body">
    <Transition name="ctx">
      <div
        v-if="visible"
        ref="menuEl"
        class="context-menu"
        role="menu"
        aria-label="Actions"
        @keydown="onKeydown"
        :style="{left: adjustedX + 'px', top: adjustedY + 'px'}"
      >
        <button
          v-for="opt in options"
          :key="opt.action"
          :class="['ctx-item', {destructive: opt.destructive}]"
          role="menuitem"
          tabindex="-1"
          @click="choose(opt.action)"
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
  min-width: min(160px, calc(100vw - 16px));
  max-width: calc(100vw - 16px);
  max-height: calc(100vh - 16px);
  overflow: auto;
  padding: 4px;
  background: var(--bg-primary, #1e1e2e);
  border: 1px solid var(--border-primary, #333);
  border-radius: var(--radius-md, 8px);
  box-shadow: var(--shadow-lg, 0 8px 32px rgba(0,0,0,0.25));
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
.ctx-item:hover, .ctx-item:focus-visible {
  background: var(--bg-hover, rgba(255,255,255,0.06));
}
.ctx-item.destructive {
  color: var(--error-text, #ef4444);
}
.ctx-item.destructive:hover {
  background: var(--error-bg, rgba(239, 68, 68, 0.12));
}

/* Mechanical snap entrance */
.ctx-enter-active {
  transition: opacity 0.1s, transform 0.1s cubic-bezier(0, 0.55, 0.45, 1);
}
.ctx-leave-active {
  transition: opacity 0.08s;
}
.ctx-enter-from {
  opacity: 0;
  transform: translateY(-8px);
}
.ctx-leave-to {
  opacity: 0;
}
</style>
