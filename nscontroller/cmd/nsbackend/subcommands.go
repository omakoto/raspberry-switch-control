package main

import (
	_ "embed"
	"fmt"
)

//go:embed "switch-controller-gadget"
var initScript string

func printUsbInitScript() {
	fmt.Printf("%s", initScript)
}
