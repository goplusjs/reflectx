// +build js,!wasm

package reflectx

import (
	"unsafe"

	"github.com/gopherjs/gopherjs/js"
)

func toStructType(typ *rtype) *structType {
	kind := js.InternalObject(typ).Get("kindType")
	return (*structType)(unsafe.Pointer(kind.Unsafe()))
}
