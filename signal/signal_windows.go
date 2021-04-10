// +build windows

package signal

import "golang.org/x/sys/windows"

const (
	_sigINT  = windows.SIGINT
	_sigTERM = windows.SIGTERM
)
