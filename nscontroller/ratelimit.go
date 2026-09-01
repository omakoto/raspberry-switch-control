package nscontroller

// This file implements the RateLimiter consumer that throttles analog axis event rates.

import (
	"io"
	"math"
	"time"

	"github.com/omakoto/raspberry-switch-control/nscontroller/utils"
)

type axisRateState struct {
	lastSentTime time.Time
	lastSentVal  float64
	currentVal   float64
	peakVal      float64
	maxDev       float64
	hasPending   bool
	timer        *time.Timer
}

type RateLimiter struct {
	syncer   *utils.Synchronized
	interval time.Duration
	next     Consumer
	axes     []axisRateState
	closed   bool
}

var _ io.Closer = (*RateLimiter)(nil)

// NewRateLimiter creates a RateLimiter that throttles axis events to rateLimitHz (e.g. 120Hz).
// If rateLimitHz <= 0, rate limiting is disabled and all events pass through immediately.
func NewRateLimiter(rateLimitHz int, next Consumer) *RateLimiter {
	var interval time.Duration
	if rateLimitHz > 0 {
		interval = time.Second / time.Duration(rateLimitHz)
	}
	return &RateLimiter{
		syncer:   utils.NewSynchronized(),
		interval: interval,
		next:     next,
		axes:     make([]axisRateState, ActionLast),
	}
}

// Consume processes an event. Button events and axis events (when rate limiting is disabled)
// are forwarded immediately. When rate limiting is active, axis events are throttled such that
// the largest change (maximum deviation from the last emitted value) within each throttling window
// is emitted, followed by an automatic settling update to the current physical position if different.
func (r *RateLimiter) Consume(ev *Event) {
	if !ev.Action.isAxis() || r.interval <= 0 {
		r.next(ev)
		return
	}

	r.syncer.Run(func() {
		if r.closed {
			return
		}

		state := &r.axes[ev.Action]
		now := time.Now()
		state.currentVal = ev.Value

		timeSinceLastSent := now.Sub(state.lastSentTime)

		if timeSinceLastSent >= r.interval && !state.hasPending {
			state.lastSentTime = now
			state.lastSentVal = ev.Value
			r.next(ev)
			return
		}

		dev := math.Abs(ev.Value - state.lastSentVal)
		if !state.hasPending || dev > state.maxDev {
			state.peakVal = ev.Value
			state.maxDev = dev
		}
		state.hasPending = true

		if state.timer == nil {
			remaining := r.interval - timeSinceLastSent
			if remaining < 0 {
				remaining = 0
			}
			action := ev.Action
			state.timer = time.AfterFunc(remaining, func() {
				r.onTimer(action)
			})
		}
	})
}

func (r *RateLimiter) onTimer(action Action) {
	r.syncer.Run(func() {
		if r.closed {
			return
		}

		state := &r.axes[action]
		state.timer = nil

		if !state.hasPending {
			return
		}

		now := time.Now()
		valToSend := state.peakVal
		currentVal := state.currentVal

		if valToSend != state.lastSentVal {
			state.lastSentTime = now
			state.lastSentVal = valToSend

			ev := Event{
				Timestamp: now,
				Action:    action,
				Value:     valToSend,
			}
			r.next(&ev)
		} else {
			state.lastSentTime = now
		}

		// If current physical position differs from what was emitted, schedule settling update.
		if currentVal != state.lastSentVal {
			state.hasPending = true
			state.peakVal = currentVal
			state.maxDev = math.Abs(currentVal - state.lastSentVal)

			state.timer = time.AfterFunc(r.interval, func() {
				r.onTimer(action)
			})
		} else {
			state.hasPending = false
			state.maxDev = 0
		}
	})
}

// Close stops all active timers and flushes the final current value for pending axes.
func (r *RateLimiter) Close() error {
	r.syncer.Run(func() {
		if r.closed {
			return
		}
		r.closed = true

		for a := ActionAxisStart; a < ActionAxisLast; a++ {
			state := &r.axes[a]
			if state.timer != nil {
				state.timer.Stop()
				state.timer = nil
			}
			if state.hasPending {
				finalVal := state.currentVal
				if finalVal != state.lastSentVal {
					ev := Event{
						Timestamp: time.Now(),
						Action:    a,
						Value:     finalVal,
					}
					r.next(&ev)
				}
				state.hasPending = false
			}
		}
	})
	return nil
}
