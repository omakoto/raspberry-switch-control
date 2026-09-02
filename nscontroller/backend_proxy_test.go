package nscontroller

// This file contains unit tests for BackendProxy to verify command mapping, concurrency, closing, and panic behavior on errors.

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"
)

type mockWriteCloser struct {
	bytes.Buffer
	closed bool
	err    error
}

func (m *mockWriteCloser) Close() error {
	m.closed = true
	return m.err
}

func (m *mockWriteCloser) Write(p []byte) (n int, err error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.Buffer.Write(p)
}

func TestBackendProxy_Mappings(t *testing.T) {
	tests := []struct {
		action Action
		val    float64
		want   string
	}{
		{ActionButtonA, 1.0, "a 1\n"},
		{ActionButtonB, 0.0, "b 0\n"},
		{ActionButtonX, 1.0, "x 1\n"},
		{ActionButtonY, 0.0, "y 0\n"},
		{ActionButtonMinus, 1.0, "- 1\n"},
		{ActionButtonPlus, 0.0, "+ 0\n"},
		{ActionButtonHome, 1.0, "h 1\n"},
		{ActionButtonCapture, 0.0, "c 0\n"},
		{ActionButtonDpadUp, 1.0, "pu 1\n"},
		{ActionButtonDpadDown, 0.0, "pd 0\n"},
		{ActionButtonDpadLeft, 1.0, "pl 1\n"},
		{ActionButtonDpadRight, 0.0, "pr 0\n"},
		{ActionButtonL, 1.0, "l1 1\n"},
		{ActionButtonR, 0.0, "r1 0\n"},
		{ActionButtonLZ, 1.0, "l2 1\n"},
		{ActionButtonRZ, 0.0, "r2 0\n"},
		{ActionButtonLeftStickPress, 1.0, "lp 1\n"},
		{ActionButtonRightStickPress, 0.0, "rp 0\n"},
		{ActionAxisLX, -0.75, "lx -0.75\n"},
		{ActionAxisLY, 0.5, "ly 0.5\n"},
		{ActionAxisRX, 0.25, "rx 0.25\n"},
		{ActionAxisRY, -0.1, "ry -0.1\n"},
		{ActionAxisLX, -0.0013428144169438765, "lx -0.0013\n"},
		{ActionAxisLY, 0.0013428144169438765, "ly 0.0013\n"},
		{ActionAxisRX, -0.00004, "rx 0\n"},
		{ActionAxisRY, 0.00004, "ry 0\n"},
		{ActionAxisLX, -0.00006, "lx -0.0001\n"},
		{ActionAxisLY, 0.00006, "ly 0.0001\n"},
		// Unmapped action should result in command being empty, writing " <value>\n"
		{ActionNone, 1.0, " 1\n"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("Action_%v", tc.action), func(t *testing.T) {
			mock := &mockWriteCloser{}
			proxy, err := NewBackendConsumer(mock)
			if err != nil {
				t.Fatalf("unexpected error creating BackendProxy: %v", err)
			}

			ev := NewEventFromAction(tc.action, tc.val)
			proxy.Consume(&ev)

			got := mock.String()
			if got != tc.want {
				t.Errorf("got message %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBackendProxy_Close(t *testing.T) {
	mock := &mockWriteCloser{}
	proxy, err := NewBackendConsumer(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := proxy.Close(); err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}

	if !mock.closed {
		t.Errorf("expected underlying writer to be closed")
	}
}

func TestBackendProxy_CloseError(t *testing.T) {
	expectedErr := errors.New("close error")
	mock := &mockWriteCloser{err: expectedErr}
	proxy, _ := NewBackendConsumer(mock)

	if err := proxy.Close(); err != expectedErr {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestBackendProxy_PanicOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on write failure, but got none")
		}
	}()

	mock := &mockWriteCloser{err: errors.New("write error")}
	proxy, _ := NewBackendConsumer(mock)
	ev := NewEventFromAction(ActionButtonA, 1.0)
	proxy.Consume(&ev)
}

func TestBackendProxy_Concurrency(t *testing.T) {
	mock := &mockWriteCloser{}
	proxy, _ := NewBackendConsumer(mock)

	const numGoroutines = 50
	const numEventsPerGoroutine = 20

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numEventsPerGoroutine; j++ {
				ev := NewEventFromAction(ActionButtonA, float64(id))
				proxy.Consume(&ev)
			}
		}(i)
	}

	wg.Wait()

	// Verify the count of lines written matches expected
	lines := bytes.Split(mock.Bytes(), []byte("\n"))
	// bytes.Split on trailing newline returns a trailing empty slice
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	expectedLines := numGoroutines * numEventsPerGoroutine
	if len(lines) != expectedLines {
		t.Errorf("got %d lines, want %d", len(lines), expectedLines)
	}
}

func TestBackendProxy_AxisRounding(t *testing.T) {
	testCases := []struct {
		input float64
		want  string
	}{
		{-0.0013428144169438765, "lx -0.0013\n"},
		{0.0013428144169438765, "lx 0.0013\n"},
		{-0.000049, "lx 0\n"},
		{0.000049, "lx 0\n"},
		{-0.00005, "lx -0.0001\n"},
		{0.00005, "lx 0.0001\n"},
		{0.0, "lx 0\n"},
		{-0.0, "lx 0\n"},
		{1.0, "lx 1\n"},
		{-1.0, "lx -1\n"},
		{0.5, "lx 0.5\n"},
		{-0.5, "lx -0.5\n"},
		{0.123456, "lx 0.1235\n"},
		{-0.123456, "lx -0.1235\n"},
	}

	for _, tc := range testCases {
		mock := &mockWriteCloser{}
		proxy, err := NewBackendConsumer(mock)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ev := NewEventFromAction(ActionAxisLX, tc.input)
		proxy.Consume(&ev)

		got := mock.String()
		if got != tc.want {
			t.Errorf("input %f: got %q, want %q", tc.input, got, tc.want)
		}
	}
}
