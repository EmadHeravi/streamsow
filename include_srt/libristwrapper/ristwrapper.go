package libristwrapper

/*
#cgo pkg-config: librist
#include <librist/librist.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

// Context wraps a RIST context (sender or receiver)
type Context struct {
	ptr *C.struct_rist_ctx
}

// SendData sends a RIST data block
func (c *Context) SendData(rb *RistDataBlock) int {
	if c.ptr == nil || rb == nil || rb.ptr == nil {
		return -1
	}
	return int(C.rist_sender_data_write(c.ptr, rb.ptr))
}

// Free releases the RIST context
func (c *Context) Free() {
	if c.ptr != nil {
		C.rist_destroy(c.ptr)
		c.ptr = nil
	}
}

// InitSender creates a new RIST sender context
func InitSender(profile int) *Context {
	var ctx *C.struct_rist_ctx
	C.rist_sender_create(&ctx, C.int(profile))
	return &Context{ptr: ctx}
}

// InitReceiver creates a new RIST receiver context
func InitReceiver(profile int) *Context {
	var ctx *C.struct_rist_ctx
	C.rist_receiver_create(&ctx, C.int(profile))
	return &Context{ptr: ctx}
}

// SetLogLevel sets the global librist logging level (0–7)
func (c *Context) SetLogLevel(level int) {
	if c.ptr != nil {
		// Modern unified logging interface
		C.rist_logging_set(nil, C.int(level))
	}
}

// SetOutputIP connects the sender to the given RIST/UDP URL
func (c *Context) SetOutputIP(ip string) {
	if c.ptr == nil {
		return
	}

	cip := C.CString(ip)
	defer C.free(unsafe.Pointer(cip))

	// New API: directly connect via URL string (rist:// or udp://)
	if C.rist_url_connect(c.ptr, cip) != 0 {
		// Connection failed – ignore or log as needed
	}
}
