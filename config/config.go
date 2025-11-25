package config

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"gopkg.in/yaml.v3"
)

// Config represents the entire configuration file.
type Config struct {
	Identifier string  `yaml:"identifier"`
	InfluxDB   Influx  `yaml:"influxdb"`
	ListenHTTP string  `yaml:"listenhttp"`
	Flows      []*Flow `yaml:"flows"`
}

// LoadConfig reads and parses a YAML config file.
func LoadConfig(path string) (*Config, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML format: %w", err)
	}

	// Validate entire config including flows, inputs, outputs
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	return &cfg, nil
}

// Validate validates the entire config.
func (c *Config) Validate() error {
	// Validate InfluxDB block
	if err := c.InfluxDB.Validate(); err != nil {
		return err
	}

	// Validate each flow
	for _, f := range c.Flows {
		if err := f.ValidateFlowConfig(); err != nil {
			return fmt.Errorf("flow %q validation failed: %w", f.Identifier, err)
		}
	}

	return nil
}

// --------------------------
// Flow Validation
// --------------------------

// ValidateFlowConfig keeps full existing functionality
// and validates inputs, outputs, and flow parameters.
func (f *Flow) ValidateFlowConfig() error {

	if f.Identifier == "" {
		return fmt.Errorf("flow missing identifier")
	}

	if len(f.Inputs) == 0 {
		return fmt.Errorf("flow %q has no inputs", f.Identifier)
	}

	// Validate inputs
	for _, in := range f.Inputs {
		if err := in.ValidateInput(); err != nil {
			return fmt.Errorf("invalid input in flow %q: %w", f.Identifier, err)
		}
	}

	// Validate outputs
	for _, out := range f.Outputs {
		if err := out.ValidateOutput(); err != nil {
			return fmt.Errorf("invalid output in flow %q: %w", f.Identifier, err)
		}
	}

	// Validate flow bitrate constraints (unchanged logic)
	if f.MinimalBitrate < 0 {
		return fmt.Errorf("flow %q has invalid minimalbitrate", f.Identifier)
	}
	if f.MaxPacketTimeMS < 0 {
		return fmt.Errorf("flow %q has invalid maxpackettime", f.Identifier)
	}

	return nil
}

// --------------------------
// Input Validation
// --------------------------

func (c *Input) ValidateInput() error {
	if c.Identifier == "" {
		return fmt.Errorf("input missing identifier")
	}
	if c.URL == "" {
		return fmt.Errorf("input %q is missing URL", c.Identifier)
	}
	_, err := url.Parse(c.URL)
	if err != nil {
		return fmt.Errorf("input %q has invalid URL: %w", c.Identifier, err)
	}
	return nil
}

// --------------------------
// Output Validation
// --------------------------

func (c *Output) ValidateOutput() error {
	if c.Identifier == "" {
		return fmt.Errorf("output missing identifier")
	}
	if c.URL == "" {
		return fmt.Errorf("output %q is missing URL", c.Identifier)
	}
	_, err := url.Parse(c.URL)
	if err != nil {
		return fmt.Errorf("output %q has invalid URL: %w", c.Identifier, err)
	}
	return nil
}
