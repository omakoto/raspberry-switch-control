// +build linux

package js

// Based on: https://gist.githubusercontent.com/rdb/8864666/raw/516178252bbe1cfe8067145b11223ee54c5d9698/js_linux.py
// API reference: https://www.kernel.org/doc/Documentation/input/joystick-api.txt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/omakoto/go-common/src/common"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"time"
	"unsafe"
)

// TODO Introduce constants

/**
 Button layout: Pro controller
    X
   Y A
    B

Button layout: X-box one controller
    Y
   X B
    A
*/

var axisNameMap = map[int]string{
	0x00: "x",  // switch/xbox L-stick
	0x01: "y",  // switch/xbox L-stick
	0x02: "z",  // xbox L2 [-1..1]
	0x03: "rx", // switch/xbox R-stick
	0x04: "ry", // switch/xbox R-stick
	0x05: "rz", // xbox R2 [-1...1]
	0x06: "trottle",
	0x07: "rudder",
	0x08: "wheel",
	0x09: "gas",
	0x0a: "brake",
	0x10: "hat0x", // switch/xbox D-pad
	0x11: "hat0y", // switch/xbox D-pad
	0x12: "hat1x",
	0x13: "hat1y",
	0x14: "hat2x",
	0x15: "hat2y",
	0x16: "hat3x",
	0x17: "hat3y",
	0x18: "pressure",
	0x19: "distance",
	0x1a: "tilt_x",
	0x1b: "tilt_y",
	0x1c: "tool_width",
	0x20: "volume",
	0x28: "misc",
}

var buttonNameMap = map[int]string{
	0x120: "trigger",
	0x121: "thumb",
	0x122: "thumb2",
	0x123: "top",
	0x124: "top2",
	0x125: "pinkie",
	0x126: "base",
	0x127: "base2",
	0x128: "base3",
	0x129: "base4",
	0x12a: "base5",
	0x12b: "base6",
	0x12f: "dead",
	0x130: "a", // switch B / xbox A
	0x131: "b", // switch A / xbox B
	0x132: "c",
	0x133: "x",      // switch Y / xbox X
	0x134: "y",      // switch X / xbox Y
	0x135: "z",      // switch Capture
	0x136: "tl",     // switch L / xbox L
	0x137: "tr",     // switch R / xbox R
	0x138: "tl2",    // switch LZ
	0x139: "tr2",    // switch RZ
	0x13a: "select", // switch - / xbox select
	0x13b: "start",  // switch + / xbox start
	0x13c: "mode",   // switch Home / xbox center
	0x13d: "thumbl",
	0x13e: "thumbr",

	0x220: "dpad_up",
	0x221: "dpad_down",
	0x222: "dpad_left",
	0x223: "dpad_right",

	// XBox 360 controller uses these codes.
	0x2c0: "dpad_left",
	0x2c1: "dpad_right",
	0x2c2: "dpad_up",
	0x2c3: "dpad_down",
}

// Element represents a single axis or button.
type Element struct {
	// Number is the number given to the axis/button.
	Number int
	// Name is the name of the axis/button.
	Name string
	// Name is the last known value of the axis/button in the range of [-1..1].
	Value float64
	// Name is the initial value of the axis/button in the range of [-1..1].
	InitialValue float64
}

func (e *Element) setInitialValue() {
	e.InitialValue = e.InitialValue
}

// Js represents a joystick input device.
type Js struct {
	DevicePath string
	Name       string
	NumAxes    int
	NumButtons int
	Axes       []Element
	Buttons    []Element
	in         io.ReadCloser
}

// JsEvent is a single joystick event.
type JsEvent struct {
	Timestamp time.Duration
	Value     float64
	Element   *Element
}

// NewJs creates a new Js instance with the given device file.
func NewJs(device string) (*Js, error) {
	common.Debugf("Opening %s ...", device)
	in, err := os.OpenFile(device, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("unable to open %#v: %w", device, err)
	}

	js := &Js{DevicePath: device, in: in}

	// Get num axes and buttons.
	js.NumAxes, err = unix.IoctlGetInt(int(in.Fd()), jsiocgaxes)
	if err != nil {
		return nil, fmt.Errorf("unable to get number of axes of %#v: %w", device, err)
	}
	js.NumButtons, err = unix.IoctlGetInt(int(in.Fd()), jsiocgbuttons)
	if err != nil {
		return nil, fmt.Errorf("unable to get number of buttons of %#v: %w", device, err)
	}

	// Get device name.
	nameBuf := make([]byte, 256, 256)
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, in.Fd(), uintptr(jsiocgnameBase+(0x10000*len(nameBuf))), uintptr(unsafe.Pointer(&nameBuf[0])))
	if errno != 0 {
		return nil, fmt.Errorf("unable to get device name of %#v: %w", device, errno)
	}
	js.Name = string(bytes.TrimRight(nameBuf, "\000"))

	// Get the axis names.
	axisCodes := make([]byte, js.NumAxes, js.NumAxes)
	_, _, errno = unix.Syscall(unix.SYS_IOCTL, in.Fd(), uintptr(jsiocgaxmap), uintptr(unsafe.Pointer(&axisCodes[0])))
	if errno != 0 {
		return nil, fmt.Errorf("unable to get axis map of %#v: %w", device, errno)
	}
	js.Axes = make([]Element, js.NumAxes)
	for i, v := range axisCodes {
		js.Axes[i].Number = int(v)
		name, found := axisNameMap[int(v)]
		if !found {
			name = fmt.Sprintf("unknown:0x%x", v)
		}
		js.Axes[i].Name = name
	}

	// Get the button names.
	buttonCodes := make([]uint16, js.NumButtons, js.NumButtons)
	_, _, errno = unix.Syscall(unix.SYS_IOCTL, in.Fd(), uintptr(jsiocgbtnmap), uintptr(unsafe.Pointer(&buttonCodes[0])))
	if errno != 0 {
		return nil, fmt.Errorf("unable to get button map of %#v: %w", device, errno)
	}
	js.Buttons = make([]Element, js.NumButtons)
	for i, v := range buttonCodes {
		js.Buttons[i].Number = int(v)
		name, found := buttonNameMap[int(v)]
		if !found {
			name = fmt.Sprintf("unknown:0x%x", v)
		}
		js.Buttons[i].Name = name
	}

	//// Read the initial state. -> not working
	//common.Debug("Reading initial state...")
	//timeout := unix.Timeval{}
	//timeout.Sec = 1
	//for {
	//	fdSet := &unix.FdSet{}
	//	fdSet.Bits[0] = 1 << in.Fd()
	//	s, err := unix.Select(1, fdSet, nil, nil, &timeout)
	//	common.Check(err, "select")
	//	common.Debugf("select returned %d", s)
	//	if s < 1 {
	//		break
	//	}
	//	_, err = js.Read()
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	js.setInitialValues()

	common.Debugf("%s ready to read", js.DevicePath)
	common.Dump("Js:", &js)

	return js, nil
}

func (js *Js) setInitialValues() {
	for i := 0; i < len(js.Axes); i++ {
		js.Axes[i].setInitialValue()
	}
	for i := 0; i < len(js.Buttons); i++ {
		js.Buttons[i].setInitialValue()
	}
}

func (js *Js) Close() error {
	if js.in == nil {
		return nil
	}
	err := js.in.Close()
	js.in = nil
	return err
}

const (
	jsEventButton = 0x01 // button pressed/released
	jsEventAxis   = 0x02 // joystick moved
	jsEventInit   = 0x80 // initial state of device

	jsiocgnameBase = 0x80006a13
	jsiocgaxes     = 0x80016a11
	jsiocgbuttons  = 0x80016a12
	jsiocgaxmap    = 0x80406a32
	jsiocgbtnmap   = 0x80406a34
)

// osJsEvent is a single event from the joystick device.
type osJsEvent struct {
	// Time is event timestamp in milliseconds
	Time uint32
	// Value is: fix an axis, [-32767 .. 32767]. for a button, 1 (pressed) or 0 (released).
	Value int16
	// EventType is a bit field of JsEventXxx values.
	EventType uint8
	// Number is an axis or button number, 0-based.
	Number uint8
}

func (js *Js) Read() (JsEvent, error) {
	event := JsEvent{}

	// Read the OS event.
	var oev osJsEvent
	err := binary.Read(js.in, binary.LittleEndian, &oev)
	if err == io.EOF {
		return event, err
	}
	if err != nil {
		return event, fmt.Errorf("unable to read from %#v: %w", js.DevicePath, err)
	}
	common.Dump("OsEvent:", &oev)

	// Convert to the result.
	event.Timestamp = time.Millisecond * time.Duration(oev.Time)

	switch oev.EventType &^ jsEventInit {
	case jsEventAxis:
		event.Element = &js.Axes[oev.Number]
		event.Value = float64(oev.Value) / 32767.0
	case jsEventButton:
		event.Element = &js.Buttons[oev.Number]
		event.Value = 0
		if event.Value != 0 {
			event.Value = 1
		}
	}
	event.Element.Value = event.Value
	common.Dump("Event:", &event)

	return event, nil
}
