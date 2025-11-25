/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021
 * SPDX-FileContributor: Author: Gijs Peskens <gijs@peskens.net>
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"

	"gopkg.in/yaml.v3"
)

// ------------------------------------------------------------
// Root configuration
// ------------------------------------------------------------

type Config struct {
	Identifier string `yaml:"identifier"`
	InfluxDB   Influx `yaml:"influxdb"`
	ListenHTTP string `yaml:"listenhttp"`
	Flows      []Flow `yaml:"flows"`
}

// ------------------------------------------------------------
// InfluxDB config
// ------------------------------------------------------------

type Influx struct {
	URL string `yaml:"url"`
}

// ------------------------------------------------------------
// Flow configuration
// ------------------------------------------------------------

type Flow struct {
	Identifier      string   `yaml:"identifier"`
	Type            string   `yaml:"type"`
	RistProfile     int      `yaml:"ristprofile"`
	Latency         int      `yaml:"latency"`
	StreamID        int      `yaml:"streamid"`
	Inputs          []Input  `yaml:"inputs"`
	Outputs         []Output `yaml:"outputs"`
	MinimalBitrate  int      `yaml:"minimalbitrate"`
	MaxPacketTimeMS int      `yaml:"maxpackettime"`
}

// ------------------------------------------------------------
// Input + Output structs
// ------------------------------------------------------------

type Input struct {
	Identifier string `yaml:"identifier"`
	URL        string `yaml:"url"`
}

type Output struct {
	Identifier string `yaml:"identifier"`
	URL        string `yaml:"url"`
}

// ------------------------------------------------------------
// Validation helpers
// ------------------------------------------------------------

func checkDuplicates(name string, items []string) error {
	seen := map[string]bool{}
	for _, v := range items {
		if seen[v] {
			return fmt.Errorf("duplicate %s: %s", name, v)
		}
		seen[v] = true
	}
	return nil
}

// ValidateFlowConfig – preserves original functionality
// now extended to allow UDP inputs.
func (f *Flow) ValidateFlowConfig() error {
	if f.Identifier == "" {
		return errors.New("flow identifier missing")
	}

	// check input identifiers
	inIDs := []string{}
	for _, i := range f.Inputs {
		if i.Identifier == "" {
			return fmt.Errorf("flow %s: input identifier missing", f.Identifier)
		}
		inIDs = append(inIDs, i.Identifier)

		// validate URL
		u, err := url.Parse(i.URL)
		if err != nil {
			return fmt.Errorf("invalid input URL: %s", i.URL)
		}

		switch u.Scheme {
		case "rist", "udp", "rtp":
			// valid
		default:
			return fmt.Errorf("unsupported input scheme: %s", u.Scheme)
		}
	}

	if err := checkDuplicates("input identifier", inIDs); err != nil {
		return err
	}

	// check outputs
	outIDs := []string{}
	for _, o := range f.Outputs {
		if o.Identifier == "" {
			return fmt.Errorf("flow %s: output identifier missing", f.Identifier)
		}
		outIDs = append(outIDs, o.Identifier)

		u, err := url.Parse(o.URL)
		if err != nil {
			return fmt.Errorf("invalid output URL: %s", o.URL)
		}

		// original behavior: only SRT supported
		if u.Scheme != "srt" {
			return fmt.Errorf("unsupported output scheme: %s", u.Scheme)
		}
	}

	return checkDuplicates("output identifier", outIDs)
}

// ------------------------------------------------------------
// Load YAML
// ------------------------------------------------------------

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
