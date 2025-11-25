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

// Flow-level input interface
type Input interface {
	Start() error
	Close()
	Identifier() string
}

// StartInput wraps input.Start()
func StartInput(ctx context.Context, in input.Input) error {
	if in == nil {
		return nil
	}
	return in.Start()
}

// setupInput chooses the correct input based on URL scheme
func (f *Flow) setupInput(c *config.Input) error {
	u, err := url.Parse(c.Url)
	if err != nil {
		return fmt.Errorf("invalid input URL: %w", err)
	}

	var in input.Input

	switch u.Scheme {
	case "rist":
		in, err = rist.SetupRistInput(u, c.Identifier, f.receiver)
		if err != nil {
			return fmt.Errorf("could not setup rist input: %w", err)
		}

	case "udp", "rtp":
		in, err = udp.SetupUDPInput(f.context, u, c.Identifier)
		if err != nil {
			return fmt.Errorf("could not setup udp input: %w", err)
		}

	default:
		return fmt.Errorf("unsupported input scheme: %s", u.Scheme)
	}

	f.configuredInputs[c.Url] = in
	return nil
}
