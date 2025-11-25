/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021 ODMedia B.V.
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package config

import (
	"fmt"
	"net/url"
)

// InfluxDB configuration: only the URL is required.
type InfluxDB struct {
	URL string `yaml:"url"`
}

// Validate checks the InfluxDB configuration for correctness.
func (c *InfluxDB) Validate() error {
	if c == nil {
		return nil
	}
	if c.URL == "" {
		return fmt.Errorf("influxdb.url is required")
	}

	_, err := url.Parse(c.URL)
	if err != nil {
		return fmt.Errorf("invalid influxdb.url %q: %w", c.URL, err)
	}

	return nil
}
