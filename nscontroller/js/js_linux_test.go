//go:build linux

package js

// This file contains unit tests for the Linux joystick event reading and state management.

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

type readCloser struct {
	io.Reader
	closed bool
}

func (r *readCloser) Close() error {
	r.closed = true
	return nil
}

func TestJs_ReadEvents(t *testing.T) {
	events := []osJsEvent{
		{
			Time:      1000,
			Value:     16383,
			EventType: jsEventAxis,
			Number:    0,
		},
		{
			Time:      2000,
			Value:     1,
			EventType: jsEventButton,
			Number:    1,
		},
		{
			Time:      3000,
			Value:     0,
			EventType: jsEventButton,
			Number:    1,
		},
	}

	buf := new(bytes.Buffer)
	for _, ev := range events {
		err := binary.Write(buf, binary.LittleEndian, &ev)
		if err != nil {
			t.Fatalf("failed to write mock event: %v", err)
		}
	}

	rc := &readCloser{Reader: buf}
	js := &Js{
		DevicePath: "/dev/input/js0",
		Name:       "Test Joystick",
		NumAxes:    1,
		NumButtons: 2,
		Axes: []Element{
			{Number: 0, Name: "x"},
		},
		Buttons: []Element{
			{Number: 0, Name: "a"},
			{Number: 1, Name: "b"},
		},
		in: rc,
	}

	// 1. Read axis event
	ev, err := js.Read()
	if err != nil {
		t.Fatalf("unexpected error reading axis event: %v", err)
	}
	if ev.Element == nil || ev.Element.Name != "x" {
		t.Errorf("expected axis element x, got %v", ev.Element)
	}
	if ev.Value < 0.49 || ev.Value > 0.51 {
		t.Errorf("expected value around 0.5, got %f", ev.Value)
	}

	// 2. Read button pressed event
	ev, err = js.Read()
	if err != nil {
		t.Fatalf("unexpected error reading button event: %v", err)
	}
	if ev.Element == nil || ev.Element.Name != "b" {
		t.Errorf("expected button element b, got %v", ev.Element)
	}
	if ev.Value != 1.0 {
		t.Errorf("expected button value 1.0, got %f", ev.Value)
	}

	// 3. Read button released event
	ev, err = js.Read()
	if err != nil {
		t.Fatalf("unexpected error reading button event: %v", err)
	}
	if ev.Element == nil || ev.Element.Name != "b" {
		t.Errorf("expected button element b, got %v", ev.Element)
	}
	if ev.Value != 0.0 {
		t.Errorf("expected button value 0.0, got %f", ev.Value)
	}

	// 4. EOF
	_, err = js.Read()
	if err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}

	// 5. Close
	if err := js.Close(); err != nil {
		t.Errorf("unexpected error closing js: %v", err)
	}
	if !rc.closed {
		t.Errorf("expected underlying reader to be closed")
	}
}

func TestJs_ReadOutOfBounds(t *testing.T) {
	// Event with axis index that is out of bounds
	event := osJsEvent{
		Time:      1000,
		Value:     1000,
		EventType: jsEventAxis,
		Number:    10, // out of range
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, &event); err != nil {
		t.Fatalf("failed to write mock event: %v", err)
	}

	js := &Js{
		DevicePath: "/dev/input/js0",
		NumAxes:    0,
		NumButtons: 0,
		Axes:       nil,
		Buttons:    nil,
		in:         &readCloser{Reader: buf},
	}

	ev, err := js.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Element != nil {
		t.Errorf("expected nil Element for out of bounds axis, got %v", ev.Element)
	}
}

func TestJs_SetInitialValues(t *testing.T) {
	js := &Js{
		Axes: []Element{
			{Number: 0, Name: "x", Value: 0.5},
		},
		Buttons: []Element{
			{Number: 0, Name: "a", Value: 1.0},
		},
	}
	js.setInitialValues()
	if js.Axes[0].InitialValue != 0.5 {
		t.Errorf("expected axis initial value 0.5, got %f", js.Axes[0].InitialValue)
	}
	if js.Buttons[0].InitialValue != 1.0 {
		t.Errorf("expected button initial value 1.0, got %f", js.Buttons[0].InitialValue)
	}
}
