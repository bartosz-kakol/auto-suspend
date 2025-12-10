package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func readConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s:\n%w", filename, err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file:\n%w", err)
	}

	return &cfg, nil
}

func run(cCtx *cli.Context) error {
	configFile := cCtx.Args().First()

	if configFile == "" {
		return cli.Exit("config_file argument is required", 1)
	}

	cfg, err := readConfig(configFile)
	if err != nil {
		return cli.Exit(err, 1)
	}

	env := NewEnvironment()

	opts := &DaemonOptions{
		Once: cCtx.Bool("once"),
	}

	return RunDaemon(cfg, env, opts)
}

func debug(cCtx *cli.Context) error {
	configFile := cCtx.Args().First()

	if configFile == "" {
		return cli.Exit("config_file argument is required", 1)
	}

	cfg, err := readConfig(configFile)
	if err != nil {
		return cli.Exit(err, 1)
	}

	env := NewEnvironment()

	return RunDebugger(cfg, env)
}
func main() {
	app := &cli.App{
		Name:  "auto-suspend",
		Usage: "An app which helps automatically suspend the computer when certain circumstances are met.",
		Commands: []*cli.Command{
			{
				Name:      "run",
				Usage:     "Run auto-suspend daemon.",
				Action:    run,
				ArgsUsage: "config_file",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "once",
						Usage: "Run the sequence only once and exit.",
					},
				},
			},
			{
				Name:      "debug",
				Usage:     "Debug your configuration by seeing its effects in detail without actually suspending the computer.",
				Action:    debug,
				ArgsUsage: "config_file",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
