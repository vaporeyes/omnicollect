<!-- ABOUTME: Native modal surface with explicit Tab boundaries and browser background inertness. -->
<!-- ABOUTME: Prevents Escape dismissal while busy and restores focus when removed. -->
<script setup lang="ts">
import {ref, onMounted, onBeforeUnmount, nextTick, watch} from 'vue'
const props = defineProps<{label: string; busy?: boolean}>()
const emit = defineEmits<{close: []}>()
const element = ref<HTMLDialogElement | null>(null)
let previousFocus: HTMLElement | null = null
onMounted(() => {
  previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null
  element.value?.showModal()
  retainModalFocus()
})
// Slot mount/loading or busy updates can disable the focused control. Browsers
// then focus body; restore focus to the native dialog without moving valid focus.
async function retainModalFocus() {
  await nextTick()
  const dialog = element.value
  const active = document.activeElement
  // Chrome can retain a disabled activeElement until its later focus-fixup task.
  if (dialog?.open && (!dialog.contains(active) || active?.matches(':disabled'))) dialog.focus()
}
watch(() => props.busy, retainModalFocus, {flush: 'post'})
onBeforeUnmount(() => {
  element.value?.close()
  nextTick(() => {if (previousFocus?.isConnected) previousFocus.focus()})
})
function cycleTab(event: KeyboardEvent) {
  const dialog = element.value
  if (!dialog) return
  const controls = Array.from(dialog.querySelectorAll<HTMLElement>('button,input,select,textarea,a[href],[tabindex],[contenteditable="true"]'))
    .filter(control => control.tabIndex >= 0 && !control.matches(':disabled') && !control.closest('[inert]') && control.getClientRects().length > 0 && getComputedStyle(control).visibility !== 'hidden')
  const first = controls[0], last = controls[controls.length - 1], active = document.activeElement
  // Native Chrome can focus the document/toolbar after the final control even
  // while a modal is open. Keep sequential keyboard navigation inside the dialog.
  if (!first || active === dialog || (event.shiftKey ? active === first : active === last)) {
    event.preventDefault()
    ;(event.shiftKey ? last : first)?.focus()
    if (!first) dialog.focus()
  }
}
function requestClose() {if (!props.busy) emit('close')}
</script>
<template>
  <dialog ref="element" :aria-label="label" :aria-busy="busy || undefined"
          @cancel.prevent="requestClose" @keydown.esc.stop @keydown.tab="cycleTab">
    <slot />
  </dialog>
</template>
<style scoped>
dialog { max-width: none; max-height: none; width: 100%; height: 100%; margin: 0; padding: 0; border: 0; color: inherit; }
dialog::backdrop { background: transparent; }
</style>
