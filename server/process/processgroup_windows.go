//go:build windows
// +build windows

package process

import (
	"syscall"
)

func newProcessGroup() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
