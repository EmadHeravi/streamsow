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
	// Start begins the input reader loop
	Start() error

	// Close terminates the input and releases resources
	Close()

	// Identifier returns the configured identifier for this input
	Identifier() string
}

// StartInput wraps input.Start() for a single input instance.
func StartInput(ctx context.Context, in input.Input) error {
	if in == nil {
		return nil
	}
	return in.Start()
}

// setupInput chooses and initializes the correct input based on URL scheme.
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
			return fmt.Errorf("could not setup rist input %q: %w", c.URL, err)
		}

	case "udp", "rtp":
		in, err = udp.SetupUDPInput(f.context, u, c.Identifier)
		if err != nil {
			return fmt.Errorf("could not setup udp input %q: %w", c.URL, err)
		}

	default:
		return fmt.Errorf("unsupported input scheme %q", u.Scheme)
	}

	f.configuredInputs[c.URL] = in
	return nil
}
