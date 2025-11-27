package libristwrapper

/*
#cgo pkg-config: librist
#include <librist/librist.h>
*/
import "C"
import "unsafe"

type Context struct {
	ptr *C.struct_rist_ctx
}

func (c *Context) SendData(rb *RistDataBlock) int {
	if c.ptr == nil || rb == nil || rb.ptr == nil {
		return -1
	}
	return int(C.rist_sender_data_write(c.ptr, rb.ptr))
}

func (c *Context) Free() {
	if c.ptr != nil {
		C.rist_destroy(c.ptr)
		c.ptr = nil
	}
}

func InitSender(profile int) *Context {
	var ctx *C.struct_rist_ctx
	C.rist_sender_create(&ctx, C.int(profile))
	return &Context{ptr: ctx}
}

func InitReceiver(profile int) *Context {
	var ctx *C.struct_rist_ctx
	C.rist_receiver_create(&ctx, C.int(profile))
	return &Context{ptr: ctx}
}

func (c *Context) SetLogLevel(level int) {
	if c.ptr != nil {
		C.rist_set_log_level(c.ptr, C.int(level))
	}
}

func (c *Context) SetOutputIP(ip string) {
	cip := C.CString(ip)
	defer C.free(unsafe.Pointer(cip))
	if c.ptr != nil {
		C.rist_add_peer(c.ptr, nil, cip, nil, nil)
	}
}
