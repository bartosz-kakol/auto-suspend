package main

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Config The main configuration struct
type Config struct {
	RunEvery     int                   `yaml:"run_every"`
	MasterScript string                `yaml:"master_script"`
	Paths        map[string]PathConfig `yaml:"paths"`
}

// PathConfig The config for each item under 'paths'
type PathConfig struct {
	OnError  string         `yaml:"on_error"`
	Sequence []SequenceStep `yaml:"sequence"`
}

// ScriptItem The complex script item within the sequence
type ScriptItem struct {
	Script  string `yaml:"script"`
	OnError string `yaml:"on_error"`
}

// RemoteItem The remote item within the sequence. 'Remote' is the address of an
// auto-suspend instance running in remote mode (e.g. "192.168.0.13:8080").
type RemoteItem struct {
	Remote  string `yaml:"remote"`
	OnError string `yaml:"on_error"`
}

// SequenceStep Custom type to hold a ScriptItem, a RemoteItem or a logical operator ("OR" or "AND")
type SequenceStep struct {
	Type     string // "script", "remote" or "operator"
	Operator string
	Script   ScriptItem
	Remote   RemoteItem
}

// stepFields is the union of every key a mapping-style sequence step may carry.
// It is only used to figure out which kind of step is being described.
type stepFields struct {
	Script  string `yaml:"script"`
	Remote  string `yaml:"remote"`
	OnError string `yaml:"on_error"`
}

// UnmarshalYAML Custom Unmarshaler for SequenceStep to handle mixed types
func (si *SequenceStep) UnmarshalYAML(value *yaml.Node) error {
	// Check if the node is a map (a ScriptItem or a RemoteItem)
	if value.Kind == yaml.MappingNode {
		var fields stepFields

		if err := value.Decode(&fields); err != nil {
			return err
		}

		switch {
		case fields.Script != "" && fields.Remote != "":
			return fmt.Errorf("a sequence item cannot define both 'script' and 'remote'")
		case fields.Script != "":
			si.Type = "script"
			si.Script = ScriptItem{Script: fields.Script, OnError: fields.OnError}
		case fields.Remote != "":
			si.Type = "remote"
			si.Remote = RemoteItem{Remote: fields.Remote, OnError: fields.OnError}
		default:
			return fmt.Errorf("a sequence item must define either 'script' or 'remote'")
		}

		return nil
	}

	// Otherwise, treat it as a scalar (an operator)
	if value.Kind == yaml.ScalarNode {
		si.Type = "operator"

		return value.Decode(&si.Operator)
	}

	return fmt.Errorf("sequence item must be a map (ScriptItem or RemoteItem) or a scalar (operator), got kind %v", value.Kind)
}

// OnError returns the step-level on_error override, or "" when the step does not set one.
func (si *SequenceStep) OnError() string {
	switch si.Type {
	case "script":
		return si.Script.OnError
	case "remote":
		return si.Remote.OnError
	}

	return ""
}

// Describe returns a human readable description of the step's target, used in error messages.
func (si *SequenceStep) Describe() string {
	switch si.Type {
	case "script":
		return fmt.Sprintf("the script at:\n%s", si.Script.Script)
	case "remote":
		return fmt.Sprintf("the remote at:\n%s", si.Remote.Remote)
	}

	return fmt.Sprintf("the '%s' step", si.Type)
}
