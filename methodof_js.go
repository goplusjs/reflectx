// +build js,!wasm

package reflectx

import "reflect"

func New(typ reflect.Type) reflect.Value {
	return reflect.New(typ)
}

func Interface(v reflect.Value) interface{} {
	return v.Interface()
}

func MethodByIndex(typ reflect.Type, index int) reflect.Method {
	return typ.Method(index)
}

func MethodByName(typ reflect.Type, name string) (m reflect.Method, ok bool) {
	return typ.MethodByName(name)
}

func methodOf(styp reflect.Type, methods []reflect.Method) reflect.Type {
	return styp
}
