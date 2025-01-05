package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/raspberry-switch-control/nscontroller"
	"github.com/omakoto/raspberry-switch-control/nscontroller/daemon"
	"github.com/pborman/getopt/v2"
)

const (
	analogToDigitalThreshold = 1.0
)

var (
	help          = getopt.BoolLong("help", 'h', "help")
	debug         = getopt.BoolLong("debug", 'd', "Enable debug output")
	device        = getopt.StringLong("device", 'f', "/dev/hidg0", "Specify device file")
	startAsDaemon = getopt.BoolLong("daemon", 'x', "Run as daemon (implies --make-fifo)")
	createFifo    = getopt.BoolLong("make-fifo", 0, "Create a FIFO and read commands from it")
	fifo          = getopt.StringLong("fifo", 0, "/tmp/nsbackend.fifo", "Specify FIFO filename")

	autoReleaseMillis = getopt.IntLong("auto-release-millis", 'a', 50, "Set auto-release delay in milliseconds")
)

// The delay needs to be bigger than the interval within startInputReport().
const AUTO_RELEASE_MILLIS_MIN = 50

func parseCommand(s string) (command string, arg float64, autoRelease bool, err error) {
	command = ""
	arg = 0

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
		if arg < -1 {
			arg = -1
		} else if arg > 1 {
			arg = 1
		}
	} else {
		arg = 1
		autoRelease = true
	}

	ar := ""
	if autoRelease {
		ar = " (auto-release)"
	}
	common.Debugf("Command=%#v arg=%f%s", command, arg, ar)
	return
}

func aToD(arg float64) (neg, pos uint8) {
	if arg >= analogToDigitalThreshold {
		pos = 1
		neg = 0
	} else if arg <= -analogToDigitalThreshold {
		pos = 0
		neg = 1
	} else {
		pos = 0
		neg = 0
	}
	return
}

type Coordinator struct {
	con     *nscontroller.Controller
	ch      chan string
	wg      sync.WaitGroup
	started bool
}

func NewCoordinator(con *nscontroller.Controller) Coordinator {
	return Coordinator{
		con: con,
		ch:  make(chan string, 100),
	}
}

func (co *Coordinator) checkStarted() {
	if !co.started {
		panic("Not started")
	}
}

func (co *Coordinator) Send(command string) {
	co.checkStarted()
	if command == "" {
		return
	}
	co.ch <- command
}

func (co *Coordinator) SendDelayed(command string, waitMillis int) {
	co.checkStarted()
	go func() {
		time.Sleep(time.Duration(waitMillis) * time.Millisecond)
		co.Send(command)
	}()
}

func (co *Coordinator) Close() {
	co.ch <- ""
}

func (co *Coordinator) Start() {
	if co.started {
		panic("Already started")
	}
	co.started = true
	co.wg.Add(1)
	go func() {
		defer co.wg.Done()
		for command := range co.ch {
			if command == "" {
				break
			}
			co.sendToController(command)
		}
	}()
}

func (co *Coordinator) Wait() {
	co.checkStarted()
	co.wg.Wait()
}

func (co *Coordinator) sendToController(command string) {
	command, arg, autoRelease, err := parseCommand(command)
	if err != nil || command == "" {
		return
	}

	// Digital button arg: 0 or 1
	var darg uint8 = 0
	if math.Abs(arg) >= analogToDigitalThreshold {
		darg = 1
	}
	fdarg := float64(darg)

	// Hmm, analog stick Y is inverted?
	con := co.con

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

	case "m", "-": // Minus
		con.Input.Button.Minus = darg
	case "p", "+": // Plus
		con.Input.Button.Plus = darg

	case "l1": // L1
		con.Input.Button.L = darg
	case "l2": // L2
		con.Input.Button.ZL = darg
	case "r1": // R1
		con.Input.Button.R = darg
	case "r2": // R2
		con.Input.Button.ZR = darg

	case "pu": // D-pad up
		con.Input.Dpad.Up = darg
	case "pd": // D-pad down
		con.Input.Dpad.Down = darg
	case "pl": // D-pad left
		con.Input.Dpad.Left = darg
	case "pr": // D-pad right
		con.Input.Dpad.Right = darg

	case "pur": // D-pad
		con.Input.Dpad.Up = darg
		con.Input.Dpad.Right = darg
	case "pul": // D-pad
		con.Input.Dpad.Up = darg
		con.Input.Dpad.Left = darg
	case "pdr": // D-pad
		con.Input.Dpad.Down = darg
		con.Input.Dpad.Right = darg
	case "pdl": // D-pad
		con.Input.Dpad.Down = darg
		con.Input.Dpad.Left = darg

	case "px": // D-pad alternative
		con.Input.Dpad.Left, con.Input.Dpad.Right = aToD(arg)
	case "py": // D-pad alternative
		con.Input.Dpad.Up, con.Input.Dpad.Down = aToD(arg)

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
	case "lur":
		con.Input.Stick.Left.X = fdarg
		con.Input.Stick.Left.Y = fdarg
	case "lul":
		con.Input.Stick.Left.X = -fdarg
		con.Input.Stick.Left.Y = fdarg
	case "ldr":
		con.Input.Stick.Left.X = fdarg
		con.Input.Stick.Left.Y = -fdarg
	case "ldl":
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
	case "rur":
		con.Input.Stick.Right.X = fdarg
		con.Input.Stick.Right.Y = fdarg
	case "rul":
		con.Input.Stick.Right.X = -fdarg
		con.Input.Stick.Right.Y = fdarg
	case "rdr":
		con.Input.Stick.Right.X = fdarg
		con.Input.Stick.Right.Y = -fdarg
	case "rdl":
		con.Input.Stick.Right.X = -fdarg
		con.Input.Stick.Right.Y = -fdarg

	default:
		common.Warnf("Unknown command: %#v\n", command)
		return
	}

	con.Send()
	con.Dump()

	if autoRelease {
		co.SendDelayed(command+" 0", *autoReleaseMillis)
	}
}

func mainLoop(con *nscontroller.Controller, input *os.File) error {
	co := NewCoordinator(con)
	scanner := bufio.NewScanner(input)

	co.Start()

	fmt.Printf("nsbackend: Waiting for input... (^D to exit)\n")
	for scanner.Scan() {
		input := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if input == "q" {
			break
		}
		co.Send(input)
	}
	fmt.Printf("nsbackend: exiting...\n")

	co.Close()
	co.Wait()

	return scanner.Err()
}

func maybeHandleSubcommand() int {
	if len(os.Args) > 1 {
		subcommand := os.Args[1]
		if !strings.HasPrefix(subcommand, "-") {
			switch subcommand {
			case "usb-init-script-path":
				printUsbInitScriptPath()
				return 0
			case "show-usb-init-script":
				printUsbInitScript()
				return 0
			default:
				common.Fatalf("Unknown subcommand: %s", subcommand)
			}
		}
	}
	return -1
}

func realMain() int {
	syscall.Umask(0)

	if ret := maybeHandleSubcommand(); ret >= 0 {
		return ret
	}

	getopt.Parse()
	if *help {
		getopt.Usage()
		return 0
	}

	if *debug {
		// con.LogLevel = 2
		common.DebugEnabled = true
		common.VerboseEnabled = true
	}

	if *startAsDaemon {
		if daemon.Start() {
			// parent
			return 0
		}
		// In daemon mode, always use FIFO
		*createFifo = true
	}

	if *autoReleaseMillis < AUTO_RELEASE_MILLIS_MIN {
		*autoReleaseMillis = AUTO_RELEASE_MILLIS_MIN
	}
	if device == nil {
		*device = "/dev/hidg0"
	}

	input := os.Stdin

	if *createFifo {
		fmt.Printf("Creating FIFO at %s...\n", input.Name())
		input = mustCreateFifo(*fifo)
		fmt.Printf("To stop it, run: echo q > '%s'\n", input.Name())
		fmt.Printf("Reading input from '%s'...\n", input.Name())
	}

	con := nscontroller.NewController(*device)
	defer con.Close()

	// Connect to Switch
	common.Debugf("Opening %s...\n", *device)

	err := con.Connect()
	common.Checkf(err, "Unable to connect to device %s", *device)

	err = mainLoop(con, input)
	common.Check(err, "Failed to read from input")

	return 0
}

func main() {
	common.RunAndExit(realMain)
}
