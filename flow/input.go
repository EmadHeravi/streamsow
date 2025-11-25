package flow

import (
	"context"

	"github.com/odmedia/streamzeug/input"
)

type Input interface {
	// Start begins the input reader loop
	Start() error

	// Close terminates the input and releases resources
	Close()

	// Identifier returns the configured identifier for this input
	Identifier() string
}

func StartInput(ctx context.Context, in input.Input) error {
	if in == nil {
		return nil
	}
	return in.Start()
}
