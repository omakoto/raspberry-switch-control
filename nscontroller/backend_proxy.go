package nscontroller

// This file implements the backend proxy which maps internal events to backend command strings and writes them to output.

import (
	"fmt"
	"io"
	"math"

	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/raspberry-switch-control/nscontroller/utils"
)

type BackendProxy struct {
	syncer *utils.Synchronized
	out    io.WriteCloser
}

var _ io.Closer = (*BackendProxy)(nil)

func NewBackendConsumer(out io.WriteCloser) (*BackendProxy, error) {
	return &BackendProxy{utils.NewSynchronized(), out}, nil
}

func (b *BackendProxy) Close() error {
	return b.out.Close()
}

func (b *BackendProxy) Consume(ev *Event) {
	b.syncer.Run(func() {
		command := ""

		switch ev.Action {
		case ActionButtonA:
			command = "a"
		case ActionButtonB:
			command = "b"
		case ActionButtonX:
			command = "x"
		case ActionButtonY:
			command = "y"

		case ActionButtonMinus:
			command = "-"
		case ActionButtonPlus:
			command = "+"

		case ActionButtonHome:
			command = "h"
		case ActionButtonCapture:
			command = "c"

		case ActionButtonDpadUp:
			command = "pu"
		case ActionButtonDpadDown:
			command = "pd"
		case ActionButtonDpadLeft:
			command = "pl"
		case ActionButtonDpadRight:
			command = "pr"

		case ActionButtonL:
			command = "l1"
		case ActionButtonR:
			command = "r1"
		case ActionButtonLZ:
			command = "l2"
		case ActionButtonRZ:
			command = "r2"

		case ActionButtonLeftStickPress:
			command = "lp"
		case ActionButtonRightStickPress:
			command = "rp"

		case ActionAxisLX:
			command = "lx"
		case ActionAxisLY:
			command = "ly"

		case ActionAxisRX:
			command = "rx"
		case ActionAxisRY:
			command = "ry"
		}

		val := ev.Value
		if ev.Action.isAxis() {
			// Round stick analog values to 4 digits below the 1's place, normalizing negative zero to zero.
			val = math.Round(val*10000) / 10000
			if val == 0 {
				val = 0
			}
		}

		msg := fmt.Sprint(command, " ", val, "\n")

		_, err := b.out.Write([]byte(msg))
		common.Checkf(err, "Unable to write the message")
	})
}
