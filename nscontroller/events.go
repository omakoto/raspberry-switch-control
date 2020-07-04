package nscontroller

import (
	"io"
	"time"
)

type Action int

// Buttons and axes for Switch.
const (
	ActionNone = Action(iota)

	ActionButtonA
	ActionButtonB
	ActionButtonX
	ActionButtonY

	ActionButtonMinus
	ActionButtonPlus

	ActionButtonHome
	ActionButtonCapture

	ActionButtonL
	ActionButtonR
	ActionButtonLZ
	ActionButtonRZ

	ActionButtonDpadUp
	ActionButtonDpadDown
	ActionButtonDpadLeft
	ActionButtonDpadRight

	ActionButtonLeftStickPress
	ActionButtonRightStickPress

	// for synthetic events
	ActionButtonSynth1
	// for synthetic events
	ActionButtonSynth2
	// for synthetic events
	ActionButtonSynth3
	// for synthetic events
	ActionButtonSynth4

	ActionAxisLX
	ActionAxisLY

	ActionAxisRX
	ActionAxisRY

	ActionLast

	ActionButtonStart = ActionButtonA
	ActionButtonLast  = ActionAxisLX

	ActionAxisStart = ActionAxisLX
	ActionAxisLast  = ActionLast
)

func (a Action) isButton() bool {
	return ActionButtonStart <= a && a < ActionButtonLast
}

func (a Action) isAxis() bool {
	return ActionAxisStart <= a && a < ActionAxisLast
}

type Event struct {
	Timestamp time.Time
	Action    Action
	Value     float64
}

func (ev *Event) pressed() bool {
	return ev.Value == 1
}

type Consumer func(ev *Event)

type Worker interface {
	io.Closer
	Run()
}

func BoolToValue(pressed bool) float64 {
	if pressed {
		return 1
	}
	return 0
}

func ValueToBool(v float64) bool {
	return v == 1
}
