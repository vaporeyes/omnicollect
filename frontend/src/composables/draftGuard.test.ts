// ABOUTME: Tests the shared navigation/unload policy without native browser dialogs.
// ABOUTME: Covers cancelled discard, pending uploads, and uninterruptible saves.
import {expect, it, vi} from 'vitest'
import {draftAtRisk, mayLeaveDraft} from './draftGuard'

it('leaves clean drafts without asking', () => {
  const confirm = vi.fn()
  expect(mayLeaveDraft({dirty: false, working: false, submitting: false}, confirm)).toBe(true)
  expect(confirm).not.toHaveBeenCalled()
})
it('requires consent for dirty drafts and pending image work', () => {
  for (const state of [{dirty: true, working: false, submitting: false}, {dirty: false, working: true, submitting: false}]) {
    expect(draftAtRisk(state)).toBe(true)
    expect(mayLeaveDraft(state, () => false)).toBe(false)
    expect(mayLeaveDraft(state, () => true)).toBe(true)
  }
})
it('does not discard a save that may already have committed', () => {
  const confirm = vi.fn(() => true)
  const state = {dirty: false, working: false, submitting: true}
  expect(draftAtRisk(state)).toBe(true)
  expect(mayLeaveDraft(state, confirm)).toBe(false)
  expect(confirm).not.toHaveBeenCalled()
})
