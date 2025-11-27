/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021 ODMedia B.V. All right reserved.
 * SPDX-FileContributor: Author: Gijs Peskens <gijs@peskens.net>
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package flow

import (
	"fmt"
	"net/url"

	"github.com/EmadHeravi/streamsow/config"
	"github.com/EmadHeravi/streamsow/output"
	"github.com/EmadHeravi/streamsow/output/dektecasi"
	"github.com/EmadHeravi/streamsow/output/srt"
	"github.com/EmadHeravi/streamsow/output/udp"
)

type outhandle struct {
	out  output.Output
	conf config.Output
}

func (f *Flow) setupOutput(c *config.Output) error {
	// Parse output URL
	outputURL, err := url.Parse(c.URL)
	if err != nil {
		return fmt.Errorf("couldn't parse output URL %s: %w", c.URL, err)
	}

	var out output.Output

	// Select correct output handler
	switch outputURL.Scheme {
	case "udp", "rtp":
		out, err = udp.ParseUdpOutput(f.context, outputURL, f.identifier, f.m)

	case "srt":
		out, err = srt.ParseSrtOutput(f.context, outputURL, f.identifier, c.Identifier, f.m, f.statsConfig, f.outputWait)

	case "dektecasi":
		out, err = dektecasi.ParseURL(f.context, outputURL, f.identifier, c.Identifier, f.m, f.statsConfig)

	default:
		return fmt.Errorf("output URL scheme not implemented: %s", outputURL.Scheme)
	}

	// If output initialization failed
	if err != nil {
		return fmt.Errorf("couldn't setup %s output (%s): %w",
			outputURL.Scheme, outputURL.String(), err)
	}

	// Store configured output
	f.configuredOutputs[c.URL] = outhandle{
		out:  out,
		conf: *c,
	}

	return nil
}
