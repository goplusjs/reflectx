package reflectx

import (
	"fmt"
	"log"
	"reflect"
	"unsafe"
)

type wrapperMethod struct {
	receiver reflect.Value  // receiver as first argument
	method   reflect.Method // type call method
	inTyp    reflect.Type   // input struct without receiver
	outTyp   reflect.Type   // output struct
}

var (
	wrapperMap = make(map[interface{}]wrapperMethod)
)

type wrapper struct {
	data unsafe.Pointer
}

func (w wrapper) call(i int, p []byte) []byte {
	typ, ok := ptrTypeMap[unsafe.Pointer(w.data)]
	if !ok {
		log.Println("cannot found ptr type", w.data)
		return nil
	}
	infos, ok := typInfoMap[typ]
	if !ok {
		log.Println("cannot found type info", typ)
	}
	info := infos[i]
	method := typ.Method(info.index)
	inCount := method.Type.NumIn()
	var in []reflect.Value
	inArgs := reflect.NewAt(info.inTyp, unsafe.Pointer(&p[0])).Elem()
	in[0] = reflect.NewAt()
	for i := 1; i < inCount; i++ {
		in[i] = inArgs.Field(i - 1)
	}
	r := v.method.Func.Call(in)
	if len(r) > 0 {
		out := reflect.New(outTyp).Elem()
		for i, v := range r {
			out.Field(i).Set(v)
		}
		return *(*[]byte)(tovalue(&out).ptr)
	}

	log.Println(typ, infos)
	ptr := unsafe.Pointer(w.data)
	log.Println("--->", ptr, p) // &w.data[0])
	return nil

	v, ok := wrapperMap[w]
	if !ok {
		log.Fatalf("invalid wrapper:%v\n", w)
		return nil
	}
	//	var v wrapperMethod

	log.Println("--->", unsafe.Pointer(&w), tovalue(&v.receiver).ptr)

	inCount := v.method.Type.NumIn()
	in := make([]reflect.Value, inCount, inCount)
	var inFields []reflect.StructField
	for i := 1; i < inCount; i++ {
		typ := v.method.Type.In(i)
		inFields = append(inFields, reflect.StructField{
			Name: fmt.Sprintf("Arg%v", i),
			Type: typ,
		})
	}
	inTyp := reflect.StructOf(inFields)
	var outFields []reflect.StructField
	for i := 0; i < v.method.Type.NumOut(); i++ {
		typ := v.method.Type.Out(i)
		outFields = append(outFields, reflect.StructField{
			Name: fmt.Sprintf("Out%v", i),
			Type: typ,
		})
	}
	outTyp := reflect.StructOf(outFields)
	inArgs := reflect.NewAt(inTyp, unsafe.Pointer(&p[0])).Elem()
	in[0] = v.receiver
	for i := 1; i < inCount; i++ {
		in[i] = inArgs.Field(i - 1)
	}
	r := v.method.Func.Call(in)
	if len(r) > 0 {
		out := reflect.New(outTyp).Elem()
		for i, v := range r {
			out.Field(i).Set(v)
		}
		return *(*[]byte)(tovalue(&out).ptr)
	}
	return nil
}

func (w wrapper) I0() []byte {
	return w.call(0, nil)
}

func (w wrapper) I0_8(p [8]byte) []byte {
	return w.call(0, p[:])
}

func (w wrapper) I0_16(p [16]byte) []byte {
	return w.call(0, p[:])
}

func (w wrapper) I0_24(p [24]byte) []byte {
	return w.call(0, p[:])
}

func (w wrapper) I0_32(p [32]byte) []byte {
	return w.call(0, p[:])
}
