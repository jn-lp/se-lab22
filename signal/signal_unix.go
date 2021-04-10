// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package signal

import "golang.org/x/sys/unix"

const (
	_sigINT  = unix.SIGINT
	_sigTERM = unix.SIGTERM
)
