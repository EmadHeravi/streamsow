package libristwrapper

/*
#cgo pkg-config: librist
#include <librist/librist.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

type RistDataBlock struct {
	Data []byte
	ptr  *C.struct_rist_data_block
}

func NewRistDataBlock(data []byte) *RistDataBlock {
	rb := &RistDataBlock{Data: data}
	rb.ptr = (*C.struct_rist_data_block)(C.calloc(1, C.size_t(unsafe.Sizeof(*rb.ptr))))
	if len(data) > 0 {
		rb.ptr.payload = (*C.uchar)(C.CBytes(data))
		rb.ptr.payload_len = C.uint(len(data))
	}
	return rb
}

func (r *RistDataBlock) Return() {
	if r.ptr != nil {
		if r.ptr.payload != nil {
			C.free(unsafe.Pointer(r.ptr.payload))
			r.ptr.payload = nil
		}
		C.free(unsafe.Pointer(r.ptr))
		r.ptr = nil
	}
}

func (r *RistDataBlock) Increment() {
	// No-op for UDP/legacy use
}
