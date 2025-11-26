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

// SequenceStep Custom type to hold either a ScriptItem or a logical operator ("OR" or "AND")
type SequenceStep struct {
	Type     string // "script" or "operator"
	Operator string
	Script   ScriptItem
}

// UnmarshalYAML Custom Unmarshaler for SequenceStep to handle mixed types
func (si *SequenceStep) UnmarshalYAML(value *yaml.Node) error {
	// Check if the node is a map (a ScriptItem)
	if value.Kind == yaml.MappingNode {
		si.Type = "script"

		return value.Decode(&si.Script)
	}

	// Otherwise, treat it as a scalar (an operator)
	if value.Kind == yaml.ScalarNode {
		si.Type = "operator"

		return value.Decode(&si.Operator)
	}

	return fmt.Errorf("sequence item must be a map (ScriptItem) or a scalar (operator), got kind %v", value.Kind)
}
