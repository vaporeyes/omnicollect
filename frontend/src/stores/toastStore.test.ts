// ABOUTME: Unit tests for the toast notification Pinia store.
// ABOUTME: Verifies show, dismiss, and auto-dismiss with fake timers.
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {useToastStore} from './toastStore'

beforeEach(() => {
  setActivePinia(createPinia())
  vi.useFakeTimers()
})

afterEach(() => {
  vi.useRealTimers()
})

describe('show', () => {
  it('adds a toast to the array', () => {
    const store = useToastStore()
    store.show('Hello', 'info')

    expect(store.toasts).toHaveLength(1)
    expect(store.toasts[0].message).toBe('Hello')
    expect(store.toasts[0].type).toBe('info')
  })

  it('supports different toast types', () => {
    const store = useToastStore()
    store.show('Success', 'success')
    store.show('Error', 'error')

    expect(store.toasts[0].type).toBe('success')
    expect(store.toasts[1].type).toBe('error')
  })
})

describe('dismiss', () => {
  it('removes toast by ID', () => {
    const store = useToastStore()
    store.show('First', 'info')
    store.show('Second', 'info')

    const firstId = store.toasts[0].id
    store.dismiss(firstId)

    expect(store.toasts).toHaveLength(1)
    expect(store.toasts[0].message).toBe('Second')
  })
})

describe('auto-dismiss', () => {
  it('removes toast after duration', () => {
    const store = useToastStore()
    store.show('Temporary', 'info', 3000)

    expect(store.toasts).toHaveLength(1)

    vi.advanceTimersByTime(3000)

    expect(store.toasts).toHaveLength(0)
  })

  it('uses default 4000ms duration', () => {
    const store = useToastStore()
    store.show('Default duration', 'info')

    vi.advanceTimersByTime(3999)
    expect(store.toasts).toHaveLength(1)

    vi.advanceTimersByTime(1)
    expect(store.toasts).toHaveLength(0)
  })
})
