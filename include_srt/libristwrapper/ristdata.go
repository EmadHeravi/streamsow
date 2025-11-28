package libristwrapper

/*
#cgo pkg-config: librist
#include <librist/librist.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

// RistDataBlock wraps the librist struct_rist_data_block
type RistDataBlock struct {
	Data []byte
	ptr  *C.struct_rist_data_block
}

// NewRistDataBlock allocates and fills a RIST data block for sending
func NewRistDataBlock(data []byte) *RistDataBlock {
	if len(data) == 0 {
		return nil
	}

	rb := &RistDataBlock{Data: data}
	rb.ptr = (*C.struct_rist_data_block)(C.calloc(1, C.size_t(unsafe.Sizeof(*rb.ptr))))

	rb.ptr.data = (*C.uint8_t)(C.CBytes(data))
	rb.ptr.size = C.ulong(len(data))

	return rb
}

// Return frees the underlying C memory
func (r *RistDataBlock) Return() {
	if r.ptr != nil {
		if r.ptr.data != nil {
			C.free(unsafe.Pointer(r.ptr.data))
			r.ptr.data = nil
		}
		C.free(unsafe.Pointer(r.ptr))
		r.ptr = nil
	}
}

// Increment is a no-op for UDP/legacy compatibility
func (r *RistDataBlock) Increment() {}
