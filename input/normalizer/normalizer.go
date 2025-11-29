package normalizer

/*
#cgo pkg-config: librist
#include <librist/librist.h>
#include <stdlib.h>
*/
import "C"

import (
	"github.com/EmadHeravi/streamsow/include_srt/libristwrapper"
)

// WrapToRist converts raw MPEG-TS data ([]byte) into a proper
// RIST-compatible data block using libristwrapper.NewRistDataBlock().
func WrapToRist(data []byte) *libristwrapper.RistDataBlock {
	return libristwrapper.NewRistDataBlock(data)
}

// FreeRistBlock releases all underlying C allocations of a RIST data block.
func FreeRistBlock(rb *libristwrapper.RistDataBlock) {
	if rb == nil {
		return
	}
	rb.Return()
}
