package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type SuspendCommand struct {
	Binary string
	Args   []string
}

func (c *SuspendCommand) String() string {
	return fmt.Sprintf("%s %s", c.Binary, strings.Join(c.Args, " "))
}

type systemSuspendCommandsMap struct {
	Windows []*SuspendCommand
	Linux   []*SuspendCommand
	Darwin  []*SuspendCommand
}

var systemSuspendCommands = systemSuspendCommandsMap{
	Windows: []*SuspendCommand{
		{
			Binary: "rundll32.exe",
			Args:   []string{"powerprof.dll,SetSuspendState", "0,1,0"},
		},
	},
	Linux: []*SuspendCommand{
		{
			Binary: "systemctl",
			Args:   []string{"suspend"},
		},
		{
			Binary: "pmsuspend",
			Args:   []string{},
		},
	},
	Darwin: []*SuspendCommand{
		{
			Binary: "pmset",
			Args:   []string{"sleepnow"},
		},
	},
}

func GetSystemSuspendCommands() ([]*SuspendCommand, error) {
	switch runtime.GOOS {
	case "windows":
		return systemSuspendCommands.Windows, nil
	case "linux":
		return systemSuspendCommands.Linux, nil
	case "darwin":
		return systemSuspendCommands.Darwin, nil
	default:
		return nil, fmt.Errorf("suspend behavior not defined for operating system: %s", runtime.GOOS)
	}
}

func RunSuspendCommand(cmd *SuspendCommand) error {
	return exec.Command(cmd.Binary, cmd.Args...).Run()
}

func AutoSystemSuspend() *map[string]error {
	cmds, err := GetSystemSuspendCommands()

	if err != nil {
		return &map[string]error{
			"GetSystemSuspendCommands": err,
		}
	}

	errors := map[string]error{}

	for _, cmd := range cmds {
		err = RunSuspendCommand(cmd)

		if err == nil {
			errors[cmd.String()] = nil
		}

		errors[cmd.String()] = err
	}

	return &errors
}
