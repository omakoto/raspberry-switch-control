package nscontroller

import (
	"bufio"
	"github.com/omakoto/go-common/src/common"
	"io"
	"regexp"
	"strings"
	"time"
)

const streamInputOffDelay = time.Millisecond * 200

type StreamInput struct {
	in   io.ReadCloser
	next Consumer
}

var _ Worker = (*StreamInput)(nil)

func NewStreamInput(in io.ReadCloser, next Consumer) (*StreamInput, error) {
	return &StreamInput{
		in,
		next,
	}, nil
}

func (t *StreamInput) Close() error {
	return t.in.Close()
}

func (t *StreamInput) press(a Action) {
	now := time.Now()
	on := Event{
		Timestamp: now,
		Action:    a,
		Value:     1,
	}
	t.next(&on)
	go (func() {
		off := Event{
			Timestamp: now.Add(streamInputOffDelay),
			Action:    a,
			Value:     0,
		}
		select {
		case <-time.After(streamInputOffDelay):
			t.next(&off)
		}
	})()
}

func (t *StreamInput) Run() {
	comment_re := regexp.MustCompile(`#.*`)
	go func() {
		scanner := bufio.NewScanner(t.in)
		for scanner.Scan() {
			in := scanner.Text()
			command := strings.TrimSpace(comment_re.ReplaceAllString(in, ""))
			switch command {
			case "a": // A
				t.press(ActionButtonA)
			case "b": // B
				t.press(ActionButtonB)
			case "x": // X
				t.press(ActionButtonX)
			case "y": // Y
				t.press(ActionButtonY)

			case "h": // Home
				t.press(ActionButtonHome)
			case "c": // Capture
				t.press(ActionButtonCapture)

			case "m": // Minus
				t.press(ActionButtonMinus)
			case "p": // plus
				t.press(ActionButtonPlus)

			case "l1": // L1
				t.press(ActionButtonL)
			case "l2": // L2
				t.press(ActionButtonLZ)
			case "r1": // R1
				t.press(ActionButtonR)
			case "r2": // R2
				t.press(ActionButtonRZ)

			case "pu": // D-pad up
				t.press(ActionButtonDpadUp)
			case "pd": // D-pad down
				t.press(ActionButtonDpadDown)
			case "pl": // D-pad left
				t.press(ActionButtonDpadLeft)
			case "pr": // D-pad right
				t.press(ActionButtonDpadRight)

			case "pur": // D-pad
				t.press(ActionButtonDpadUp)
				t.press(ActionButtonDpadRight)
			case "pul": // D-pad
				t.press(ActionButtonDpadUp)
				t.press(ActionButtonDpadLeft)
			case "pdr": // D-pad
				t.press(ActionButtonDpadDown)
				t.press(ActionButtonDpadRight)
			case "pdl": // D-pad
				t.press(ActionButtonDpadDown)
				t.press(ActionButtonDpadLeft)

			case "lp": // Left stick press
				t.press(ActionButtonLeftStickPress)
			case "rp": // Right stick press
				t.press(ActionButtonRightStickPress)

			default:
				common.Warnf("Unknown command: %#v\n", command)
				continue
			}
		}
	}()
}
