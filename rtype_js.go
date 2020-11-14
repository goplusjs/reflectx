// +build js,!wasm

package reflectx

import (
	"unsafe"

	"github.com/goplusjs/gopherjs/js"
)

func toStructType(typ *rtype) *structType {
	kind := js.InternalObject(typ).Get("kindType")
	return (*structType)(unsafe.Pointer(kind.Unsafe()))
}
