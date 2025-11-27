package mainloop

import "code.videolan.org/rist/ristgo/libristwrapper"

func NewUDPBlock(data []byte) *libristwrapper.RistDataBlock {
	rb := libristwrapper.AllocDataBlock()  // existing API
	rb.Data = append([]byte(nil), data...) // copy data
	return rb
}
