package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/raspberry-switch-control/nscontroller"
	"github.com/omakoto/raspberry-switch-control/nscontroller/js"
	"github.com/pborman/getopt/v2"
)

var (
	debug    = getopt.BoolLong("debug", 'd', "Enable debug output")
	joystick = getopt.StringLong("joystick", 'j', "/dev/input/js0", "Specify joystick device file")
	out      = getopt.StringLong("out", 'o', "/dev/stdout", "Specify backend file")

	myName = common.MustGetBinName()
)

func mustGetDispatcher(js *js.Js) nscontroller.JoystickDispatcher {
	if strings.Contains(js.Name, "X-Box One") || strings.Contains(js.Name, "Xbox") {
		return nscontroller.XBoxOneJoystickDispatcher
	}
	if strings.Contains(js.Name, "Nintendo Switch Pro Controller") {
		return nscontroller.NSProJoystickDispatcher
	}
	if strings.Contains(js.Name, "Sony Interactive Entertainment Wireless Controller") {
		return nscontroller.PsJoystickDispatcher
	}
	// Old PS4 controller: https://github.com/omakoto/raspberry-switch-control/issues/4
	if strings.Contains(js.Name, "Sony Computer Entertainment Wireless Controller") {
		return nscontroller.PsJoystickDispatcher
	}
	common.Fatalf("Unknown joystick: %s", js.Name)
	return nil
}

func realMain() int {
	getopt.Parse()

	if *debug {
		common.DebugEnabled = true
	}

	out, err := os.OpenFile(*out, os.O_WRONLY, 0)
	common.Checkf(err, "open failed")

	js, err := js.NewJs(*joystick)
	common.Checke(err)

	backend, err := nscontroller.NewBackendConsumer(out)
	common.Checke(err)
	defer backend.Close()

	autoFirer := nscontroller.NewAutoFirer(backend.Consume)
	defer autoFirer.Close()

	joystick, err := nscontroller.NewJoystickInput(js, mustGetDispatcher(js), autoFirer.Consume)
	common.Checke(err)
	defer joystick.Close()

	stdinProxy, err := nscontroller.NewStreamInput(os.Stdin, backend.Consume)
	common.Checke(err)
	defer stdinProxy.Close()

	autoFirer.Run()
	joystick.Run()
	stdinProxy.Run()

	fmt.Printf("nsfrontend started: Accepting command from stdin... (^D to finish)\n")

	// ^D to finish
	stdinProxy.WaitClose()

	common.Debugf("%s finishing", myName)

	return 0
}

func main() {
	common.RunAndExit(realMain)
}
