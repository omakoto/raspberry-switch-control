package nscontroller

import (
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/raspberry-switch-control/nscontroller/js"
	"io"
)

type JoystickDispatcher interface {
	Dispatch(ev *js.JsEvent, con *Consumer)
}

type JoystickInput struct {
	js         *js.Js
	dispatcher JoystickDispatcher
	con        *Consumer
}

var _ Worker = (*JoystickInput)(nil)

func NewJoystickInput(device string, dispatcher JoystickDispatcher, con *Consumer) (*JoystickInput, error) {
	js, err := js.NewJs(device)
	if err != nil {
		return nil, err
	}
	return &JoystickInput{js, dispatcher, con}, nil
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
			common.Debugf("Joystick input=%x", ev.Element.Number)

			j.dispatcher.Dispatch(&ev, j.con)
		}
	}()
}
