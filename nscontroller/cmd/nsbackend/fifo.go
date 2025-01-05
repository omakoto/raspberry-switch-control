package main

import (
	"errors"
	"os"
	"syscall"

	"github.com/omakoto/go-common/src/common"
)

func mustCreateFifo(path string) *os.File {
	common.Debugf("Creating FIFO at '%s'", path)

	_, err := os.Stat(path)
	if err == nil {
		// File exists. Delete it.
		err = os.Remove(path)
		common.Checkf(err, "Cannot delete file: '%s'", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		common.Checkf(err, "Cannot create file '%s': stat failed", path)
	}
	err = syscall.Mkfifo(path, 0600)
	common.Checkf(err, "Makefifo failed for '%s'", path)

	file, err := os.OpenFile(path, os.O_RDWR, 0600)
	common.Checkf(err, "Open failed for '%s'", path)

	return file
}
