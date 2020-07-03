package nscontroller

import (
	"fmt"
	"github.com/omakoto/go-common/src/common"
	"io"
)

type BackendConsumer struct {
	out io.WriteCloser
	ch  chan Event
}

var _ Consumer = (*BackendConsumer)(nil)
var _ Worker = (*BackendConsumer)(nil)

func NewBackendConsumer(out io.WriteCloser) (*BackendConsumer, error) {
	return &BackendConsumer{out, make(chan Event)}, nil
}

func (b *BackendConsumer) Close() error {
	close(b.ch)
	return b.out.Close()
}

func (b *BackendConsumer) Intake() <-chan Event {
	return b.ch
}

func (b *BackendConsumer) Run() {

	go func() {
		for {
			ev := <-b.ch

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

			msg := fmt.Sprint("%s %f\n", command, ev.Value)

			_, err := b.out.Write([]byte(msg))
			common.Checkf(err, "Unable to write the message: %w", err)
		}
	}()
}
