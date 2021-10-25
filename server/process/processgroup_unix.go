//go:build linux || darwin
// +build linux darwin

package process

import (
	"syscall"
)

func newProcessGroup() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true,
	}
}
