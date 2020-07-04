package nscontroller

import (
	"io"
	"time"
)

// Buttons and axes for Switch.
const (
	ActionNone = iota

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

	// NumActionButton is the number of buttons
	NumActionButton

	ActionAxisLX
	ActionAxisLY

	ActionAxisRX
	ActionAxisRY
)

type Action int

type Event struct {
	timestamp time.Duration
	Action    Action
	Value     float64
}

type Consumer func(ev *Event)

type Worker interface {
	io.Closer
	Run()
}
