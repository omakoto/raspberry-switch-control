package nscontroller

import "github.com/omakoto/raspberry-switch-control/nscontroller/js"

// PsJoystickDispatcher is a dispatcher for the PS controller. (Only tested with a PS4 controller.)
func PsJoystickDispatcher(ev *js.JsEvent, ch chan<- Event) {
	var action Action = ActionNone

	value := ev.Value

	switch ev.Element.Number {
	case 0x00: // "x"
		action = ActionAxisLX
	case 0x01: // "y"
		action = ActionAxisLY
	case 0x03: // "rx"
		action = ActionAxisRX
	case 0x04: // "ry"
		action = ActionAxisRY

	case 0x130: // "a"
		action = ActionButtonB
	case 0x131: // "b"
		action = ActionButtonA
	case 0x133: // "x"
		action = ActionButtonX
	case 0x134: // "y"
		action = ActionButtonY

	case 0x136: // "tl"
		action = ActionButtonL
	case 0x137: // "tr"
		action = ActionButtonR
	case 0x138: // "tl2"
		action = ActionButtonLZ
	case 0x139: // "tr2"
		action = ActionButtonRZ


	case 0x13a: // "select", // switch - / xbox select
		action = ActionButtonMinus
	case 0x13b: // "start",  // switch + / xbox start
		action = ActionButtonPlus
	case 0x13c: // "mode",   // switch Home / xbox center
		action = ActionButtonHome

	case 0x13d: // "thumbl", // switch / xbox left stick press
		action = ActionButtonLeftStickPress
	case 0x13e: // "thumbr", // switch / xbox right stick press
		action = ActionButtonRightStickPress
	}
	if action != ActionNone {
		ch <- Event{-1, action, value}
		return
	}

	// D-pad requires a special handling
	switch ev.Element.Number {
	case 0x10: // "hat0x", // switch/xbox D-pad
		left := 0.0
		right := 0.0
		if value < 0 {
			left = 1
			right = 0
		} else if value > 0 {
			left = 0
			right = 1
		}
		ch <- Event{-1, ActionButtonDpadLeft, left}
		ch <- Event{-1, ActionButtonDpadRight, right}
	case 0x11: // "hat0y", // switch/xbox D-pad
		up := 0.0
		down := 0.0
		if value < 0 {
			up = 1
			down = 0
		} else if value > 0 {
			up = 0
			down = 1
		}
		ch <- Event{-1, ActionButtonDpadUp, up}
		ch <- Event{-1, ActionButtonDpadDown, down}
	}
}
