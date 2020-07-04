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

	// NumActionButtons is the number of buttons
	NumActionButtons

	ActionAxisLX
	ActionAxisLY

	ActionAxisRX
	ActionAxisRY
)

type Event struct {
	Timestamp time.Time
	Action    Action
	Value     float64
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
