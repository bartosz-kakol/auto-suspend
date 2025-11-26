package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/fatih/color"
)

type DebuggerLogger struct {
	// cache commonly used styles
	primary     func(a ...any) string
	secondary   func(a ...any) string
	bold        func(a ...any) string
	boldRed     func(a ...any) string
	boldHiRed   func(a ...any) string
	boldGreen   func(a ...any) string
	boldYellow  func(a ...any) string
	boldMagenta func(a ...any) string
	red         func(a ...any) string
	lightGray   func(a ...any) string
}

func NewDebuggerLogger() *DebuggerLogger {
	return &DebuggerLogger{
		primary:     color.New(color.FgBlue, color.Bold).SprintFunc(),
		secondary:   color.New(color.FgCyan, color.Bold).SprintFunc(),
		bold:        color.New(color.Bold).SprintFunc(),
		boldRed:     color.New(color.FgRed, color.Bold).SprintFunc(),
		boldHiRed:   color.New(color.FgHiRed, color.Bold).SprintFunc(),
		boldGreen:   color.New(color.FgGreen, color.Bold).SprintFunc(),
		boldYellow:  color.New(color.FgYellow, color.Bold).SprintFunc(),
		boldMagenta: color.New(color.FgMagenta, color.Bold).SprintFunc(),
		red:         color.New(color.FgRed).SprintFunc(),
		lightGray:   color.New(color.FgWhite).SprintFunc(),
	}
}

func (l *DebuggerLogger) Begin() {
	// quiet
}

func (l *DebuggerLogger) RunningMasterScript(script string) {
	fmt.Printf("%s %s\n", l.primary("⚙️ running master script:"), script)
}

func (l *DebuggerLogger) MasterScriptOutput(output string) {
	fmt.Printf("%s\n\n", l.lightGray(output))
}

func (l *DebuggerLogger) NoMasterScriptDefaultPath() {
	fmt.Printf(
		"ℹ️ a %s was not specified. the '%s' path will be used (if it exists)\n\n",
		l.bold("master script"),
		l.bold("default"),
	)
}

func (l *DebuggerLogger) ChosenPath(path string) {
	fmt.Printf("%s %s\n", l.secondary("chosen path:"), path)
}

func (l *DebuggerLogger) OnErrorInfo(mode string) {
	format := "ℹ️ when a step in the sequence fails, it will %s by default.\n\n"

	switch mode {
	case "terminate":
		fmt.Printf(format, l.boldRed("terminate the sequence"))
	case "ignore":
		fmt.Printf(format, l.boldHiRed("NOT agree to suspend"))
	case "treat-as-yes":
		fmt.Printf(format, l.boldGreen("agree to suspend"))
	case "instant-suspend":
		fmt.Printf(format, l.boldYellow("ignore other steps and suspend immediately"))
	default:
		// Fallback: just print the mode
		fmt.Printf(format, l.bold(mode))
	}
}

func (l *DebuggerLogger) OnErrorInfoDefault() {
	fmt.Printf(
		"ℹ️ this sequence does not override the default %s behavior. when a step in the sequence fails, it will %s by default\n\n",
		l.bold("on_error"),
		l.boldRed("terminate the sequence"),
	)
}

func (l *DebuggerLogger) SequenceHeader() {
	fmt.Printf("%s\n", l.secondary("sequence:"))
}

func (l *DebuggerLogger) Operator(op string) {
	fmt.Printf("%s\n\n", l.boldMagenta(op))
}

func (l *DebuggerLogger) StepStart(scriptPath string) {
	fmt.Printf("▶︎ %s ", l.bold(scriptPath))
}

func (l *DebuggerLogger) StepInvalidScriptPath() {
	fmt.Printf("\n")
}

func (l *DebuggerLogger) StepErrorAction(action string, err error) {
	// action expected values: "terminated sequence", "did not agree to suspend", "agreed to suspend", "immediately suspend"
	// Choose color based on action
	var coloredAction string

	switch action {
	case "terminated sequence":
		coloredAction = l.boldRed(action)
	case "did not agree to suspend":
		coloredAction = l.boldHiRed(action)
	case "agreed to suspend":
		coloredAction = l.boldGreen(action)
	case "immediately suspend":
		coloredAction = l.boldYellow(action)
	default:
		coloredAction = l.bold(action)
	}

	fmt.Printf("| %s\n", coloredAction)

	if err != nil {
		fmt.Printf("%s\n", l.red(err.Error()))
	}
}

func (l *DebuggerLogger) StepErrorActionInvalid() {
	fmt.Printf("\n")
}

func (l *DebuggerLogger) StepOutputDecision(decision string, output string) {
	switch decision {
	case "agreed to suspend":
		fmt.Printf("| %s\n", l.boldGreen(decision))
	case "did not agree to suspend":
		fmt.Printf("| %s\n", l.boldHiRed(decision))
	default:
		fmt.Printf("| %s\n", l.bold(decision))
	}

	fmt.Printf("%s\n\n", l.lightGray(output))
}

func RunDebugger(cfg *Config, env *Environment) error {
	scanner := bufio.NewScanner(os.Stdin)

	logger := NewDebuggerLogger()

	inverted := color.New(color.FgBlack, color.BgHiWhite, color.Bold).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	primary := color.New(color.FgBlue, color.Bold).SprintFunc()
	boldGreen := color.New(color.FgGreen, color.Bold).SprintFunc()
	boldRed := color.New(color.FgRed, color.Bold).SprintFunc()

	// Clear the console
	// \033[H  -> Move cursor to top-left
	// \033[2J -> Clear the entire screen
	fmt.Print("\033[H\033[2J")

	fmt.Printf(
		"%s %s\n\n",
		bold("interpreter:"),
		env.PythonInterpreterPath,
	)

	for {
		suspend, err := AutoRunSequence(cfg, env, logger)

		if err != nil {
			err.Print()
		} else {
			fmt.Printf(
				"%s ",
				primary("The computer would"),
			)

			if suspend {
				fmt.Printf("%s", boldGreen("suspend"))
			} else {
				fmt.Printf("%s", boldRed("not suspend"))
			}

			fmt.Printf("\n")
		}

		fmt.Printf("%s", inverted("Press Enter/Return to repeat the sequence, or Ctrl+C to exit..."))
		scanner.Scan()
		fmt.Println("_______________________________________________________________")
	}
}
