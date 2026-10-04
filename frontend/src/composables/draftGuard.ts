// ABOUTME: Shared policy for preventing accidental loss of an item draft.
// ABOUTME: Saves cannot be interrupted; dirty drafts and pending image work need consent.
export interface DraftState {dirty: boolean; working: boolean; submitting: boolean}
export function draftAtRisk(state: DraftState): boolean {
  return state.dirty || state.working || state.submitting
}
export function mayLeaveDraft(state: DraftState, confirmDiscard: () => boolean): boolean {
  if (state.submitting) return false
  return !(state.dirty || state.working) || confirmDiscard()
}
