package main

import (
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/raspberry-switch-control/nscontroller/js"
	"github.com/pborman/getopt/v2"
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

	js, err := js.NewJs(*device)
	common.Checke(err)
	defer js.Close()

	for ;; {
		_, err := js.Read()
		common.Checke(err)
	}


	return 0
}

func main() {
	common.RunAndExit(realMain)
}
