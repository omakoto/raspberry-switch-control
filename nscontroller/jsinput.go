package nscontroller

import (
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/raspberry-switch-control/nscontroller/js"
	"io"
)

type JoystickDispatcher func(ev *js.JoystickEvent, ch chan<- Event)

type JoystickInput struct {
	js         *js.Js
	dispatcher JoystickDispatcher
	con        Consumer
}

var _ Worker = (*JoystickInput)(nil)

func NewJoystickInput(js *js.Js, dispatcher JoystickDispatcher, con Consumer) (*JoystickInput, error) {
	return &JoystickInput{js, dispatcher, con}, nil
}

func (j *JoystickInput) Close() error {
	return j.js.Close()
}

func (j *JoystickInput) Run() {
	go func() {
		next := j.con.Intake()
		for {
			ev, err := j.js.Read()
			if err == io.EOF {
				common.Debug("Joystick closing")
				return
			}
			common.Checke(err)
			common.Debugf("Joystick input=%x", ev.Element.Number)

			j.dispatcher(&ev, next)
		}
	}()
}
