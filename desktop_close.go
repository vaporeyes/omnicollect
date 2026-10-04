// ABOUTME: Conservative native close confirmation independent of browser unload support.
// ABOUTME: Defaults to keeping the window open and deduplicates concurrent close requests.
package main

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Native window closure is not guaranteed to run browser beforeunload handlers.
// Always ask, including when the frontend is unavailable, rather than trusting a
// potentially stale dirty/busy flag. This is consent, not a promise of rollback.
func newDesktopCloseGuard(ask func(context.Context, runtime.MessageDialogOptions) (string, error)) func(context.Context) bool {
	var pending sync.Mutex
	return func(ctx context.Context) bool {
		if ctx.Err() != nil || !pending.TryLock() {
			return true
		}
		defer pending.Unlock()
		choice, err := ask(ctx, runtime.MessageDialogOptions{
			Type:          runtime.QuestionDialog,
			Title:         "Close OmniCollect?",
			Message:       "Closing can discard unsaved drafts and interrupt unfinished work. A request already sent may still complete; closing does not undo it. Choose No to keep working or wait for an operation to finish. Close anyway?",
			Buttons:       []string{"No", "Yes"},
			DefaultButton: "No",
			CancelButton:  "No",
		})
		return err != nil || ctx.Err() != nil || choice != "Yes"
	}
}
