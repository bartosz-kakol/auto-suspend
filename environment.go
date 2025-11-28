package main

import (
	"log"
	"os"
	"os/exec"
)

type Environment struct {
	PythonInterpreterPath string
	SuspendCommand        string
}

func NewEnvironment() *Environment {
	return &Environment{
		PythonInterpreterPath: getPythonInterpreterPath(),
		SuspendCommand:        getSuspendCommand(),
	}
}

func getPythonInterpreterPath() string {
	envPath, envExists := os.LookupEnv("AUTOSUSPEND_PYTHON_INTERPRETER_PATH")

	if envExists && envPath != "" {
		return envPath
	}

	// Try finding "python3" in the PATH
	path, err := exec.LookPath("python3")

	if err != nil {
		// Fallback to "python" (common on Windows)
		path, err = exec.LookPath("python")
	}

	if err != nil {
		log.Fatalf("failed to find python interpreter")
	}

	return path
}

func getSuspendCommand() string {
	return os.Getenv("AUTOSUSPEND_SUSPEND_COMMAND")
}
