// ABOUTME: Native-close decision policy tests without launching Wails or touching user data.
// ABOUTME: Exercises fail-closed results, safe defaults and concurrent close deduplication.
package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func TestDesktopCloseConsent(t *testing.T) {
	for _, choice := range []string{"", "No", "Cancel", "unexpected", "Yes"} {
		t.Run(choice, func(t *testing.T) {
			guard := newDesktopCloseGuard(func(_ context.Context, options runtime.MessageDialogOptions) (string, error) {
				if options.Type != runtime.QuestionDialog || options.DefaultButton != "No" || options.CancelButton != "No" {
					t.Fatalf("unsafe dialog defaults: %+v", options)
				}
				if len(options.Buttons) != 2 || options.Buttons[0] != "No" || options.Buttons[1] != "Yes" || !strings.Contains(options.Message, "may still complete") {
					t.Fatalf("missing choices or uncertain outcome disclosure: %+v", options)
				}
				return choice, nil
			})
			if prevent := guard(context.Background()); prevent != (choice != "Yes") {
				t.Fatalf("choice %q: prevent=%v", choice, prevent)
			}
		})
	}
}

func TestDesktopCloseFailsClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	guard := newDesktopCloseGuard(func(context.Context, runtime.MessageDialogOptions) (string, error) {
		return "Yes", errors.New("native dialog unavailable")
	})
	if !guard(ctx) {
		t.Fatal("dialog error allowed closure")
	}
	guard = newDesktopCloseGuard(func(context.Context, runtime.MessageDialogOptions) (string, error) {
		cancel()
		return "Yes", nil
	})
	if !guard(ctx) {
		t.Fatal("cancelled context allowed closure")
	}
	guard = newDesktopCloseGuard(func(context.Context, runtime.MessageDialogOptions) (string, error) {
		t.Fatal("dialog invoked with cancelled context")
		return "Yes", nil
	})
	if !guard(ctx) {
		t.Fatal("already-cancelled context allowed closure")
	}
}

func TestDesktopCloseDeduplicatesPendingDialogs(t *testing.T) {
	entered, release, result := make(chan struct{}), make(chan struct{}), make(chan bool, 1)
	guard := newDesktopCloseGuard(func(context.Context, runtime.MessageDialogOptions) (string, error) {
		close(entered)
		<-release
		return "No", nil
	})
	go func() { result <- guard(context.Background()) }()
	<-entered
	for i := 0; i < 20; i++ {
		if !guard(context.Background()) {
			t.Fatal("concurrent close bypassed pending consent")
		}
	}
	close(release)
	if !<-result {
		t.Fatal("No did not prevent closing")
	}
}
