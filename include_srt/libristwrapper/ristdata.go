package libristwrapper

/*
#cgo pkg-config: librist
#include <librist/librist.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

// RistDataBlock wraps the C struct rist_data_block used by librist.
type RistDataBlock struct {
	Data []byte
	Ptr  *C.struct_rist_data_block // Exported so other packages can access
}

// NewRistDataBlock allocates and fills a RIST data block for sending.
func NewRistDataBlock(data []byte) *RistDataBlock {
	if len(data) == 0 {
		return nil
	}

	rb := &RistDataBlock{Data: data}
	rb.Ptr = (*C.struct_rist_data_block)(C.calloc(1, C.size_t(unsafe.Sizeof(*rb.Ptr))))

	rb.Ptr.payload = (*C.uint8_t)(C.CBytes(data))
	rb.Ptr.payload_len = C.size_t(len(data))

	return rb
}

// Return frees the underlying C memory.
func (r *RistDataBlock) Return() {
	if r.Ptr != nil {
		if r.Ptr.payload != nil {
			C.free(unsafe.Pointer(r.Ptr.payload))
			r.Ptr.payload = nil
		}
		C.free(unsafe.Pointer(r.Ptr))
		r.Ptr = nil
	}
}

// Increment is a no-op for UDP/legacy compatibility.
func (r *RistDataBlock) Increment() {}
