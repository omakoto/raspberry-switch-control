package main

import (
	"bufio"
	"github.com/mzyy94/nscon"
	"github.com/omakoto/go-common/src/common"
	"github.com/pborman/getopt/v2"
	"math"
	"os"
	"strconv"
	"strings"
)

// TODO Introduce constants for commands.

const (
	analogToDigitalThreshold = 1.0
)

var (
	debug  = getopt.BoolLong("debug", 'd', "Enable debug output")
	device = getopt.StringLong("device", 'f', "/dev/hidg0", "Specify device file")
)

func parseCommand(s string) (command string, arg float64, hasArg bool, err error) {
	command = ""
	arg = 0
	hasArg = false

	arr := strings.Fields(s)
	if len(arr) == 0 {
		return "", 0, false, nil
	}
	command = strings.ToLower(arr[0])
	if len(arr) > 1 {
		arg, err = strconv.ParseFloat(arr[1], 32)
		if err != nil {
			common.Warnf("Invalid float: %#v", arr[1])
			return "", 0, false, err
		}
		if arg < -1 {
			arg = -1
		} else if arg > 1 {
			arg = 1
		}
		hasArg = true
	}
	common.Debugf("Command=%#v arg=%f", command, arg)
	return
}

func aToD(arg float64, neg, pos *uint8) {
	if arg >= analogToDigitalThreshold {
		*pos = 1
		*neg = 0
	} else if arg <= -analogToDigitalThreshold {
		*pos = 0
		*neg = 1
	} else {
		*pos = 0
		*neg = 0
	}
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
		if !hasArg || math.Abs(arg) >= analogToDigitalThreshold {
			darg = 1
		}
		fdarg := float64(darg)

		// Hmm, analog stick Y is inverted?

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

		case "l", "l1": // L1
			con.Input.Button.L = darg
		case "l2", "lz": // L2
			con.Input.Button.ZL = darg
		case "r", "r1": // R1
			con.Input.Button.R = darg
		case "r2", "rz": // R2
			con.Input.Button.ZR = darg

		case "pu": // D-pad up
			con.Input.Dpad.Up = darg
		case "pd": // D-pad down
			con.Input.Dpad.Down = darg
		case "pl": // D-pad left
			con.Input.Dpad.Left = darg
		case "pr": // D-pad right
			con.Input.Dpad.Right = darg

		case "pur", "pru": // D-pad
			con.Input.Dpad.Up = darg
			con.Input.Dpad.Right = darg
		case "pul", "plu": // D-pad
			con.Input.Dpad.Up = darg
			con.Input.Dpad.Left = darg
		case "pdr", "prd": // D-pad
			con.Input.Dpad.Down = darg
			con.Input.Dpad.Right = darg
		case "pdl", "pld": // D-pad
			con.Input.Dpad.Down = darg
			con.Input.Dpad.Left = darg

		case "px": // D-pad alternative
			aToD(arg, &con.Input.Dpad.Right, &con.Input.Dpad.Right)
		case "py": // D-pad alternative
			aToD(arg, &con.Input.Dpad.Down, &con.Input.Dpad.Up)

		case "lp": // Left stick press
			con.Input.Stick.Left.Press = darg
		case "rp": // Right stick press
			con.Input.Stick.Right.Press = darg

		case "lx": // Left stick X
			con.Input.Stick.Left.X = arg
		case "ly": // Left stick Y
			con.Input.Stick.Left.Y = -arg

			// Left stick alternative
		case "lu":
			con.Input.Stick.Left.X = 0
			con.Input.Stick.Left.Y = -fdarg
		case "ld":
			con.Input.Stick.Left.X = 0
			con.Input.Stick.Left.Y = -fdarg
		case "ll":
			con.Input.Stick.Left.X = -fdarg
			con.Input.Stick.Left.Y = 0
		case "lr":
			con.Input.Stick.Left.X = fdarg
			con.Input.Stick.Left.Y = 0
		case "lur", "lru":
			con.Input.Stick.Left.X = fdarg
			con.Input.Stick.Left.Y = fdarg
		case "lul", "llu":
			con.Input.Stick.Left.X = -fdarg
			con.Input.Stick.Left.Y = fdarg
		case "ldr", "lrd":
			con.Input.Stick.Left.X = fdarg
			con.Input.Stick.Left.Y = -fdarg
		case "ldl", "lld":
			con.Input.Stick.Left.X = -fdarg
			con.Input.Stick.Left.Y = -fdarg

		case "rx": // Right stick X
			con.Input.Stick.Right.X = arg
		case "ry": // Right stick Y
			con.Input.Stick.Right.Y = -arg

			// Right stick alternative
		case "ru":
			con.Input.Stick.Right.X = 0
			con.Input.Stick.Right.Y = fdarg
		case "rd":
			con.Input.Stick.Right.X = 0
			con.Input.Stick.Right.Y = -fdarg
		case "rl":
			con.Input.Stick.Right.X = -fdarg
			con.Input.Stick.Right.Y = 0
		case "rr":
			con.Input.Stick.Right.X = fdarg
			con.Input.Stick.Right.Y = 0
		case "rur", "rru":
			con.Input.Stick.Right.X = fdarg
			con.Input.Stick.Right.Y = fdarg
		case "rul", "rlu":
			con.Input.Stick.Right.X = -fdarg
			con.Input.Stick.Right.Y = fdarg
		case "rdr", "rrd":
			con.Input.Stick.Right.X = fdarg
			con.Input.Stick.Right.Y = -fdarg
		case "rdl", "rld":
			con.Input.Stick.Right.X = -fdarg
			con.Input.Stick.Right.Y = -fdarg

		default:
			common.Warnf("Unknown command: %#v\n", command)
			continue
		}
		common.Dump("State:", con.Input)
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

	err := con.Connect()
	common.Checkf(err, "Unable to connect to device %s", *device)

	mainLoop(con)

	return 0
}

func main() {
	common.RunAndExit(realMain)
}
