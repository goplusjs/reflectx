package reflectx

import (
	"log"
	"reflect"
	"unsafe"
)

type wrapper struct {
	data unsafe.Pointer
}

func (w wrapper) call(i int, p []byte) []byte {
	ptr := unsafe.Pointer(w.data)
	typ, ok := ptrTypeMap[ptr]
	if !ok {
		log.Println("cannot found ptr type", w.data)
		return nil
	}
	infos, ok := typInfoMap[typ]
	if !ok {
		log.Println("cannot found type info", typ)
	}
	info := infos[i]
	method := MethodByType(typ, info.index)
	var in []reflect.Value
	inCount := method.Type.NumIn()
	in = make([]reflect.Value, inCount, inCount)
	in[0] = reflect.NewAt(typ, ptr).Elem()
	if inCount > 1 {
		inArgs := reflect.NewAt(info.inTyp, unsafe.Pointer(&p[0])).Elem()
		for i := 1; i < inCount; i++ {
			in[i] = inArgs.Field(i - 1)
		}
	}
	r := method.Func.Call(in)
	if len(r) > 0 {
		out := reflect.New(info.outTyp).Elem()
		for i, v := range r {
			out.Field(i).Set(v)
		}
		return *(*[]byte)(tovalue(&out).ptr)
	}
	return nil
}

// func (w wrapper) I0_0() []byte {
// 	return w.call(0, nil)
// }

// func (w wrapper) I0_8(p [8]byte) []byte {
// 	return w.call(0, p[:])
// }

// func (w wrapper) I1_8(p [8]byte) []byte {
// 	return w.call(1, p[:])
// }

// func (w wrapper) I0_16(p [16]byte) []byte {
// 	return w.call(0, p[:])
// }

// func (w wrapper) I0_24(p [24]byte) []byte {
// 	return w.call(0, p[:])
// }

// func (w wrapper) I0_32(p [32]byte) []byte {
// 	return w.call(0, p[:])
// }
