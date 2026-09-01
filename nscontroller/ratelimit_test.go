package nscontroller

// This file contains unit tests for RateLimiter to verify axis throttling, peak deviation tracking, settling, button passthrough, independence across axes, flush on close, and concurrent execution.

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_Disabled(t *testing.T) {
	for _, rate := range []int{0, -1, -60} {
		var received []*Event
		var mu sync.Mutex
		consumer := func(ev *Event) {
			mu.Lock()
			defer mu.Unlock()
			received = append(received, ev)
		}

		limiter := NewRateLimiter(rate, consumer)
		defer limiter.Close()

		ev1 := NewEventFromAction(ActionAxisLX, 0.1)
		ev2 := NewEventFromAction(ActionAxisLX, 0.2)
		ev3 := NewEventFromAction(ActionButtonA, 1.0)

		limiter.Consume(&ev1)
		limiter.Consume(&ev2)
		limiter.Consume(&ev3)

		mu.Lock()
		count := len(received)
		mu.Unlock()

		if count != 3 {
			t.Errorf("expected 3 events for rate %d, got %d", rate, count)
		}
	}
}

func TestRateLimiter_ButtonPassthrough(t *testing.T) {
	var received []*Event
	var mu sync.Mutex
	consumer := func(ev *Event) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, ev)
	}

	limiter := NewRateLimiter(10, consumer) // 100ms interval
	defer limiter.Close()

	for i := 0; i < 5; i++ {
		ev := NewEventFromAction(ActionButtonA, 1.0)
		limiter.Consume(&ev)
	}

	mu.Lock()
	count := len(received)
	mu.Unlock()

	if count != 5 {
		t.Fatalf("expected 5 button events forwarded immediately, got %d", count)
	}
}

func TestRateLimiter_PeakTrackingAndSettling(t *testing.T) {
	var received []Event
	var mu sync.Mutex
	consumer := func(ev *Event) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, *ev)
	}

	// 20 Hz = 50ms interval
	limiter := NewRateLimiter(20, consumer)
	defer limiter.Close()

	// Initial event sent immediately
	ev0 := NewEventFromAction(ActionAxisLX, 0.05)
	limiter.Consume(&ev0)

	mu.Lock()
	if len(received) != 1 || received[0].Value != 0.05 {
		t.Fatalf("expected first event 0.05 immediately, got %+v", received)
	}
	mu.Unlock()

	// Transient flick: goes up to 1.0 and snaps back to 0.0 within the 50ms window
	ev1 := NewEventFromAction(ActionAxisLX, 0.4)
	ev2 := NewEventFromAction(ActionAxisLX, 1.0) // Peak
	ev3 := NewEventFromAction(ActionAxisLX, 0.0) // Return to neutral
	limiter.Consume(&ev1)
	limiter.Consume(&ev2)
	limiter.Consume(&ev3)

	// Wait for the first window to fire (50ms interval + buffer)
	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	if len(received) != 2 {
		t.Fatalf("expected 2 events after first window (initial + peak), got %d: %+v", len(received), received)
	}
	if received[1].Value != 1.0 {
		t.Errorf("expected peak event value 1.0, got %f", received[1].Value)
	}
	mu.Unlock()

	// Wait for the second settling window to fire (another 50ms interval + buffer)
	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 3 {
		t.Fatalf("expected 3 events total (initial + peak + settled), got %d: %+v", len(received), received)
	}
	if received[2].Value != 0.0 {
		t.Errorf("expected settled event value 0.0, got %f", received[2].Value)
	}
}

func TestRateLimiter_MultipleAxesIndependent(t *testing.T) {
	var received []Event
	var mu sync.Mutex
	consumer := func(ev *Event) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, *ev)
	}

	limiter := NewRateLimiter(20, consumer) // 50ms interval
	defer limiter.Close()

	evLX := NewEventFromAction(ActionAxisLX, 0.1)
	evLY := NewEventFromAction(ActionAxisLY, -0.1)

	limiter.Consume(&evLX)
	limiter.Consume(&evLY)

	// Both LX and LY should have their first event emitted immediately
	mu.Lock()
	if len(received) != 2 {
		t.Fatalf("expected 2 immediate events (one for LX, one for LY), got %d", len(received))
	}
	mu.Unlock()

	// Rapid updates to both
	evLX2 := NewEventFromAction(ActionAxisLX, 0.8)
	evLY2 := NewEventFromAction(ActionAxisLY, -0.8)
	limiter.Consume(&evLX2)
	limiter.Consume(&evLY2)

	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 4 {
		t.Fatalf("expected 4 total events, got %d: %+v", len(received), received)
	}
}

func TestRateLimiter_CloseFlushesPending(t *testing.T) {
	var received []Event
	var mu sync.Mutex
	consumer := func(ev *Event) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, *ev)
	}

	limiter := NewRateLimiter(10, consumer) // 100ms interval

	ev1 := NewEventFromAction(ActionAxisRX, 0.2)
	limiter.Consume(&ev1)

	ev2 := NewEventFromAction(ActionAxisRX, 0.9)
	limiter.Consume(&ev2)

	mu.Lock()
	if len(received) != 1 {
		t.Fatalf("expected 1 event before close, got %d", len(received))
	}
	mu.Unlock()

	// Close should flush the current resting position 0.9 immediately
	if err := limiter.Close(); err != nil {
		t.Fatalf("unexpected error closing limiter: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 2 {
		t.Fatalf("expected 2 events after close, got %d", len(received))
	}
	if received[1].Value != 0.9 {
		t.Errorf("expected flushed event value to be 0.9, got %f", received[1].Value)
	}
}

func TestRateLimiter_Concurrency(t *testing.T) {
	var receivedCount int
	var mu sync.Mutex
	consumer := func(ev *Event) {
		mu.Lock()
		defer mu.Unlock()
		receivedCount++
	}

	limiter := NewRateLimiter(60, consumer)
	defer limiter.Close()

	const numGoroutines = 20
	const eventsPerGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				if j%2 == 0 {
					ev := NewEventFromAction(ActionAxisLX, float64(j)/float64(eventsPerGoroutine))
					limiter.Consume(&ev)
				} else {
					ev := NewEventFromAction(ActionButtonA, 1.0)
					limiter.Consume(&ev)
				}
			}
		}(i)
	}

	wg.Wait()
}
