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

func toUncommonType(typ *rtype) *uncommonType {
	kind := js.InternalObject(typ).Get("uncommonType")
	if kind == js.Undefined {
		return nil
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

func setUncommonTypePkgPath(typ *rtype, n nameOff) {
	ut := toUncommonType(typ)
	if ut == nil {
		ut = &uncommonType{pkgPath: n}
		js.InternalObject(typ).Set("uncommonType", js.InternalObject(ut))
	} else {
		ut.pkgPath = n
	}
	typ.tflag |= tflagUncommon
}
