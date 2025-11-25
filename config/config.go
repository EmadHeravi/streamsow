/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021 ODMedia B.V.
 * SPDX-FileContributor: Author: Gijs Peskens <gijs@peskens.net>
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Config is the root YAML structure.
type Config struct {
	Identifier string `yaml:"identifier"`
	InfluxDB   Influx `yaml:"influxdb"`
	ListenHTTP string `yaml:"listenhttp"`
	Flows      []Flow `yaml:"flows"`
}

// Validate runs sanity checks on the full config.
func (c *Config) Validate() error {
	if c.Identifier == "" {
		return errors.New("identifier must not be empty")
	}

	// Validate influxdb section
	if err := c.InfluxDB.Validate(); err != nil {
		return fmt.Errorf("invalid influxdb config: %w", err)
	}

	// Validate flows
	if len(c.Flows) == 0 {
		return errors.New("no flows defined")
	}

	if err := checkDuplicates(c.Flows); err != nil {
		return err
	}

	for _, f := range c.Flows {
		if err := f.ValidateFlowConfig(); err != nil {
			return fmt.Errorf("invalid flow %q: %w", f.Identifier, err)
		}
	}

	return nil
}

// checkDuplicates ensures no duplicate flow identifiers.
func checkDuplicates(flows []Flow) error {
	seen := make(map[string]bool)
	for _, f := range flows {
		if seen[f.Identifier] {
			return fmt.Errorf("duplicate flow identifier: %s", f.Identifier)
		}
		seen[f.Identifier] = true
	}
	return nil
}

// LoadFromFile loads and parses a YAML configuration file.
func LoadFromFile(filename string) (*Config, error) {
	yamlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	conf := Config{}
	if err := yaml.Unmarshal(yamlData, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
