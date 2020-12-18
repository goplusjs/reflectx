package reflectx

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

var (
	byteTyp = reflect.TypeOf(byte('a'))
	boolTyp = reflect.TypeOf(true)
	intTyp  = reflect.TypeOf(0)
	strTyp  = reflect.TypeOf("")
	iType   = reflect.TypeOf((*interface{})(nil)).Elem()
)

type My struct {
	x int
	y int
}

func (m *My) Set(x int, y int) {
	m.x = x
	m.y = x
}

func (m My) Get() (int, int) {
	return m.x, m.y
}

func (m My) ABC() {

}

func (m *My) A00() {

}

func _TestMy(t *testing.T) {
	ptr := reflect.TypeOf((*My)(nil))
	log.Println(ptr.NumMethod(), ptr.Elem().NumMethod())
	for i := 0; i < ptr.NumMethod(); i++ {
		log.Println(ptr.Method(i))
	}
	for _, m := range totype(ptr).exportedMethods() {
		log.Println(m)
	}
	for i := 0; i < ptr.Elem().NumMethod(); i++ {
		log.Println(ptr.Elem().Method(i))
	}
	for _, m := range totype(ptr.Elem()).exportedMethods() {
		log.Println(m)
	}
	v := reflect.New(ptr.Elem()).Elem()
	log.Println(v)
	m, _ := ptr.MethodByName("Get")
	r := m.Func.Call([]reflect.Value{v.Addr()})
	log.Println(r, tovalue(&m.Func))
	m2, _ := ptr.Elem().MethodByName("Get")
	r = m2.Func.Call([]reflect.Value{v})
	log.Println(r, tovalue(&m2.Func))
}

func TestDynamicMethod(t *testing.T) {
	fs := []reflect.StructField{
		reflect.StructField{Name: "X", Type: reflect.TypeOf(0)},
		reflect.StructField{Name: "Y", Type: reflect.TypeOf(0)},
	}
	styp := NamedStructOf("main", "Point", fs)
	mString := MakeMethod(
		"String",
		false,
		reflect.FuncOf(nil, []reflect.Type{strTyp}, false),
		func(args []reflect.Value) (result []reflect.Value) {
			log.Println("call String", args)
			v := args[0] //.Elem()
			s := fmt.Sprintf("%v-%v", v.Field(0), v.Field(1))
			result = append(result, reflect.ValueOf(s))
			return
		})
	mSet := MakeMethod(
		"Set",
		true,
		reflect.FuncOf([]reflect.Type{intTyp, intTyp}, nil, false),
		func(args []reflect.Value) (result []reflect.Value) {
			log.Println("call Set", args)
			v := args[0].Elem()
			v.Field(0).Set(args[1])
			v.Field(1).Set(args[2])
			return
		})
	mGet := MakeMethod(
		"Get",
		false,
		reflect.FuncOf(nil, []reflect.Type{intTyp, intTyp}, false),
		func(args []reflect.Value) (result []reflect.Value) {
			log.Println("call Get", args)
			v := args[0]
			return []reflect.Value{v.Field(0), v.Field(1)}
		})
	mAppend := MakeMethod(
		"Append",
		false,
		reflect.FuncOf([]reflect.Type{reflect.SliceOf(intTyp)}, []reflect.Type{intTyp}, true),
		func(args []reflect.Value) (result []reflect.Value) {
			var sum int64
			log.Println("append", args, args[1].Len())
			for i := 0; i < args[1].Len(); i++ {
				sum += args[1].Index(i).Int()
			}
			return []reflect.Value{reflect.ValueOf(int(sum))}
		})

	typ := MethodOf(styp, []Method{
		mString,
		mSet,
		mGet,
		mAppend,
	})
	ptrType := reflect.PtrTo(typ)
	for i := 0; i < ptrType.NumMethod(); i++ {
		log.Println("ptr", ptrType.Method(i))
	}
	for i := 0; i < typ.NumMethod(); i++ {
		log.Println("struct", typ.Method(i))
	}
	pt := New(typ).Elem()
	pt.Field(0).SetInt(100)
	pt.Field(1).SetInt(200)
	r := pt.MethodByName("Get").Call(nil)
	log.Println(r[0], r[1])
	r = pt.Addr().MethodByName("Get").Call(nil)
	log.Println(r[0], r[1])
	m, _ := MethodByName(typ, "Get")
	r = m.Func.Call([]reflect.Value{pt})
	log.Println(r)
	m, _ = MethodByName(ptrType, "Get")
	r = m.Func.Call([]reflect.Value{pt.Addr()})
	log.Println(r)
	pt.Addr().MethodByName("Set").Call([]reflect.Value{reflect.ValueOf(300), reflect.ValueOf(400)})
	log.Println(pt, pt.Addr())
}

func _TestValueMethod2(t *testing.T) {
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

	nt := MethodOf(typ, []Method{
		Method{
			Name: "String",
			Type: tyString,
			Func: fnString,
		},
		Method{
			Name: "Test",
			Type: tyTest,
			Func: fnTest,
		},
		Method{
			Name: "Set",
			Type: tySet,
			Func: fnSet,
		},
	})
	v0 := New(nt)
	v := v0.Elem()
	v.Field(0).SetInt(1)
	v.Field(1).SetInt(100)

	MethodByType(nt, 0).Func.Call([]reflect.Value{v})

	r := v.MethodByName("String").Call(nil)

	log.Println(r)
	log.Println(v)
	r = v.Method(1).Call([]reflect.Value{reflect.ValueOf(100)})
	log.Println(r)
	return

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
	nt := MethodOf(typ, []Method{
		Method{
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
