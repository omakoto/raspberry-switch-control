package nscontroller

import (
	"io"
	"time"
)

const (
	ActionButtonA = iota
	ActionButtonB
	ActionButtonX
	ActionButtonY

	ActionButtonL1
	ActionButtonL2
	ActionButtonR1
	ActionButtonR2

	ActionButtonSelect
	ActionButtonStart

	ActionButtonHome
	ActionButtonCapture

	ActionUp = iota
	ActionUp = iota
	ActionUp = iota
	ActionUp = iota
	ActionUp = iota
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