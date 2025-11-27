/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021
 * SPDX-FileContributor: Author: Gijs Peskkens <gijs@peskens.net>
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
	Identifier string         `yaml:"identifier"`
	InfluxDB   InfluxDBConfig `yaml:"influxdb"`
	ListenHTTP string         `yaml:"listenhttp"`
	Flows      []Flow         `yaml:"flows"`
}

// ------------------------------------------------------------
// FULL InfluxDBConfig (required by stats/influxdb.go)
// ------------------------------------------------------------

type InfluxDBConfig struct {
	// InfluxDB server URL
	Url string `yaml:"url"`

	// API token
	Token string `yaml:"token"`

	// InfluxDB organization
	Org string `yaml:"org"`

	// InfluxDB bucket
	Bucket string `yaml:"bucket"`

	// Measurement overrides
	SrtMeasurement         string `yaml:"srtmeasurement"`
	RistRXMeasurement      string `yaml:"ristrxmeasurement"`
	RistTXMeasurement      string `yaml:"risttxmeasurement"`
	ApplicationMeasurement string `yaml:"applicationmeasurement"`
}

func (c *InfluxDBConfig) Validate() error {
	if c == nil {
		return nil
	}

	if c.Url == "" {
		return fmt.Errorf("influxdb.url is required")
	}

	_, err := url.Parse(c.Url)
	if err != nil {
		return fmt.Errorf("invalid influxdb.url %q: %w", c.Url, err)
	}

	if c.Token == "" {
		return fmt.Errorf("influxdb.token is required")
	}

	if c.Org == "" {
		return fmt.Errorf("influxdb.org is required")
	}

	if c.Bucket == "" {
		return fmt.Errorf("influxdb.bucket is required")
	}

	return nil
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

// ------------------------------------------------------------
// Flow configuration validation
// ------------------------------------------------------------

func (f *Flow) ValidateFlowConfig() error {
	if f.Identifier == "" {
		return errors.New("flow identifier missing")
	}

	// ----------------------------
	// Validate Inputs
	// ----------------------------
	inIDs := []string{}
	for _, in := range f.Inputs {
		if in.Identifier == "" {
			return fmt.Errorf("flow %s: input identifier missing", f.Identifier)
		}

		inIDs = append(inIDs, in.Identifier)

		u, err := url.Parse(in.URL)
		if err != nil {
			return fmt.Errorf("invalid input URL: %s", in.URL)
		}

		switch u.Scheme {
		case "rist", "udp", "rtp":
		default:
			return fmt.Errorf("unsupported input scheme: %s", u.Scheme)
		}
	}

	if err := checkDuplicates("input identifier", inIDs); err != nil {
		return err
	}

	// ----------------------------
	// Validate Outputs
	// ----------------------------
	outIDs := []string{}
	for _, out := range f.Outputs {
		if out.Identifier == "" {
			return fmt.Errorf("flow %s: output identifier missing", f.Identifier)
		}

		outIDs = append(outIDs, out.Identifier)

		u, err := url.Parse(out.URL)
		if err != nil {
			return fmt.Errorf("invalid output URL: %s", out.URL)
		}

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

	// validate influx
	if err := conf.InfluxDB.Validate(); err != nil {
		return nil, err
	}

	// validate flows
	for _, fl := range conf.Flows {
		if err := fl.ValidateFlowConfig(); err != nil {
			return nil, err
		}
	}

	return &conf, nil
}
