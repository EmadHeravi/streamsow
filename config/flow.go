/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021 ODMedia B.V.
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package config

import (
	"errors"
	"fmt"
	"net/url"
)

func ValidateFlowConfig(c *Flow) error {

	if c.Identifier == "" {
		return errors.New("flow identifier missing")
	}
	if len(c.Inputs) == 0 {
		return fmt.Errorf("flow %s must have at least one input", c.Identifier)
	}
	if len(c.Outputs) == 0 {
		return fmt.Errorf("flow %s must have at least one output", c.Identifier)
	}

	// --------------------------------------
	// INPUT VALIDATION (RIST OR UDP/RTP)
	// --------------------------------------
	for _, in := range c.Inputs {
		u, err := url.Parse(in.URL)
		if err != nil {
			return fmt.Errorf("invalid input URL %s: %w", in.URL, err)
		}

		switch u.Scheme {
		case "rist":
		case "udp":
		case "rtp":
			// accepted
		default:
			return fmt.Errorf("input scheme %s not supported (rist, udp, rtp allowed)", u.Scheme)
		}
	}

	// --------------------------------------
	// OUTPUT VALIDATION
	// --------------------------------------
	for _, out := range c.Outputs {
		u, err := url.Parse(out.URL)
		if err != nil {
			return fmt.Errorf("invalid output URL %s: %w", out.URL, err)
		}

		switch u.Scheme {
		case "srt", "udp", "rtp", "dektecasi":
			// accepted
		default:
			return fmt.Errorf("output scheme %s not supported", u.Scheme)
		}
	}

	return nil
}
