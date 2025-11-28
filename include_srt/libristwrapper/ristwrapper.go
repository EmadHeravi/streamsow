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
		// Modern librist unified logging interface
		C.rist_logging_set(nil, C.int(level))
	}
}

// SetOutputIP adds a peer for output using the new librist 0.2.x API
func (c *Context) SetOutputIP(ip string) {
	if c.ptr == nil {
		return
	}

	cip := C.CString(ip)
	defer C.free(unsafe.Pointer(cip))

	// Initialize default peer configuration
	var peerConf C.struct_rist_peer_config
	C.rist_peer_config_defaults_set(&peerConf)

	// Parse the IP/port into the config
	if C.rist_parse_address(cip, &peerConf) != 0 {
		return
	}

	// Create and add the peer
	var peer *C.struct_rist_peer
	if C.rist_peer_create(&peer, c.ptr, &peerConf) == 0 {
		C.rist_peer_insert(c.ptr, peer)
	}
}
