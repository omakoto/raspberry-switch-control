package main

import (
	"bufio"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/raspberry-switch-control/nscontroller"
	"github.com/omakoto/raspberry-switch-control/nscontroller/js"
	"github.com/pborman/getopt/v2"
	"os"
)

var (
	debug  = getopt.BoolLong("debug", 'd', "Enable debug output")
	joystick = getopt.StringLong("joystick", 'j', "/dev/input/js0", "Specify joystick device file")
	out = getopt.StringLong("out", 'o', "/dev/stdout", "Specify backend stdin")
)

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

	joystick, err := nscontroller.NewJoystickInput(js, nscontroller.XBoxOneJoystickDispatcher, backend)
	common.Checke(err)
	defer joystick.Close()

	backend.Run()
	joystick.Run()

	// Wait for enter press
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()


	return 0
}

func main() {
	common.RunAndExit(realMain)
}
