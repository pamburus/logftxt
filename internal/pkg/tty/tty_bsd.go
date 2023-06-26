//go:build (darwin || freebsd || openbsd || netbsd || dragonfly) && !appengine
// +build darwin freebsd openbsd netbsd dragonfly
// +build !appengine

package tty

import (
	"syscall"
	"unsafe"
)

func enableSeqTTY(fd uintptr, _ bool) error {
	var termios syscall.Termios
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TIOCGETA, uintptr(unsafe.Pointer(&termios)), 0, 0, 0) //nolint:gosec
	if errno != 0 {
		return errno
	}

	return nil
}
