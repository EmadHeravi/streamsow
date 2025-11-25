/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021 ODMedia B.V. All right reserved.
 * SPDX-FileContributor: Author: Gijs Peskens <gijs@peskens.net>
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package udp

import (
	"context"
	"net/url"

	"github.com/odmedia/streamzeug/input"
	"github.com/odmedia/streamzeug/logging"
)

type UdpInput struct {
	ctx        context.Context
	cancel     context.CancelFunc
	url        *url.URL
	identifier string
}

// NewUdpInput creates a basic UDP input instance.
// For now it only sets up structure; actual socket handling will be added later.
func NewUdpInput(parentCtx context.Context, u *url.URL, identifier string) (input.Input, error) {
	logger := logging.Log.With().
		Str("module", "udp-input").
		Str("identifier", identifier).
		Logger()

	ctx, cancel := context.WithCancel(parentCtx)

	logger.Info().Msgf("setting up UDP input: %s", u.String())

	return &UdpInput{
		ctx:        ctx,
		cancel:     cancel,
		url:        u,
		identifier: identifier,
	}, nil
}

// Close stops the UDP input. Actual socket cleanup will be added later.
func (i *UdpInput) Close() {
	if i == nil {
		return
	}
	i.cancel()
}
