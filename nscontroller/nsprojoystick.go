package nscontroller

import "github.com/omakoto/raspberry-switch-control/nscontroller/js"

// NSProJoystickDispatcher is a dispatcher for the Switch Pro controller.
func NSProJoystickDispatcher(ev *js.JsEvent, ch chan<- Event) {
	var action Action = ActionNone

	value := ev.Value

	switch ev.Element.Number {
	case 0x00: // "x",  // switch/xbox L-stick
		action = ActionAxisLX
	case 0x01: // "y",  // switch/xbox L-stick
		action = ActionAxisLY
	case 0x03: // "rx", // switch/xbox R-stick
		action = ActionAxisRX
	case 0x04: // "ry", // switch/xbox R-stick
		action = ActionAxisRY

	case 0x130: // "a", // switch B / xbox A
		action = ActionButtonB
	case 0x131: // "b", // switch A / xbox B
		action = ActionButtonA
	case 0x133: // "x",      // switch Y / xbox X
		action = ActionButtonX
	case 0x134: // "y",      // switch X / xbox Y
		action = ActionButtonY
	case 0x135: // "z"
		action = ActionButtonCapture

	case 0x136: // "tl",     // switch L / xbox L
		action = ActionButtonL
	case 0x137: // "tr",     // switch R / xbox R
		action = ActionButtonR

	case 0x138: // "tl2",    // switch LZ
		action = ActionButtonLZ
	case 0x139: // "tr2",    // switch RZ
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
