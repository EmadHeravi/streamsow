package stats

import (
	"fmt"
	"net/url"
)

// ------------------------------------------------------------
// InfluxDB config (FULL STRUCT REQUIRED BY stats/influxdb.go)
// ------------------------------------------------------------

type InfluxDBConfig struct {
	// InfluxDB server URL
	Url string `yaml:"url"`

	// API token for authentication
	Token string `yaml:"token"`

	// Organization name in InfluxDB
	Org string `yaml:"org"`

	// Bucket where data will be written
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
