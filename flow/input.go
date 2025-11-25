/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021 ODMedia B.V.
 * SPDX-FileContributor: Author: Gijs Peskens <gijs@peskens.net>
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package flow

import (
	"context"
	"fmt"
	"net/url"

	"github.com/odmedia/streamzeug/config"
	"github.com/odmedia/streamzeug/input"
	"github.com/odmedia/streamzeug/input/rist"
	"github.com/odmedia/streamzeug/input/udp"
)

// FlowInput is the interface used by the Flow for all inputs.
type FlowInput interface {
	StartReader(ch chan<- input.Packet) error
	Close()
	Identifier() string
}

// StartInput starts the input reader loop for a given input.
func StartInput(ctx context.Context, in input.Input) error {
	if in == nil {
		return nil
	}
	return in.Start()
}

// setupInput selects and initializes the correct input type (RIST or UDP/RTP).
func (f *Flow) setupInput(c *config.Input) error {
	u, err := url.Parse(c.URL)
	if err != nil {
		return fmt.Errorf("invalid input URL %q: %w", c.URL, err)
	}

	var in input.Input

	switch u.Scheme {
	case "rist":
		in, err = rist.SetupRistInput(u, c.Identifier, f.receiver)
		if err != nil {
			return fmt.Errorf("failed to setup RIST input %q: %w", c.URL, err)
		}

	case "udp", "rtp":
		in, err = udp.NewUdpInput(f.context, u, c.Identifier)
		if err != nil {
			return fmt.Errorf("failed to setup UDP/RTP input %q: %w", c.URL, err)
		}

	default:
		return fmt.Errorf("unsupported input scheme %q", u.Scheme)
	}

	f.configuredInputs[c.URL] = in
	return nil
}
