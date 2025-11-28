package libristwrapper

/*
#cgo pkg-config: librist
#include <librist/librist.h>
*/
import "C"

// LibraryVersion returns the linked librist library version string.
func LibraryVersion() string {
	return C.GoString(C.librist_version())
}

// LibraryAPIVersion returns the librist API version string.
func LibraryAPIVersion() string {
	return C.GoString(C.librist_api_version())
}
