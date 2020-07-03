package nscontroller

import "time"

type Action int;
type Value float64;

type Event struct {
	delay time.Duration
	action Action
	value Value
}
