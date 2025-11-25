package flow

import (
	"fmt"
	"net/url"

	"github.com/odmedia/streamzeug/config"
	"github.com/odmedia/streamzeug/input"
	"github.com/odmedia/streamzeug/input/rist"
	"github.com/odmedia/streamzeug/input/udp"
)

// setupInput chooses and initializes the correct input based on URL scheme.
// For RIST inputs we use the existing rist.SetupRistInput helper.
// For UDP/RTP inputs we construct a udp.UdpInput; the reader goroutine
// is started later once the mainloop is available.
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

// startUDPInputs is called once the mainloop has been created.
// It finds all configured UDP/RTP inputs and starts their reader
// goroutines, wiring them into the mainloop's UDP channel.
func (f *Flow) startUDPInputs() error {
	if f.m == nil {
		return nil
	}

	udpChan := f.m.UDPChannel()

	for urlStr, in := range f.configuredInputs {
		udpInput, ok := in.(*udp.UdpInput)
		if !ok {
			continue
		}

		if err := udpInput.StartReader(udpChan); err != nil {
			return fmt.Errorf("failed to start udp input %s: %w", urlStr, err)
		}
	}

	return nil
}
