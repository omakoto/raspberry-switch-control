package nscontroller

import (
	"io"
	"time"
)

// Buttons and axes for Switch.
const (
	ActionButtonA = iota
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

	ActionAxisLX
	ActionAxisLY

	ActionAxisRX
	ActionAxisRY
)

type Action int;
type Value float64;

type Event struct {
	Delay time.Duration
	Action Action
	Value Value
}

type Consumer interface {
	io.Closer
	Intake() <-chan Event
}

type Worker interface {
	io.Closer
	Run()
}