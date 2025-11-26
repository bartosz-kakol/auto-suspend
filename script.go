package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func RunScript(interpreterPath string, scriptPath string) (string, error) {
	cmd := exec.Command(interpreterPath, scriptPath)

	output, err := cmd.Output()
	if err != nil {
		var exitError *exec.ExitError

		if errors.As(err, &exitError) {
			return "", fmt.Errorf(string(exitError.Stderr))
		}
	}

	strOutput := strings.TrimSpace(string(output))

	return strOutput, nil
}
