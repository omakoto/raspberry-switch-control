package main

import (
	"fmt"
	"path/filepath"

	"github.com/omakoto/go-common/src/common"
)

func printUsbInitScriptPath() {
	thisFile, _ := common.GetSourceInfo()
	script := filepath.Clean(filepath.Dir(thisFile) + "/../../../scripts/switch-controller-gadget")
	fmt.Printf("%s\n", script)
}
