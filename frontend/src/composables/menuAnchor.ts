// ABOUTME: Anchors pointer and keyboard-opened menus to a focusable triggering control.
// ABOUTME: Keyboard context-menu coordinates fall back to the element bounds.
export function menuAnchor(event: MouseEvent): {x: number; y: number} {
  event.preventDefault()
  event.stopPropagation()
  const element = event.currentTarget as HTMLElement
  const opener = element.tabIndex >= 0 ? element : element.querySelector<HTMLElement>('button')
  opener?.focus()
  const rect = (opener || element).getBoundingClientRect()
  return event.type === 'contextmenu' && (event.clientX || event.clientY)
    ? {x: event.clientX, y: event.clientY}
    : {x: rect.left, y: rect.bottom}
}
