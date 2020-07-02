package main

import (
	"bufio"
	"github.com/mzyy94/nscon"
	"github.com/omakoto/go-common/src/common"
	"github.com/pborman/getopt/v2"
	"os"
	"strconv"
	"strings"
)

var (
	debug  = getopt.BoolLong("debug", 'd', "Enable debug output")
	device = getopt.StringLong("device", 'f', "Specify device file (default: /dev/hidg0)")
)

func parseCommand(s string) (command string, arg float64, hasArg bool, err error) {
	command = ""
	arg = 0
	hasArg = true

	arr := strings.Fields(s)
	if len(arr) == 0 {
		return "", 0, false, nil
	}
	command = arr[0]
	if len(arr) > 1 {
		arg, err = strconv.ParseFloat(arr[1], 32)
		if err != nil {
			common.Warnf("Invalid float: %#v", arr[1])
			return "", 0, false, err
		}
		hasArg = true
	}
	common.Debugf("Command=%s arg=%d", command, arg)
	return
}

func mainLoop(con *nscon.Controller) (err error) {
	// Wait for stdin and convert to the command.
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command, arg, hasArg, err := parseCommand(scanner.Text())
		if err != nil || command == "" {
			continue
		}

		// Digital button arg: 0 or 1
		var darg uint8 = 0
		if !hasArg || arg != 0 {
			darg = 1
		}

		switch command {
		case "a": // A
			con.Input.Button.A = darg
		case "b": // B
			con.Input.Button.B = darg
		case "x": // X
			con.Input.Button.X = darg
		case "y": // Y
			con.Input.Button.Y = darg

		case "h": // Home
			con.Input.Button.Home = darg
		case "c": // Capture
			con.Input.Button.Capture = darg

		case "-", "m": // Minus
			con.Input.Button.Minus = darg
		case "+", "p": // plus
			con.Input.Button.Plus = darg

		case "l1": // L1
			con.Input.Button.L = darg
		case "l2": // L2
			con.Input.Button.ZL = darg
		case "r1": // R1
			con.Input.Button.R = darg
		case "r2": // R2
			con.Input.Button.ZR = darg

		case "u": // D-pad up
			con.Input.Dpad.Up = darg
		case "d": // D-pad down
			con.Input.Dpad.Down = darg
		case "l": // D-pad left
			con.Input.Dpad.Left = darg
		case "r": // D-pad right
			con.Input.Dpad.Right = darg

		case "lx": // Left stick X
			con.Input.Stick.Left.X = arg
		case "ly": // Left stick Y
			con.Input.Stick.Left.Y = arg

		case "rx": // Right stick X
			con.Input.Stick.Right.X = arg
		case "ry": // Right stick Y
			con.Input.Stick.Right.X = arg
		default:
			common.Warnf("Unknown command: %#v\n", command)
		}
	}
	return nil
}

func realMain() int {
	getopt.Parse()
	if device == nil {
		*device = "/dev/hidg0"
	}
	con := nscon.NewController(*device)
	if *debug {
		con.LogLevel = 2
		common.DebugEnabled = true
	}
	defer con.Close()

	// Connect to Switch
	common.Debugf("Opening %s...\n", *device)

	//err := con.Connect()
	//common.Checkf(err, "Unable to connect to device %s", *device)

	mainLoop(con)

	return 0
}

func main() {
	common.RunAndExit(realMain)
}
