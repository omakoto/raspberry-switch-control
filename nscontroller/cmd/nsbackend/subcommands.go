package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/omakoto/go-common/src/common"
)

func usbInitScriptPath() string {
	thisFile, _ := common.GetSourceInfo()
	return filepath.Clean(filepath.Dir(thisFile) + "/../../../scripts/switch-controller-gadget")
}

func printUsbInitScriptPath() {
	fmt.Printf("%s\n", usbInitScriptPath())
}

func printUsbInitScript() {
	script, err := os.Open(usbInitScriptPath())
	common.Checke(err)

	content, err := io.ReadAll(script)
	common.Checke(err)

	fmt.Printf("%s", content)
}
