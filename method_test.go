package reflectx

import (
	"fmt"
	"log"
	"reflect"
	"testing"
	"unsafe"
)

var (
	byteTyp = reflect.TypeOf(byte('a'))
	boolTyp = reflect.TypeOf(true)
	intTyp  = reflect.TypeOf(0)
	strTyp  = reflect.TypeOf("")
	iType   = reflect.TypeOf((*interface{})(nil)).Elem()
)

type My struct {
	id unsafe.Pointer
}

func (w My) Test(this unsafe.Pointer, p [8]byte) string {
	log.Println("---> my", this, unsafe.Pointer(&w), unsafe.Pointer(w.id), *(*int)(unsafe.Pointer(&p[0])))
	return "hello"
}

func MyTest(this unsafe.Pointer, p [8]byte) string {
	log.Println("---> mytest", this, *(*int)(unsafe.Pointer(&p[0])))
	return "hello"
}

var (
	saved = make(map[interface{}]bool)
)

func TestValueMethod2(t *testing.T) {
	fs := []reflect.StructField{
		reflect.StructField{Name: "X", Type: reflect.TypeOf(0)},
		reflect.StructField{Name: "Y", Type: reflect.TypeOf(0)},
	}
	typ := NamedStructOf("main", "Point", fs)
	tyString := reflect.FuncOf(nil, []reflect.Type{strTyp}, false)
	fnString := reflect.MakeFunc(tyString, func(args []reflect.Value) []reflect.Value {
		log.Println("---> call String args", args)
		info := fmt.Sprintf("%v-%v", args[0].Field(0), args[0].Field(1))
		//info := fmt.Sprintf("info:{%v %v}", args[0].Field(0), args[0].Field(1))
		return []reflect.Value{reflect.ValueOf(info) /*, reflect.ValueOf(-1024)*/}
	})
	tyTest := reflect.FuncOf([]reflect.Type{intTyp}, []reflect.Type{strTyp}, false)
	fnTest := reflect.MakeFunc(tyTest, func(args []reflect.Value) []reflect.Value {
		log.Println("---> call Test args", args)
		info := fmt.Sprintf("%v-%v-%v", args[1], args[0].Field(0), args[0].Field(1))
		//info := fmt.Sprintf("info:{%v %v}", args[0].Field(0), args[0].Field(1))
		return []reflect.Value{reflect.ValueOf(info) /*, reflect.ValueOf(-1024)*/}
	})
	tySet := reflect.FuncOf([]reflect.Type{intTyp, strTyp}, nil, false)
	fnSet := reflect.MakeFunc(tySet, func(args []reflect.Value) []reflect.Value {
		log.Println("----> set", args[0])
		// args[0].Field(0).Set(args[1])
		// args[0].Field(1).Set(args[2])
		//log.Println(" set ", args)
		return nil
	})

	nt := MethodOf(typ, []reflect.Method{
		reflect.Method{
			Name: "String",
			Type: tyString,
			Func: fnString,
		},
		reflect.Method{
			Name: "Test",
			Type: tyTest,
			Func: fnTest,
		},
		reflect.Method{
			Name: "Set",
			Type: tySet,
			Func: fnSet,
		},
	})
	v0 := New(nt)
	v := v0.Elem()
	v.Field(0).SetInt(1)
	v.Field(1).SetInt(1)

	rt := totype(nt)

	wp := My{}
	fn := reflect.ValueOf(wp.Test)
	wp2 := My{}
	fn2 := reflect.ValueOf(wp2.Test)
	log.Println("vvv", tovalue(&v).ptr, fn.Type())
	log.Println("www", unsafe.Pointer(&wp), unsafe.Pointer(&wp2), unsafe.Pointer(&wp.id), unsafe.Pointer(&wp2.id))
	log.Println(fn.Pointer(), fn2.Pointer())
	myTest := func(this unsafe.Pointer, p [8]byte) string {
		log.Println("---> mytest2", this, *(*int)(unsafe.Pointer(&p[0])))
		return "hello"
	}

	fn = reflect.ValueOf(myTest)

	rt.exportedMethods()[1].ifn = resolveReflectText(unsafe.Pointer(fn.Pointer()))
	r0 := v.Method(1).Call([]reflect.Value{reflect.ValueOf(100)})
	log.Println("---> return1", r0)

	//	var check bool
	// var entry uintptr
	rfn := reflect.MakeFunc(fn.Type(), func(args []reflect.Value) []reflect.Value {
		log.Println("---> make func")
		// if !check {
		// 	pc, _, _, ok := runtime.Caller(0)
		// 	if ok {
		// 		entry = runtime.FuncForPC(pc).Entry()
		// 	}
		// }
		return []reflect.Value{reflect.ValueOf("hello")}
	})
	myTest = rfn.Interface().(func(unsafe.Pointer, [8]byte) string)
	fn = reflect.ValueOf(myTest)
	rt.exportedMethods()[1].ifn = resolveReflectText(unsafe.Pointer(fn.Pointer()))
	// saved[rfn] = true
	// rfn.Call([]reflect.Value{reflect.ValueOf(unsafe.Pointer(nil)), reflect.ValueOf([8]byte{})})
	// check = true
	// log.Println("entry", fn.Type(), rfn.Type(), entry)
	//rt.exportedMethods()[1].ifn = resolveReflectText(unsafe.Pointer(myTest))
	r0 = v.Method(1).Call([]reflect.Value{reflect.ValueOf(100)})
	log.Println("---> return2", r0)

	//log.Println(v)
	return
	MethodByType(nt, 0).Func.Call([]reflect.Value{v})

	r := v.Method(0).Call(nil)
	i := v.Interface()
	iv := reflect.ValueOf(i)
	storeValue(iv)

	r = v.Method(1).Call([]reflect.Value{reflect.ValueOf(100)})
	log.Println(r)

	fmt.Println(v)
	v.Method(2).Call([]reflect.Value{reflect.ValueOf(-1), reflect.ValueOf("word")})
	fmt.Println(v)
}

func _TestTypeMethod(t *testing.T) {
	fs := []reflect.StructField{
		reflect.StructField{Name: "X", Type: reflect.TypeOf(0)},
		reflect.StructField{Name: "Y", Type: reflect.TypeOf(0)},
	}
	typ := NamedStructOf("main", "Point", fs)
	mtyp := reflect.FuncOf([]reflect.Type{intTyp}, []reflect.Type{strTyp}, false)
	mfn := reflect.MakeFunc(mtyp, func(args []reflect.Value) []reflect.Value {
		info := fmt.Sprintf("info:%v-%v", args[0], args[1])
		return []reflect.Value{reflect.ValueOf(info)}
	})
	nt := MethodOf(typ, []reflect.Method{
		reflect.Method{
			Name: "Test",
			Type: mtyp,
			Func: mfn,
		},
	})
	m := MethodByType(nt, 0)
	v := reflect.New(nt).Elem()
	v.Field(0).SetInt(100)
	v.Field(1).SetInt(200)
	r := m.Func.Call([]reflect.Value{v, reflect.ValueOf(300)})
	if len(r) != 1 || r[0].String() != "info:{100 200}-300" {
		t.Fatal("bad method call", r)
	}
}
