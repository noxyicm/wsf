// +build windows

package utils

import (
	"os/exec"
	"syscall"
)

// IsolateProcess change gpid for the process to avoid bypassing signals to sub processes
func IsolateProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}
