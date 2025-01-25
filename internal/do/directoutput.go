package do

import "syscall"

var (
	directinput, _ = syscall.LoadLibrary("DirectOutput.dll")
)
