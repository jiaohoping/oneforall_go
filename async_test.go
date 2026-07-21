package oneforall_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	oneforall "github.com/jiaohoping/oneforall_go"
)

// makeEchoScanner creates a Scanner whose "oneforall.py" is a small shell
// script that prints lines and exits successfully. Used to test async event
// delivery without a real OneForAll installation.
func makeEchoScanner(t *testing.T, scriptLines string) *oneforall.Scanner {
	t.Helper()
	pyPath := findPython3(t)

	dir := t.TempDir()
	script := filepath.Join(dir, "oneforall.py")
	// Write a Python script that prints two progress lines then exits.
	os.WriteFile(script, []byte(scriptLines), 0755)

	// We need a fake result DB so processResult doesn't fail.
	// Point resultDBPath at a non-existent path to isolate the test to the
	// channel event ordering (processResult will fail, which we catch in
	// EventCompleted.Err — that's acceptable for these event-ordering tests).
	s, err := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(script),
		oneforall.WithTarget("example.com"),
		oneforall.WithResultDBPath("/nonexistent/result.sqlite3"),
	)
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	return s
}

func TestRunAsyncWithProgress_ChannelClosed(t *testing.T) {
	s := makeEchoScanner(t, `
import sys
print("progress 1")
print("progress 2")
sys.exit(0)
`)
	ch := s.RunAsyncWithProgress()

	timeout := time.After(10 * time.Second)
	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				// Channel closed — test passes.
				return
			}
			_ = evt
		case <-timeout:
			t.Fatal("RunAsyncWithProgress channel was not closed within 10 seconds")
		}
	}
}

func TestRunAsyncWithProgress_EventOrder(t *testing.T) {
	s := makeEchoScanner(t, `
import sys
print("line one")
print("line two")
sys.exit(0)
`)
	ch := s.RunAsyncWithProgress()

	var events []oneforall.ProgressEvent
	timeout := time.After(10 * time.Second)
	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				goto done
			}
			events = append(events, evt)
		case <-timeout:
			t.Fatal("timed out waiting for progress events")
		}
	}
done:
	if len(events) == 0 {
		t.Fatal("received no events")
	}

	// First event must be EventStarted.
	if events[0].Type != oneforall.EventStarted {
		t.Errorf("events[0].Type = %v, want EventStarted", events[0].Type)
	}

	// Last event must be EventCompleted.
	last := events[len(events)-1]
	if last.Type != oneforall.EventCompleted {
		t.Errorf("last event type = %v, want EventCompleted", last.Type)
	}

	// stdout lines must appear between Started and Completed.
	for i := 1; i < len(events)-1; i++ {
		if events[i].Type != oneforall.EventStdoutLine {
			t.Errorf("events[%d].Type = %v, want EventStdoutLine", i, events[i].Type)
		}
		if events[i].Line == "" {
			t.Errorf("events[%d].Line is empty", i)
		}
	}
}

func TestRunAsyncWithProgress_StdoutLines(t *testing.T) {
	s := makeEchoScanner(t, `
import sys
print("hello from oneforall")
print("scanning...")
sys.exit(0)
`)
	ch := s.RunAsyncWithProgress()

	var lines []string
	for evt := range ch {
		if evt.Type == oneforall.EventStdoutLine {
			lines = append(lines, evt.Line)
		}
	}

	if len(lines) < 2 {
		t.Errorf("expected at least 2 stdout lines, got %d: %v", len(lines), lines)
	}
}

func TestRunAsyncWithProgress_CompletedErrOnMissingDB(t *testing.T) {
	// The script exits 0 but the DB does not exist — EventCompleted.Err must be set.
	s := makeEchoScanner(t, `
import sys
sys.exit(0)
`)
	var completedEvt oneforall.ProgressEvent
	for evt := range s.RunAsyncWithProgress() {
		if evt.Type == oneforall.EventCompleted {
			completedEvt = evt
		}
	}
	if completedEvt.Err == nil {
		t.Error("EventCompleted.Err should be non-nil when result DB is missing")
	}
	if completedEvt.Result != nil {
		t.Error("EventCompleted.Result should be nil when there is an error")
	}
}

func TestRunAsyncWithProgress_MissingTarget(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	// No target configured → first event should be EventCompleted with error.
	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
	)
	var first oneforall.ProgressEvent
	for evt := range s.RunAsyncWithProgress() {
		first = evt
		break
	}
	// Drain remaining events.
	// (channel may already be closed since error is immediate)
	if first.Type != oneforall.EventCompleted || first.Err == nil {
		t.Errorf("expected immediate EventCompleted with error for missing target, got type=%v err=%v", first.Type, first.Err)
	}
}

// --- v0.4.0: RunWithProgress tests ---

func TestRunWithProgress_ReceivesEvents(t *testing.T) {
	s := makeEchoScanner(t, `
import sys
print("progress line")
sys.exit(0)
`)
	var types []oneforall.ProgressEventType
	s.RunWithProgress(func(evt oneforall.ProgressEvent) {
		types = append(types, evt.Type)
	})

	if len(types) == 0 {
		t.Fatal("no events received from RunWithProgress")
	}
	// Last event must be EventCompleted.
	if types[len(types)-1] != oneforall.EventCompleted {
		t.Errorf("last event type = %v, want EventCompleted", types[len(types)-1])
	}
}

func TestRunWithProgress_ReturnsError(t *testing.T) {
	s := makeEchoScanner(t, `
import sys
sys.exit(0)
`)
	// DB is missing → EventCompleted.Err set; RunWithProgress must surface it.
	_, err := s.RunWithProgress(func(_ oneforall.ProgressEvent) {})
	if err == nil {
		t.Error("expected error from RunWithProgress when DB missing")
	}
}

func TestRunWithProgress_EventOrderMatchesAsync(t *testing.T) {
	makeScanner := func() *oneforall.Scanner {
		return makeEchoScanner(t, `
import sys
print("line one")
print("line two")
sys.exit(0)
`)
	}

	// Collect types from RunWithProgress.
	var syncTypes []oneforall.ProgressEventType
	makeScanner().RunWithProgress(func(evt oneforall.ProgressEvent) {
		syncTypes = append(syncTypes, evt.Type)
	})

	// Collect types from RunAsyncWithProgress.
	var asyncTypes []oneforall.ProgressEventType
	for evt := range makeScanner().RunAsyncWithProgress() {
		asyncTypes = append(asyncTypes, evt.Type)
	}

	if len(syncTypes) != len(asyncTypes) {
		t.Errorf("sync event count %d != async event count %d", len(syncTypes), len(asyncTypes))
		return
	}
	for i := range syncTypes {
		if syncTypes[i] != asyncTypes[i] {
			t.Errorf("event[%d] sync=%v async=%v", i, syncTypes[i], asyncTypes[i])
		}
	}
}

func TestRunAsyncWithProgress_InitErrSurfaced(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	// Use WithTargets with a single domain (valid), then manually verify
	// the no-initErr path works via RunAsyncWithProgress.
	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTargets("example.com"),
		oneforall.WithResultDBPath("/nonexistent/result.sqlite3"),
	)

	var completedEvt oneforall.ProgressEvent
	for evt := range s.RunAsyncWithProgress() {
		if evt.Type == oneforall.EventCompleted {
			completedEvt = evt
		}
	}
	// No initErr; error should be from missing DB, not from option application.
	if completedEvt.Err == nil {
		t.Error("expected EventCompleted.Err from missing DB")
	}
}
