// +build js,!wasm

package reflectx

import (
	"unsafe"

	"github.com/goplusjs/gopherjs/js"
)

func toStructType(t *rtype) *structType {
	kind := js.InternalObject(t).Get("kindType")
	return (*structType)(unsafe.Pointer(kind.Unsafe()))
}

func toUncommonType(t *rtype) *uncommonType {
	kind := js.InternalObject(t).Get("uncommonType")
	if kind == js.Undefined {
		ut := &uncommonType{}
		js.InternalObject(t).Set("uncommonType", js.InternalObject(ut))
		js.InternalObject(ut).Set("rtype", js.InternalObject(t))
		return ut
	}
	return (*uncommonType)(unsafe.Pointer(kind.Unsafe()))
}

type uncommonType struct {
	pkgPath nameOff
	mcount  uint16
	xcount  uint16
	moff    uint32

	_methods []method
}
