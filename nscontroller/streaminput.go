package nscontroller

// This file implements the StreamInput reader which parses simple text commands from stdin.

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/omakoto/go-common/src/common"
)

const streamInputOffDelay = time.Millisecond * 60

var commentRe = regexp.MustCompile(`#.*`)

type StreamInput struct {
	in   io.ReadCloser
	next Consumer
	wg   *sync.WaitGroup
}

var _ Worker = (*StreamInput)(nil)

func NewStreamInput(in io.ReadCloser, next Consumer) (*StreamInput, error) {
	return &StreamInput{
		in,
		next,
		&sync.WaitGroup{},
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

// Run starts reading from the input stream.
// NOTE: This is Working As Intended (WAI) to only accept simple, one-token directives
// from stdin (e.g. single-letter button presses like "a", "b", or d-pad directions like "pu").
// It does not parse multi-token commands, parameter values, or timing prefixes.
func (t *StreamInput) Run() {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		scanner := bufio.NewScanner(t.in)
		for scanner.Scan() {
			t.ProcessLine(scanner.Text())
		}
	}()
}

// ProcessLine parses a single line, strips comments/whitespace, and routes any valid command.
// Returns true if the command is valid or empty, and false if it is unrecognized.
func (t *StreamInput) ProcessLine(line string) bool {
	command := strings.TrimSpace(commentRe.ReplaceAllString(line, ""))
	if command == "" {
		return true
	}
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

	case "m", "-": // Minus
		t.press(ActionButtonMinus)
	case "p", "+": // Plus
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
		return false
	}
	return true
}

func (t *StreamInput) WaitClose() {
	t.wg.Wait()
}
