package memongohelper

import (
	"os/exec"
	"runtime"
)

func RunCommand(cmds ...string) *exec.Cmd {
	var cmdObj *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		commandParms := append([]string{"/c"}, cmds...)
		cmdObj = exec.Command("cmd", commandParms...)

	default:
		commandParms := append([]string{"-c"}, cmds...)
		cmdObj = exec.Command("/bin/sh", commandParms...)
	}

	return cmdObj
}
