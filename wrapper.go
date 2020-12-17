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
		if t := foundTypeByPtr(ptr); t != nil {
			if !AddVerifyField {
				log.Printf("warring, guess type %v by %v\n", t, ptr)
			}
			typ = t
		} else {
			log.Println("cannot found ptr type", w.data)
			return nil
		}
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
