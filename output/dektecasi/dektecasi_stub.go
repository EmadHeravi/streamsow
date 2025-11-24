//go:build !dektec
// +build !dektec

package dektecasi

// This is a stub for DekTec ASI output.
// The real DekTec implementation requires proprietary SDK and hardware.
// Unless built with -tags=dektec, this stub is used.

import "errors"

type DekTecOutput struct{}

func NewDekTecOutput() (*DekTecOutput, error) {
	return nil, errors.New("DekTec ASI support not enabled (missing SDK)")
}
