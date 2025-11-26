package main

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

type DaemonLogger struct {
	// cache commonly used styles
	primary   func(a ...any) string
	secondary func(a ...any) string
	bold      func(a ...any) string
	red       func(a ...any) string
	yellow    func(a ...any) string
	lightGray func(a ...any) string
}

func NewDaemonLogger() *DaemonLogger {
	return &DaemonLogger{
		primary:   color.New(color.FgBlue, color.Bold).SprintFunc(),
		secondary: color.New(color.FgCyan, color.Bold).SprintFunc(),
		bold:      color.New(color.Bold).SprintFunc(),
		red:       color.New(color.FgRed).SprintFunc(),
		yellow:    color.New(color.FgYellow).SprintFunc(),
		lightGray: color.New(color.FgWhite).SprintFunc(),
	}
}

func (l *DaemonLogger) log(message string) {
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	fmt.Printf(
		"%s %s\n",
		l.bold(fmt.Sprintf("[%s]", formattedTime)),
		message,
	)
}

func (l *DaemonLogger) Begin(env *Environment) {
	l.log(l.primary("auto-suspend daemon started"))
	l.log(fmt.Sprintf("interpreter: %s", env.PythonInterpreterPath))
}

func (l *DaemonLogger) RunningMasterScript(script string) {
	l.log(fmt.Sprintf("running master script: %s", script))
}

func (l *DaemonLogger) MasterScriptOutput(output string) {
	l.log(fmt.Sprintf("master script output:\n%s", l.lightGray(output)))
}

func (l *DaemonLogger) NoMasterScriptDefaultPath() {
	l.log(l.yellow("no master script was specified. the default path will be used."))
}

func (l *DaemonLogger) ChosenPath(path string) {
	l.log(fmt.Sprintf("chosen path: %s", path))
}

func (l *DaemonLogger) OnErrorInfo(mode string) {
	l.log(fmt.Sprintf("on_error mode: %s", mode))
}

func (l *DaemonLogger) OnErrorInfoDefault() {
	l.log("on_error mode: terminate (default)")
}

func (l *DaemonLogger) SequenceHeader() {
	l.log(l.secondary("running sequence"))
}

func (l *DaemonLogger) Operator(op string) {
	// quiet
}

func (l *DaemonLogger) StepStart(scriptPath string) {
	l.log(fmt.Sprintf("> %s", scriptPath))
}

func (l *DaemonLogger) StepInvalidScriptPath() {
	l.log(l.red("invalid script path"))
}

func (l *DaemonLogger) StepErrorAction(action string, err error) {
	l.log(action)
	l.log(fmt.Sprintf("script error:\n%s", l.red(err.Error())))
}

func (l *DaemonLogger) StepErrorActionInvalid() {
	l.log(l.red("invalid on_error value"))
}

func (l *DaemonLogger) StepOutputDecision(decision string, output string) {
	l.log(fmt.Sprintf("%s\nscript output:\n%s", decision, l.lightGray(output)))
}

func RunDaemon(cfg *Config, env *Environment) error {
	logger := NewDaemonLogger()
	sleepDuration := time.Duration(cfg.RunEvery) * time.Second
	logger.Begin(env)

	for {
		time.Sleep(sleepDuration)

		suspend, err := AutoRunSequence(cfg, env, logger)

		if err != nil {
			err.Print()
			panic("critical error. check logs.")
		} else {
			if suspend {
				// TODO actually suspend
			}
		}
	}
}
