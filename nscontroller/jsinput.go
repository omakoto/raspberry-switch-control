package nscontroller

// This file implements the JoystickInput worker that reads from the OS joystick and pumps events.

import (
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/raspberry-switch-control/nscontroller/js"
	"io"
)

type JoystickDispatcher func(ev *js.JoystickEvent, con Consumer)

type JoystickInput struct {
	js         *js.Js
	dispatcher JoystickDispatcher
	next       Consumer
}

var _ Worker = (*JoystickInput)(nil)

func NewJoystickInput(js *js.Js, dispatcher JoystickDispatcher, next Consumer) (*JoystickInput, error) {
	return &JoystickInput{js, dispatcher, next}, nil
}

func (j *JoystickInput) Close() error {
	return j.js.Close()
}

func (j *JoystickInput) Run() {
	go func() {
		for {
			ev, err := j.js.Read()
			if err == io.EOF {
				common.Debug("Joystick closing")
				return
			}
			common.Checke(err)
			if ev.Element == nil {
				continue
			}
			common.Debugf("Joystick input=%x", ev.Element.Number)

			j.dispatcher(&ev, j.next)
		}
	}()
}
