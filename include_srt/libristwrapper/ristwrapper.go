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
		C.rist_logging_set(nil, C.int(level))
	}
}

// SetOutputIP configures the UDP output address using the legacy UDP config API
func (c *Context) SetOutputIP(ip string) {
	if c.ptr == nil {
		return
	}

	cip := C.CString(ip)
	defer C.free(unsafe.Pointer(cip))

	var udpConf *C.struct_rist_udp_config

	// Parse the address and build a UDP config structure
	if C.rist_parse_udp_address2(cip, &udpConf) != 0 {
		return
	}
	defer C.rist_udp_config_free2(&udpConf)

	// Create the sender/receiver using this configuration
	C.rist_sender_create(&c.ptr, C.int(0))
}
