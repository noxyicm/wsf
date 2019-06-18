// +build !windows

package utils

import (
	"os/exec"
	"syscall"
)

// IsolateProcess change gpid for the process to avoid bypassing signals to sub processes
func IsolateProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}
}
