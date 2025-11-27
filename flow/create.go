/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021 ODMedia
 * SPDX-FileContributor: Author: Gijs Peskens
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package flow

import (
	"context"
	"fmt"
	"sync"

	"code.videolan.org/rist/ristgo/libristwrapper"
	"github.com/odmedia/streamzeug/config"
	"github.com/odmedia/streamzeug/input"
	"github.com/odmedia/streamzeug/input/rist"
	"github.com/odmedia/streamzeug/logging"
	"github.com/odmedia/streamzeug/mainloop"
	"github.com/odmedia/streamzeug/stats"
)

// CreateFlow initializes and configures a Flow instance.
func CreateFlow(ctx context.Context, c *config.Flow) (*Flow, error) {
	var (
		flow Flow
		err  error
	)

	// base context and identifiers
	flow.rcontext = ctx
	flow.identifier = c.Identifier
	flow.outputWait = new(sync.WaitGroup)
	flow.config = *c

	// validate configuration
	if err := config.ValidateFlowConfig(c); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	logging.Log.Info().
		Str("identifier", c.Identifier).
		Msg("setting up flow")

	flow.context, flow.cancel = context.WithCancel(ctx)

	// statistics setup
	flow.statsConfig, err = stats.SetupStats(c.StatsStdOut, c.Identifier, c.StatsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to setup stats: %w", err)
	}

	// default latency
	if c.Latency == 0 {
		logging.Log.Info().
			Str("identifier", c.Identifier).
			Msg("latency not set, using default 1000ms")
		c.Latency = 1000
	}

	// cast RistProfile (config is int → expected libristwrapper enum)
	profile := libristwrapper.RistProfile(c.RistProfile)

	// setup rist receiver
	flow.receiver, err = rist.SetupReceiver(
		flow.context,
		c.Identifier,
		profile,
		c.Latency,
		flow.statsConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to setup rist receiver: %w", err)
	}

	// INPUTS
	flow.configuredInputs = make(map[string]input.Input)
	for _, in := range c.Inputs {
		if err := flow.setupInput(&in); err != nil {
			return nil, fmt.Errorf("failed to setup input %s: %w", in, err)
		}
	}

	// RIST destination port
	destinationPort := uint16(0)
	if profile != libristwrapper.RistProfileSimple {
		destinationPort = uint16(c.StreamID)
	}

	// start RIST receiver
	if err := flow.receiver.Start(); err != nil {
		return nil, fmt.Errorf("failed to start rist receiver: %w", err)
	}

	// configure flow
	rf, err := flow.receiver.ConfigureFlow(destinationPort)
	if err != nil {
		return nil, fmt.Errorf("failed to configure rist flow: %w", err)
	}

	// create mainloop
	flow.m = mainloop.NewMainloop(flow.context, rf, c.Identifier)

	// start UDP inputs (only now that mainloop / channels exist)
	if err := flow.startUDPInputs(); err != nil {
		return nil, fmt.Errorf("failed to start udp inputs: %w", err)
	}

	// OUTPUTS
	flow.configuredOutputs = make(map[string]outhandle)
	for _, out := range c.Outputs {
		if err := flow.setupOutput(&out); err != nil {
			return nil, fmt.Errorf("failed to setup output %s: %w", out, err)
		}
	}

	return &flow, nil
}
