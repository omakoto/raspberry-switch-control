package nscontroller

// This file contains unit tests for StreamInput command parsing, comments, and auto-releases.

import (
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockReadCloser wraps a strings.Reader to satisfy io.ReadCloser
type mockReadCloser struct {
	io.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}

func TestStreamInput_SingleButton(t *testing.T) {
	inputStr := "a\n"
	in := &mockReadCloser{Reader: strings.NewReader(inputStr)}

	var mu sync.Mutex
	events := make([]*Event, 0)

	consumer := func(ev *Event) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, ev)
	}

	si, err := NewStreamInput(in, consumer)
	if err != nil {
		t.Fatalf("Failed to create StreamInput: %v", err)
	}

	si.Run()
	si.WaitClose()

	// Immediately, we should have a press event
	mu.Lock()
	if len(events) != 1 {
		mu.Unlock()
		t.Fatalf("Expected 1 immediate event, got %d", len(events))
	}
	evPress := events[0]
	if evPress.Action != ActionButtonA || evPress.Value != 1 {
		mu.Unlock()
		t.Errorf("Expected press event for ActionButtonA (1.0), got %v (value: %f)", evPress.Action, evPress.Value)
	}
	mu.Unlock()

	// Wait for the release event (streamInputOffDelay is 60ms)
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(events) != 2 {
		t.Fatalf("Expected 2 events total after delay, got %d", len(events))
	}
	evRelease := events[1]
	if evRelease.Action != ActionButtonA || evRelease.Value != 0 {
		t.Errorf("Expected release event for ActionButtonA (0.0), got %v (value: %f)", evRelease.Action, evRelease.Value)
	}
}

func TestStreamInput_CommentAndWhitespace(t *testing.T) {
	inputStr := "   b  # press B button \n"
	in := &mockReadCloser{Reader: strings.NewReader(inputStr)}

	var mu sync.Mutex
	events := make([]*Event, 0)

	consumer := func(ev *Event) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, ev)
	}

	si, err := NewStreamInput(in, consumer)
	if err != nil {
		t.Fatalf("Failed to create StreamInput: %v", err)
	}

	si.Run()
	si.WaitClose()

	mu.Lock()
	if len(events) != 1 {
		mu.Unlock()
		t.Fatalf("Expected 1 immediate event, got %d", len(events))
	}
	evPress := events[0]
	if evPress.Action != ActionButtonB || evPress.Value != 1 {
		mu.Unlock()
		t.Errorf("Expected press event for ActionButtonB, got action:%v value:%f", evPress.Action, evPress.Value)
	}
	mu.Unlock()

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(events) != 2 {
		t.Fatalf("Expected release event after delay, got %d events", len(events))
	}
	evRelease := events[1]
	if evRelease.Action != ActionButtonB || evRelease.Value != 0 {
		t.Errorf("Expected release event for ActionButtonB, got action:%v value:%f", evRelease.Action, evRelease.Value)
	}
}

func TestStreamInput_DiagonalDpad(t *testing.T) {
	// 'pdr' is Down + Right
	inputStr := "pdr\n"
	in := &mockReadCloser{Reader: strings.NewReader(inputStr)}

	var mu sync.Mutex
	events := make([]*Event, 0)

	consumer := func(ev *Event) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, ev)
	}

	si, err := NewStreamInput(in, consumer)
	if err != nil {
		t.Fatalf("Failed to create StreamInput: %v", err)
	}

	si.Run()
	si.WaitClose()

	mu.Lock()
	if len(events) != 2 {
		mu.Unlock()
		t.Fatalf("Expected 2 immediate press events, got %d", len(events))
	}

	// Verify both D-pad buttons were pressed
	actionsPressed := map[Action]bool{
		events[0].Action: true,
		events[1].Action: true,
	}
	if !actionsPressed[ActionButtonDpadDown] || !actionsPressed[ActionButtonDpadRight] {
		mu.Unlock()
		t.Errorf("Expected press events for DpadDown and DpadRight, got events: %v, %v", events[0].Action, events[1].Action)
	}
	if events[0].Value != 1 || events[1].Value != 1 {
		mu.Unlock()
		t.Error("Expected press values to be 1.0")
	}
	mu.Unlock()

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(events) != 4 {
		t.Fatalf("Expected 4 events total (2 presses, 2 releases), got %d", len(events))
	}
	actionsReleased := map[Action]bool{
		events[2].Action: true,
		events[3].Action: true,
	}
	if !actionsReleased[ActionButtonDpadDown] || !actionsReleased[ActionButtonDpadRight] {
		t.Errorf("Expected release events for DpadDown and DpadRight, got events: %v, %v", events[2].Action, events[3].Action)
	}
	if events[2].Value != 0 || events[3].Value != 0 {
		t.Error("Expected release values to be 0.0")
	}
}

func TestStreamInput_InvalidAndSkip(t *testing.T) {
	inputStr := "invalid_command\n# only a comment\n\nx\n"
	in := &mockReadCloser{Reader: strings.NewReader(inputStr)}

	var mu sync.Mutex
	events := make([]*Event, 0)

	consumer := func(ev *Event) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, ev)
	}

	si, err := NewStreamInput(in, consumer)
	if err != nil {
		t.Fatalf("Failed to create StreamInput: %v", err)
	}

	si.Run()
	si.WaitClose()

	mu.Lock()
	// 'invalid_command', comment, and blank line should be skipped. Only 'x' should trigger.
	if len(events) != 1 {
		mu.Unlock()
		t.Fatalf("Expected 1 immediate event for 'x', got %d", len(events))
	}
	evPress := events[0]
	if evPress.Action != ActionButtonX || evPress.Value != 1 {
		mu.Unlock()
		t.Errorf("Expected press event for ActionButtonX, got action:%v value:%f", evPress.Action, evPress.Value)
	}
	mu.Unlock()
}

func TestStreamInput_AllSwitchCases(t *testing.T) {
	tests := []struct {
		cmd             string
		expectedActions []Action
		isValid         bool
	}{
		// Buttons
		{"a", []Action{ActionButtonA}, true},
		{"b", []Action{ActionButtonB}, true},
		{"x", []Action{ActionButtonX}, true},
		{"y", []Action{ActionButtonY}, true},
		{"h", []Action{ActionButtonHome}, true},
		{"c", []Action{ActionButtonCapture}, true},
		{"m", []Action{ActionButtonMinus}, true},
		{"-", []Action{ActionButtonMinus}, true},
		{"p", []Action{ActionButtonPlus}, true},
		{"+", []Action{ActionButtonPlus}, true},
		{"l1", []Action{ActionButtonL}, true},
		{"l2", []Action{ActionButtonLZ}, true},
		{"r1", []Action{ActionButtonR}, true},
		{"r2", []Action{ActionButtonRZ}, true},

		// D-pad
		{"pu", []Action{ActionButtonDpadUp}, true},
		{"pd", []Action{ActionButtonDpadDown}, true},
		{"pl", []Action{ActionButtonDpadLeft}, true},
		{"pr", []Action{ActionButtonDpadRight}, true},
		{"pur", []Action{ActionButtonDpadUp, ActionButtonDpadRight}, true},
		{"pul", []Action{ActionButtonDpadUp, ActionButtonDpadLeft}, true},
		{"pdr", []Action{ActionButtonDpadDown, ActionButtonDpadRight}, true},
		{"pdl", []Action{ActionButtonDpadDown, ActionButtonDpadLeft}, true},

		// Stick press
		{"lp", []Action{ActionButtonLeftStickPress}, true},
		{"rp", []Action{ActionButtonRightStickPress}, true},

		// Empty and comment cases
		{"", []Action{}, true},
		{"# comment", []Action{}, true},

		// Invalid cases
		{"invalid", []Action{}, false},
		{"a 1", []Action{}, false},
	}

	for _, tt := range tests {
		in := &mockReadCloser{Reader: strings.NewReader("")}

		var mu sync.Mutex
		var actualActions []Action

		consumer := func(ev *Event) {
			mu.Lock()
			defer mu.Unlock()
			// We only care about press events in this test (Value == 1)
			if ev.Value == 1 {
				actualActions = append(actualActions, ev.Action)
			}
		}

		si, err := NewStreamInput(in, consumer)
		if err != nil {
			t.Fatalf("Failed to create StreamInput: %v", err)
		}

		res := si.ProcessLine(tt.cmd)
		if res != tt.isValid {
			t.Errorf("ProcessLine(%q) returned %t, expected %t", tt.cmd, res, tt.isValid)
		}

		mu.Lock()
		if len(actualActions) != len(tt.expectedActions) {
			t.Errorf("ProcessLine(%q) triggered %d actions, expected %d", tt.cmd, len(actualActions), len(tt.expectedActions))
		} else {
			for i, act := range tt.expectedActions {
				if actualActions[i] != act {
					t.Errorf("ProcessLine(%q) action[%d] was %v, expected %v", tt.cmd, i, actualActions[i], act)
				}
			}
		}
		mu.Unlock()
	}
}
