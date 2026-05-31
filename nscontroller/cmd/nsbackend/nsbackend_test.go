package main

// This file contains backend parser unit tests verifying buttons, sticks, timings, and shortcuts.

import (
	"testing"
	"time"

	"github.com/omakoto/raspberry-switch-control/nscontroller"
)

func newTestCoordinator() (*Coordinator, *nscontroller.Controller) {
	// Initialize with empty device file path and dummy tick interval
	con := nscontroller.NewController("", 0)
	co := &Coordinator{
		con:     con,
		ch:      make(chan string, 100),
		started: true,
	}
	// Default autoReleaseDur for parsing command timings
	autoReleaseDur = 50 * time.Millisecond
	return co, con
}

func TestSendToController_Buttons(t *testing.T) {
	tests := []struct {
		cmd      string
		validate func(*nscontroller.Controller) bool
		desc     string
	}{
		{"a 1", func(c *nscontroller.Controller) bool { return c.Input.Button.A == 1 }, "Press A"},
		{"a 0", func(c *nscontroller.Controller) bool { return c.Input.Button.A == 0 }, "Release A"},
		{"b 1", func(c *nscontroller.Controller) bool { return c.Input.Button.B == 1 }, "Press B"},
		{"x 1", func(c *nscontroller.Controller) bool { return c.Input.Button.X == 1 }, "Press X"},
		{"y 1", func(c *nscontroller.Controller) bool { return c.Input.Button.Y == 1 }, "Press Y"},
		{"h 1", func(c *nscontroller.Controller) bool { return c.Input.Button.Home == 1 }, "Press Home"},
		{"c 1", func(c *nscontroller.Controller) bool { return c.Input.Button.Capture == 1 }, "Press Capture"},
		{"- 1", func(c *nscontroller.Controller) bool { return c.Input.Button.Minus == 1 }, "Press Minus (-)"},
		{"m 1", func(c *nscontroller.Controller) bool { return c.Input.Button.Minus == 1 }, "Press Minus (m)"},
		{"+ 1", func(c *nscontroller.Controller) bool { return c.Input.Button.Plus == 1 }, "Press Plus (+)"},
		{"p 1", func(c *nscontroller.Controller) bool { return c.Input.Button.Plus == 1 }, "Press Plus (p)"},
		{"l1 1", func(c *nscontroller.Controller) bool { return c.Input.Button.L == 1 }, "Press L1"},
		{"l2 1", func(c *nscontroller.Controller) bool { return c.Input.Button.ZL == 1 }, "Press L2"},
		{"r1 1", func(c *nscontroller.Controller) bool { return c.Input.Button.R == 1 }, "Press R1"},
		{"r2 1", func(c *nscontroller.Controller) bool { return c.Input.Button.ZR == 1 }, "Press R2"},
	}

	for _, tt := range tests {
		co, con := newTestCoordinator()
		co.sendToController(tt.cmd)
		if !tt.validate(con) {
			t.Errorf("%s failed: state not set correctly for cmd %q", tt.desc, tt.cmd)
		}
	}
}

func TestSendToController_Dpad(t *testing.T) {
	tests := []struct {
		cmd      string
		validate func(*nscontroller.Controller) bool
		desc     string
	}{
		{"pu 1", func(c *nscontroller.Controller) bool { return c.Input.Dpad.Up == 1 }, "D-pad Up"},
		{"pd 1", func(c *nscontroller.Controller) bool { return c.Input.Dpad.Down == 1 }, "D-pad Down"},
		{"pl 1", func(c *nscontroller.Controller) bool { return c.Input.Dpad.Left == 1 }, "D-pad Left"},
		{"pr 1", func(c *nscontroller.Controller) bool { return c.Input.Dpad.Right == 1 }, "D-pad Right"},
		{"pur 1", func(c *nscontroller.Controller) bool { return c.Input.Dpad.Up == 1 && c.Input.Dpad.Right == 1 }, "D-pad Up-Right"},
		{"pul 1", func(c *nscontroller.Controller) bool { return c.Input.Dpad.Up == 1 && c.Input.Dpad.Left == 1 }, "D-pad Up-Left"},
		{"pdr 1", func(c *nscontroller.Controller) bool { return c.Input.Dpad.Down == 1 && c.Input.Dpad.Right == 1 }, "D-pad Down-Right"},
		{"pdl 1", func(c *nscontroller.Controller) bool { return c.Input.Dpad.Down == 1 && c.Input.Dpad.Left == 1 }, "D-pad Down-Left"},
		{"px -1.0", func(c *nscontroller.Controller) bool { return c.Input.Dpad.Left == 1 && c.Input.Dpad.Right == 0 }, "D-pad Alternative Left"},
		{"px 1.0", func(c *nscontroller.Controller) bool { return c.Input.Dpad.Left == 0 && c.Input.Dpad.Right == 1 }, "D-pad Alternative Right"},
		{"py -1.0", func(c *nscontroller.Controller) bool { return c.Input.Dpad.Up == 1 && c.Input.Dpad.Down == 0 }, "D-pad Alternative Up"},
		{"py 1.0", func(c *nscontroller.Controller) bool { return c.Input.Dpad.Up == 0 && c.Input.Dpad.Down == 1 }, "D-pad Alternative Down"},
	}

	for _, tt := range tests {
		co, con := newTestCoordinator()
		co.sendToController(tt.cmd)
		if !tt.validate(con) {
			t.Errorf("%s failed for cmd %q", tt.desc, tt.cmd)
		}
	}
}

func TestSendToController_StickAnalogAndPress(t *testing.T) {
	co, con := newTestCoordinator()

	// Stick presses
	co.sendToController("lp 1")
	if con.Input.Stick.Left.Press != 1 {
		t.Error("Left stick press failed")
	}
	co.sendToController("rp 1")
	if con.Input.Stick.Right.Press != 1 {
		t.Error("Right stick press failed")
	}

	// Left stick coordinates (analog)
	co.sendToController("lx 0.5")
	if con.Input.Stick.Left.X != 0.5 {
		t.Errorf("Left stick X mapping failed, got %f", con.Input.Stick.Left.X)
	}

	co.sendToController("ly 0.75")
	// Left.Y should equal -arg (-0.75) due to Cartesian orientation inversion
	if con.Input.Stick.Left.Y != -0.75 {
		t.Errorf("Left stick Y mapping failed, got %f", con.Input.Stick.Left.Y)
	}

	// Right stick coordinates (analog)
	co.sendToController("rx -0.25")
	if con.Input.Stick.Right.X != -0.25 {
		t.Errorf("Right stick X mapping failed, got %f", con.Input.Stick.Right.X)
	}

	co.sendToController("ry -0.75")
	// Right.Y should equal -arg (0.75)
	if con.Input.Stick.Right.Y != 0.75 {
		t.Errorf("Right stick Y mapping failed, got %f", con.Input.Stick.Right.Y)
	}
}

func TestSendToController_StickShortcuts(t *testing.T) {
	tests := []struct {
		cmd      string
		validate func(*nscontroller.Controller) bool
		desc     string
	}{
		// Left Stick Shortcuts (Up is positive Y internally, Down is negative Y)
		{"lu 1", func(c *nscontroller.Controller) bool { return c.Input.Stick.Left.X == 0 && c.Input.Stick.Left.Y == 1.0 }, "Left Stick Up"},
		{"ld 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Left.X == 0 && c.Input.Stick.Left.Y == -1.0
		}, "Left Stick Down"},
		{"ll 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Left.X == -1.0 && c.Input.Stick.Left.Y == 0
		}, "Left Stick Left"},
		{"lr 1", func(c *nscontroller.Controller) bool { return c.Input.Stick.Left.X == 1.0 && c.Input.Stick.Left.Y == 0 }, "Left Stick Right"},
		{"lur 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Left.X == 1.0 && c.Input.Stick.Left.Y == 1.0
		}, "Left Stick Up-Right"},
		{"lul 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Left.X == -1.0 && c.Input.Stick.Left.Y == 1.0
		}, "Left Stick Up-Left"},
		{"ldr 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Left.X == 1.0 && c.Input.Stick.Left.Y == -1.0
		}, "Left Stick Down-Right"},
		{"ldl 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Left.X == -1.0 && c.Input.Stick.Left.Y == -1.0
		}, "Left Stick Down-Left"},

		// Right Stick Shortcuts (Up is positive Y internally, Down is negative Y)
		{"ru 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Right.X == 0 && c.Input.Stick.Right.Y == 1.0
		}, "Right Stick Up"},
		{"rd 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Right.X == 0 && c.Input.Stick.Right.Y == -1.0
		}, "Right Stick Down"},
		{"rl 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Right.X == -1.0 && c.Input.Stick.Right.Y == 0
		}, "Right Stick Left"},
		{"rr 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Right.X == 1.0 && c.Input.Stick.Right.Y == 0
		}, "Right Stick Right"},
		{"rur 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Right.X == 1.0 && c.Input.Stick.Right.Y == 1.0
		}, "Right Stick Up-Right"},
		{"rul 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Right.X == -1.0 && c.Input.Stick.Right.Y == 1.0
		}, "Right Stick Up-Left"},
		{"rdr 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Right.X == 1.0 && c.Input.Stick.Right.Y == -1.0
		}, "Right Stick Down-Right"},
		{"rdl 1", func(c *nscontroller.Controller) bool {
			return c.Input.Stick.Right.X == -1.0 && c.Input.Stick.Right.Y == -1.0
		}, "Right Stick Down-Left"},
	}

	for _, tt := range tests {
		co, con := newTestCoordinator()
		co.sendToController(tt.cmd)
		if !tt.validate(con) {
			t.Errorf("%s failed for shortcut cmd %q: got stick values X:%f, Y:%f", tt.desc, tt.cmd, con.Input.Stick.Left.X, con.Input.Stick.Left.Y)
		}
	}
}

func TestSendToController_AutoReleaseAndTimings(t *testing.T) {
	// Standard Auto-release trigger (no argument supplied)
	co, con := newTestCoordinator()
	co.sendToController("a")

	if con.Input.Button.A != 1 {
		t.Error("Auto-release press failed")
	}

	// Verify that a release event was queued on the channel
	select {
	case releaseCmd := <-co.ch:
		if releaseCmd != "a 0" {
			t.Errorf("Expected delayed release command 'a 0', got %q", releaseCmd)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timed out waiting for auto-release command on coordinator channel")
	}

	// Timed Auto-release command: "0.5 a"
	co, con = newTestCoordinator()
	co.sendToController("0.5 a")

	if con.Input.Button.A != 1 {
		t.Error("Timed auto-release press failed")
	}

	// With duration 0.5s and autoReleaseDur 50ms,
	// expected delayed command is (0.5s - 50ms = 0.45s) -> "0.450000 a 0"
	select {
	case releaseCmd := <-co.ch:
		expected := "0.450000 a 0"
		if releaseCmd != expected {
			t.Errorf("Expected delayed command %q, got %q", expected, releaseCmd)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timed out waiting for delayed command")
	}

	// Explicit Timed Command: "0.1 a 1"
	// This command executes an explicit value setting and goes to Sleep(dur) instead of SendDelayed.
	co, con = newTestCoordinator()
	start := time.Now()
	co.sendToController("0.1 a 1")
	elapsed := time.Since(start)

	if con.Input.Button.A != 1 {
		t.Error("Explicit timed press failed")
	}
	// Verify that it blocked / slept for approximately 100ms
	if elapsed < 80*time.Millisecond {
		t.Errorf("Expected explicit command to sleep for 100ms, slept for %v", elapsed)
	}

	// Verify no release is queued on the channel for explicit commands
	select {
	case releaseCmd := <-co.ch:
		t.Errorf("Unexpected release command received: %q", releaseCmd)
	default:
		// Correct behavior: nothing in queue
	}
}

func TestSendToController_InvalidAndParsing(t *testing.T) {
	co, con := newTestCoordinator()

	// Invalid command name
	co.sendToController("invalid_command 1")
	if con.Input.Button.A != 0 {
		t.Error("State changed on invalid command")
	}

	// Malformed inputs
	co.sendToController("a invalid_float")
	if con.Input.Button.A != 0 {
		t.Error("State changed on malformed float argument")
	}

	co.sendToController("0.5invalid_duration a")
	if con.Input.Button.A != 0 {
		t.Error("State changed on malformed duration argument")
	}

	// Boundary capping for analog values (limits are [-1..1])
	co.sendToController("lx 5.0")
	if con.Input.Stick.Left.X != 1.0 {
		t.Errorf("Upper bound capping failed, got %f", con.Input.Stick.Left.X)
	}

	co.sendToController("lx -2.5")
	if con.Input.Stick.Left.X != -1.0 {
		t.Errorf("Lower bound capping failed, got %f", con.Input.Stick.Left.X)
	}
}

func TestAToD(t *testing.T) {
	tests := []struct {
		val  float64
		neg  uint8
		pos  uint8
		desc string
	}{
		{1.0, 0, 1, "Positive Extreme"},
		{0.5, 0, 0, "Inside Positive Threshold"},
		{-1.0, 1, 0, "Negative Extreme"},
		{-0.5, 0, 0, "Inside Negative Threshold"},
		{0.0, 0, 0, "Center"},
	}

	for _, tt := range tests {
		neg, pos := aToD(tt.val)
		if neg != tt.neg || pos != tt.pos {
			t.Errorf("%s failed for %f: expected (neg:%d, pos:%d), got (neg:%d, pos:%d)", tt.desc, tt.val, tt.neg, tt.pos, neg, pos)
		}
	}
}
