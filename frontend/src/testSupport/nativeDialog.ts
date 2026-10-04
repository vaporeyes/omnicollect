// ABOUTME: Minimal native dialog lifecycle stub for DOM unit tests without a browser.
// ABOUTME: Does not claim to test browser focus trapping or inert background behavior.
export function stubNativeDialogs(): () => void {
  const prototype = HTMLDialogElement.prototype
  const originalShow = Object.getOwnPropertyDescriptor(prototype, 'showModal')
  const originalClose = Object.getOwnPropertyDescriptor(prototype, 'close')
  Object.defineProperty(prototype, 'showModal', {configurable: true, value(this: HTMLDialogElement) {this.open = true}})
  Object.defineProperty(prototype, 'close', {configurable: true, value(this: HTMLDialogElement) {this.open = false}})
  return () => {
    for (const [name, descriptor] of [['showModal', originalShow], ['close', originalClose]] as const) {
      if (descriptor) Object.defineProperty(prototype, name, descriptor)
      else Reflect.deleteProperty(prototype, name)
    }
  }
}
