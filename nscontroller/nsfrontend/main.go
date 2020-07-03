package main

import (
	"bufio"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/raspberry-switch-control/nscontroller"
	"github.com/pborman/getopt/v2"
	"os"
)

var (
	debug  = getopt.BoolLong("debug", 'd', "Enable debug output")
	device = getopt.StringLong("device", 'f', "/dev/input/js0", "Specify device file")
)

func realMain() int {
	getopt.Parse()

	if device == nil {
		*device = "/dev/input/js"
	}
	if *debug {
		common.DebugEnabled = true
	}

	ji, err := nscontroller.NewJoystickInput(*device, nil)
	common.Checke(err)
	defer ji.Close()

	ji.Run()

	// Wait for enter press
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()


	return 0
}

func main() {
	common.RunAndExit(realMain)
}
