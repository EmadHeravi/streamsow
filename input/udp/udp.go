/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021 ODMedia B.V.
 * SPDX-FileContributor: Author: Gijs Peskens <gijs@peskens.net>
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package udp

import (
	"context"
	"net/url"

	"github.com/EmadHeravi/streamsow/input"
	"github.com/EmadHeravi/streamsow/logging"
)

// UdpInput implements input.Input and represents a UDP-based input source.
// Actual socket handling is implemented in reader.go.
type UdpInput struct {
	ctx        context.Context
	cancel     context.CancelFunc
	url        *url.URL
	identifier string
}

// NewUdpInput sets up a UDP input object.
// The packet reader loop will be started separately in reader.go.
func NewUdpInput(parentCtx context.Context, u *url.URL, identifier string) (input.Input, error) {
	logger := logging.Log.With().
		Str("module", "udp-input").
		Str("identifier", identifier).
		Str("url", u.String()).
		Logger()

	ctx, cancel := context.WithCancel(parentCtx)

	logger.Info().Msg("initializing UDP input")

	return &UdpInput{
		ctx:        ctx,
		cancel:     cancel,
		url:        u,
		identifier: identifier,
	}, nil
}

// Close stops the UDP input. The reader loop listens to the context.
func (i *UdpInput) Close() {
	if i == nil {
		return
	}
	i.cancel()
}
