package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
)

type SequenceError struct {
	message string
	details string
}

func NewSequenceError(message string, details string) *SequenceError {
	return &SequenceError{message, details}
}

func (e *SequenceError) Print() {
	titleText := color.New(color.FgHiRed, color.Bold).SprintFunc()
	errorText := color.New(color.FgRed).SprintFunc()

	fmt.Printf(
		"🛑 %s\n%s\n",
		titleText(e.message),
		errorText(e.details),
	)
}

// ILogger defines logging hooks used by AutoRunSequence and RunSequence.
type ILogger interface {
	RunningMasterScript(script string)
	MasterScriptOutput(output string)
	NoMasterScriptDefaultPath()
	ChosenPath(path string)

	OnErrorInfo(mode string)
	OnErrorInfoDefault()

	SequenceHeader()
	Operator(op string)

	StepStart(scriptPath string)
	StepInvalidScriptPath()
	StepErrorAction(action string, err error)
	StepErrorActionInvalid()
	StepOutputDecision(decision string, output string) // `decision` is "yes" or "no"
}

func RunSequence(path *PathConfig, env *Environment, logger ILogger) (bool, *SequenceError) {
	if path.OnError != "" {
		switch path.OnError {
		case "terminate", "ignore", "treat-as-yes", "instant-suspend":
			logger.OnErrorInfo(path.OnError)
		default:
			return false, NewSequenceError(
				"invalid on_error value",
				fmt.Sprintf("'%s' is not a valid on_error value", path.OnError),
			)
		}
	} else {
		logger.OnErrorInfoDefault()
	}

	logger.SequenceHeader()

	andGroups := make([]bool, 0, len(path.Sequence))
	var currentOrGroup []bool
	lastOrIndex := -1

	for index, step := range path.Sequence {
		if step.Type == "operator" {
			switch step.Operator {
			case "OR":
				if index == 0 || index == len(path.Sequence)-1 {
					return false, NewSequenceError(
						"invalid operator",
						fmt.Sprintf("'%s' cannot be the first or last operator in a sequence", step.Operator),
					)
				}

				logger.Operator(step.Operator)
				lastOrIndex = index
			default:
				return false, NewSequenceError(
					"invalid operator",
					fmt.Sprintf("'%s' is not a valid operator", step.Operator),
				)
			}

			continue
		}

		logger.StepStart(step.Script.Script)

		if _, err := os.Stat(step.Script.Script); errors.Is(err, os.ErrNotExist) {
			logger.StepInvalidScriptPath()

			return false, NewSequenceError(
				"invalid script path",
				fmt.Sprintf("the script at:\n%s\ndoes not exist", step.Script.Script),
			)
		}

		output, err := RunScript(env.PythonInterpreterPath, step.Script.Script)
		var suspendDecision bool

		if err != nil {
			onErrorBehavior := step.Script.OnError

			if onErrorBehavior == "" {
				onErrorBehavior = path.OnError
			}

			if onErrorBehavior == "" {
				onErrorBehavior = "terminate"
			}

			switch onErrorBehavior {
			case "terminate":
				logger.StepErrorAction("terminated sequence", err)
				return false, nil
			case "ignore":
				logger.StepErrorAction("did not agree to suspend", err)
				suspendDecision = false
			case "treat-as-yes":
				logger.StepErrorAction("agreed to suspend", err)
				suspendDecision = true
			case "instant-suspend":
				logger.StepErrorAction("immediately suspend", err)
				return true, nil
			default:
				logger.StepErrorActionInvalid()
				return false, NewSequenceError(
					"invalid on_error value",
					fmt.Sprintf("'%s' is not a valid on_error value", onErrorBehavior),
				)
			}
		} else {
			switch output {
			case "yes":
				logger.StepOutputDecision("agreed to suspend", output)
				suspendDecision = true
			case "no":
				logger.StepOutputDecision("did not agree to suspend", output)
				suspendDecision = false
			default:
				return false, NewSequenceError(
					"invalid script output",
					fmt.Sprintf("the script at:\n%s\nreturned an invalid output: '%s'", step.Script.Script, output),
				)
			}
		}

		if lastOrIndex < index-1 {
			andGroups = append(andGroups, HasTrue(currentOrGroup))
			currentOrGroup = make([]bool, 0, len(path.Sequence)-index)
		}

		currentOrGroup = append(currentOrGroup, suspendDecision)
	}

	andGroups = append(andGroups, HasTrue(currentOrGroup))
	doSuspend := AllTrue(andGroups)

	return doSuspend, nil
}

func AutoRunSequence(cfg *Config, env *Environment, logger ILogger) (bool, *SequenceError) {
	if _, err := os.Stat(env.PythonInterpreterPath); errors.Is(err, os.ErrNotExist) {
		return false, NewSequenceError(
			"invalid interpreter path",
			fmt.Sprintf("the interpreter path at:\n%s\ndoes not exist.", env.PythonInterpreterPath),
		)
	}

	var pathName string

	if cfg.MasterScript != "" {
		logger.RunningMasterScript(cfg.MasterScript)

		output, err := RunScript(env.PythonInterpreterPath, cfg.MasterScript)
		if err != nil {
			return false, NewSequenceError(
				"error while running master script",
				err.Error(),
			)
		}

		logger.MasterScriptOutput(output)

		pathName = output
	} else {
		pathName = "default"
		logger.NoMasterScriptDefaultPath()
	}

	logger.ChosenPath(pathName)

	path, ok := cfg.Paths[pathName]
	if !ok {
		return false, NewSequenceError(
			"invalid path chosen by the master script",
			fmt.Sprintf("the master script output '%s' is not a valid path", pathName),
		)
	}

	return RunSequence(&path, env, logger)
}
