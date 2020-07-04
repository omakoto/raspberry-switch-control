package main

import (
	"bufio"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/raspberry-switch-control/nscontroller"
	"github.com/omakoto/raspberry-switch-control/nscontroller/js"
	"github.com/pborman/getopt/v2"
	"os"
	"strings"
)

var (
	debug    = getopt.BoolLong("debug", 'd', "Enable debug output")
	joystick = getopt.StringLong("joystick", 'j', "/dev/input/js0", "Specify joystick device file")
	out      = getopt.StringLong("out", 'o', "/dev/stdout", "Specify backend stdin")

	myName = common.MustGetBinName()
)

func mustGetDispatcher(js *js.Js) nscontroller.JoystickDispatcher {
	if strings.Contains(js.Name, "X-Box One") {
		return nscontroller.XBoxOneJoystickDispatcher
	}
	if strings.Contains(js.Name, "Nintendo Switch Pro Controller") {
		return nscontroller.NSProJoystickDispatcher
	}
	if strings.Contains(js.Name, "Sony Interactive Entertainment Wireless Controller") {
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

	backendStdin, err := os.OpenFile(*out, os.O_WRONLY, 0)
	common.Checkf(err, "open failed")

	js, err := js.NewJs(*joystick)
	common.Checke(err)

	backend, err := nscontroller.NewBackendConsumer(backendStdin)
	common.Checke(err)
	defer backend.Close()

	autoFirer := nscontroller.NewAutoFirer(backend.Consume)
	defer autoFirer.Close()

	autoFirer.SetAutofire(nscontroller.ActionButtonX, nscontroller.AutofireModeNormal, 16)
	autoFirer.SetAutofire(nscontroller.ActionButtonY, nscontroller.AutofireModeNormal, 16)

	joystick, err := nscontroller.NewJoystickInput(js, mustGetDispatcher(js), autoFirer.Consume)
	common.Checke(err)
	defer joystick.Close()

	autoFirer.Run()
	joystick.Run()

	// Wait for enter press
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	common.Debugf("%s finishing", myName)

	return 0
}

func main() {
	common.RunAndExit(realMain)
}
