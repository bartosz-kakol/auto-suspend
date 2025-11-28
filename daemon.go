package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

type DaemonLogger struct {
	// cache commonly used styles
	Primary   func(a ...any) string
	Secondary func(a ...any) string
	Bold      func(a ...any) string
	Red       func(a ...any) string
	Green     func(a ...any) string
	Yellow    func(a ...any) string
	LightGray func(a ...any) string

	noMasterScriptWarningShown bool
}

func NewDaemonLogger() *DaemonLogger {
	return &DaemonLogger{
		Primary:   color.New(color.FgBlue, color.Bold).SprintFunc(),
		Secondary: color.New(color.FgCyan, color.Bold).SprintFunc(),
		Bold:      color.New(color.Bold).SprintFunc(),
		Red:       color.New(color.FgRed).SprintFunc(),
		Green:     color.New(color.FgGreen).SprintFunc(),
		Yellow:    color.New(color.FgYellow).SprintFunc(),
		LightGray: color.New(color.FgWhite).SprintFunc(),

		noMasterScriptWarningShown: false,
	}
}

func (l *DaemonLogger) Log(message string) {
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	fmt.Printf(
		"%s %s\n",
		l.Bold(fmt.Sprintf("[%s]", formattedTime)),
		message,
	)
}

func (l *DaemonLogger) Begin(env *Environment) {
	l.Log(l.Primary("auto-suspend daemon started"))
	l.Log(fmt.Sprintf("interpreter: %s", env.PythonInterpreterPath))
}

func (l *DaemonLogger) RunningMasterScript(script string) {
	l.Log(fmt.Sprintf("running master script: %s", script))
}

func (l *DaemonLogger) MasterScriptOutput(output string) {
	l.Log(fmt.Sprintf("master script output:\n%s", l.LightGray(output)))
}

func (l *DaemonLogger) NoMasterScriptDefaultPath() {
	if l.noMasterScriptWarningShown {
		return
	}

	l.Log(l.Yellow("no master script was specified. the default path will be used."))

	l.noMasterScriptWarningShown = true
}

func (l *DaemonLogger) ChosenPath(path string) {
	l.Log(fmt.Sprintf("chosen path: %s", path))
}

func (l *DaemonLogger) OnErrorInfo(mode string) {
	l.Log(fmt.Sprintf("on_error mode: %s", mode))
}

func (l *DaemonLogger) OnErrorInfoDefault() {
	l.Log("on_error mode: terminate (default)")
}

func (l *DaemonLogger) SequenceHeader() {
	l.Log(l.Secondary("running sequence"))
}

func (l *DaemonLogger) Operator(op string) {
	// quiet
}

func (l *DaemonLogger) StepStart(scriptPath string) {
	l.Log(fmt.Sprintf("> %s", scriptPath))
}

func (l *DaemonLogger) StepInvalidScriptPath() {
	l.Log(l.Red("invalid script path"))
}

func (l *DaemonLogger) StepErrorAction(action string, err error) {
	l.Log(action)
	l.Log(fmt.Sprintf("script error:\n%s", l.Red(err.Error())))
}

func (l *DaemonLogger) StepErrorActionInvalid() {
	l.Log(l.Red("invalid on_error value"))
}

func (l *DaemonLogger) StepOutputDecision(decision string, output string) {
	l.Log(fmt.Sprintf("%s\nscript output:\n%s", decision, l.LightGray(output)))
}

type DaemonOptions struct {
	Once bool
}

func RunDaemon(cfg *Config, env *Environment, opts *DaemonOptions) error {
	logger := NewDaemonLogger()
	sleepDuration := time.Duration(cfg.RunEvery) * time.Second
	logger.Begin(env)

	for {
		if !opts.Once {
			time.Sleep(sleepDuration)
		}

		suspend, err := AutoRunSequence(cfg, env, logger)

		if err != nil {
			err.Print()
			panic("critical error. check logs.")
		} else {
			if suspend {
				suspendErrors := AutoSystemSuspend()

				var sb strings.Builder
				sb.WriteString(
					fmt.Sprintf(
						"💤 %s using the following methods:\n",
						logger.Primary("suspending"),
					),
				)

				for title, err := range *suspendErrors {
					var summary string
					var details string

					if err != nil {
						summary = "❌ did not work"
						details = fmt.Sprintf("%s", logger.Red(err.Error()))
					} else {
						summary = "✅ worked"
						details = logger.Green("success")
					}

					sb.WriteString(
						fmt.Sprintf(
							"> %s: %s\n%s\n\n",
							logger.LightGray(title),
							summary,
							details,
						),
					)
				}

				logger.Log(sb.String())
			}
		}

		if opts.Once {
			break
		}
	}

	return nil
}
