package libristwrapper

/*
#cgo pkg-config: librist
#include <librist/librist.h>
*/
import "C"

func LibraryVersion() string {
	return C.GoString(C.rist_version())
}
