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

type Key interface {
	call([]byte) []byte
}

var (
	wrapperMap = make(map[Key]wrapperMethod)
)

type wrapper struct {
	data []byte
}

func (w wrapper) call(p []byte) []byte {
	log.Println("---------------", p)
	return nil

	type M struct {
		X bool
		Y int
	}
	p0 := (uintptr)(unsafe.Pointer(&w))
	log.Println("-->", *(*M)(unsafe.Pointer(&w)), &p0)
	return nil
	v, ok := wrapperMap[&w]
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
	return w.call(nil)
}

func (w wrapper) I8(p [8]byte) []byte {
	return w.call(p[:])
}

func (w wrapper) I16(p [16]byte) []byte {
	return w.call(p[:])
}

func (w wrapper) I24(p [24]byte) []byte {
	return w.call(p[:])
}

func (w wrapper) I32(p [32]byte) []byte {
	return w.call(p[:])
}
