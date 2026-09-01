package main

// This file contains unit tests for nsfrontend dispatcher selection and flag parsing.

import (
	"testing"

	"github.com/omakoto/raspberry-switch-control/nscontroller/js"
)

func TestMustGetDispatcher(t *testing.T) {
	tests := []struct {
		name       string
		deviceName string
	}{
		{"XboxOne", "Microsoft X-Box One pad"},
		{"Xbox360", "Xbox 360 Wireless Receiver"},
		{"NintendoSwitchPro", "Nintendo Switch Pro Controller"},
		{"PS4", "Sony Interactive Entertainment Wireless Controller"},
		{"PS4_Old", "Sony Computer Entertainment Wireless Controller"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dummyJs := &js.Js{Name: tc.deviceName}
			dispatcher := mustGetDispatcher(dummyJs)
			if dispatcher == nil {
				t.Fatalf("expected non-nil dispatcher for device %q", tc.deviceName)
			}
		})
	}
}
