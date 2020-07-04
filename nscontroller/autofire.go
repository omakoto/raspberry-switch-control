package nscontroller

import (
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/raspberry-switch-control/nscontroller/utils"
	"github.com/pborman/getopt/v2"
	"time"
)

var (
	tickInterval = getopt.IntLong("tick", 't', 10, "Tick interval in milliseconds")
)

type AutofireMode int

const (
	AutofireModeDeactivated = AutofireMode(iota)
	AutofireModeNormal
	AutofireModeInvert
	AutofireModeToggle
)

type buttonState struct {
	mode AutofireMode

	// interval is the autofire interval for each button.
	interval time.Duration

	// autoLastTimestamp is the timestamp of the last on or off autofire event.
	autoLastTimestamp time.Time

	// realButtonPressed is whether the button is actually pressed or not.
	realButtonPressed bool

	// autofireEnabled is whether autofire is on (so events will be continuously produced) or off.
	autofireOn bool

	// lastValue is the last reported value to the next consumer.
	lastValue bool
}

type AutoFirer struct {
	syncer *utils.Synchronized
	next   Consumer
	states []buttonState
	ticker *time.Ticker
	stop   chan bool
}

var _ Worker = (*AutoFirer)(nil)

func NewAutoFirer(next Consumer) *AutoFirer {
	return &AutoFirer{
		utils.NewSynchronized(),
		next,
		make([]buttonState, ActionButtonLast),
		nil,
		nil,
	}
}

func (af *AutoFirer) Run() {
	af.syncer.Run(func() {
		if af.ticker != nil {
			common.Fatal("AutoFirer already running")
			return
		}
		common.Debug("AutoFirer started")
		af.ticker = time.NewTicker(time.Duration(*tickInterval) * time.Millisecond)
		af.stop = make(chan bool)

		ticker := af.ticker
		stop := af.stop

		go func() {
		loop:
			for {
				select {
				case <-ticker.C:
					af.tick()
				case <-stop:
					break loop
				}
			}
			af.syncer.Run(func() {
				af.ticker = nil
				af.stop = nil
			})
			common.Debug("AutoFirer stopped")
		}()
	})
}

func (af *AutoFirer) Close() error {
	af.syncer.Run(func() {
		if af.ticker != nil {
			common.Debug("AutoFirer stopping")
			af.stop <- true
		} else {
			common.Debug("AutoFirer not running")
		}
	})
	return nil
}

func (af *AutoFirer) SetAutofire(a Action, mode AutofireMode, interval time.Duration) {
	common.OrFatalf(interval >= 0, "interval must be >= 0 but was: %d", interval)
	af.syncer.Run(func() {
		af.states[a].mode = mode
		af.states[a].interval = interval

		af.setAutofireLocked(&Event{time.Now(), a, 0}, true)
	})
}

func (af *AutoFirer) setAutofireLocked(ev *Event, force bool) {
	bs := &af.states[ev.Action]
	pressed := ev.Value == 1

	if bs.realButtonPressed == pressed && !force {
		return // Button state hasn't changed; ignore.
	}

	bs.realButtonPressed = pressed

	switch bs.mode {
	case AutofireModeDeactivated:
		af.next(ev)
	case AutofireModeNormal:
		bs.autofireOn = ev.pressed()
		af.sendAutofireEventLocked(ev.Timestamp, ev.Action, pressed)
	case AutofireModeInvert:
		bs.autofireOn = !ev.pressed()
		af.sendAutofireEventLocked(ev.Timestamp, ev.Action, !pressed)
	case AutofireModeToggle:
		if pressed {
			bs.autofireOn = !bs.autofireOn
			af.sendAutofireEventLocked(ev.Timestamp, ev.Action, bs.autofireOn)
		}
	}
}

func (af *AutoFirer) sendAutofireEventLocked(timestamp time.Time, a Action, pressed bool) {
	bs := &af.states[a]

	ev := Event{Timestamp: timestamp, Action: a, Value: BoolToValue(pressed)}
	af.next(&ev)

	bs.autoLastTimestamp = ev.Timestamp
	bs.lastValue = pressed
}

func (af *AutoFirer) Consume(ev *Event) {
	af.syncer.Run(func() {
		if ev.Action.isButton() {
			af.setAutofireLocked(ev, false)
		} else {
			// Just forward any axis events.
			af.next(ev)
		}
	})
}

func (af *AutoFirer) tick() {
	af.syncer.Run(func() {
		common.Debug("AutoFirer tick")

		now := time.Now()
		for a := ActionButtonStart; a < ActionButtonLast; a++ {
			bs := &af.states[a]

			if bs.mode == AutofireModeDeactivated || !bs.autofireOn {
				continue
			}
			nextTimestamp := bs.autoLastTimestamp.Add(bs.interval)
			if nextTimestamp.After(now) {
				return
			}

			// Synthesis an event.
			af.sendAutofireEventLocked(nextTimestamp, a, !bs.lastValue)
		}

	})
}
