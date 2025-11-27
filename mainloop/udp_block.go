package mainloop

import "code.videolan.org/rist/ristgo/libristwrapper"

// NewUDPBlock converts raw UDP bytes into a RIST-compatible data block
func NewUDPBlock(data []byte) *libristwrapper.RistDataBlock {
	rb := libristwrapper.NewRistDataBlock() // CORRECT constructor
	if rb == nil {
		return nil
	}
	// Make a private copy of the bytes
	rb.Data = append([]byte(nil), data...)
	return rb
}
