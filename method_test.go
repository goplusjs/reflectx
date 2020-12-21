package reflectx_test

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/goplusjs/reflectx"
)

var (
	byteTyp = reflect.TypeOf(byte('a'))
	boolTyp = reflect.TypeOf(true)
	intTyp  = reflect.TypeOf(0)
	strTyp  = reflect.TypeOf("")
	iType   = reflect.TypeOf((*interface{})(nil)).Elem()
)

func TestDynamicPoint(t *testing.T) {
	fs := []reflect.StructField{
		reflect.StructField{Name: "X", Type: reflect.TypeOf(0)},
		reflect.StructField{Name: "Y", Type: reflect.TypeOf(0)},
	}
	styp := reflectx.NamedStructOf("main", "Point", fs)
	var typ reflect.Type
	mString := reflectx.MakeMethod(
		"String",
		false,
		reflect.FuncOf(nil, []reflect.Type{strTyp}, false),
		func(args []reflect.Value) []reflect.Value {
			v := args[0]
			info := fmt.Sprintf("(%v,%v)", v.Field(0), v.Field(1))
			return []reflect.Value{reflect.ValueOf(info)}
		},
	)
	mAdd := reflectx.MakeMethod(
		"Add",
		false,
		reflect.FuncOf([]reflect.Type{styp}, []reflect.Type{styp}, false),
		func(args []reflect.Value) []reflect.Value {
			v := reflectx.New(typ).Elem()
			v.Field(0).SetInt(args[0].Field(0).Int() + args[1].Field(0).Int())
			v.Field(1).SetInt(args[0].Field(1).Int() + args[1].Field(1).Int())
			return []reflect.Value{v}
		},
	)
	mSet := reflectx.MakeMethod(
		"Set",
		true,
		reflect.FuncOf([]reflect.Type{intTyp, intTyp}, nil, false),
		func(args []reflect.Value) (result []reflect.Value) {
			v := args[0].Elem()
			v.Field(0).Set(args[1])
			v.Field(1).Set(args[2])
			return
		},
	)
	typ = reflectx.MethodOf(styp, []reflectx.Method{
		mAdd,
		mString,
		mSet,
	})
	ptrType := reflect.PtrTo(typ)

	for i := 0; i < typ.NumMethod(); i++ {
		log.Println(typ.Method(i))
	}

	pt1 := reflectx.New(typ).Elem()
	pt1.Field(0).SetInt(100)
	pt1.Field(1).SetInt(200)

	pt2 := reflectx.New(typ).Elem()
	pt2.Field(0).SetInt(300)
	pt2.Field(1).SetInt(400)

	// log.Println(pt1.MethodByName("String").Call(nil))

	log.Println(pt1, pt2)

	m, _ := reflectx.MethodByName(typ, "Add")
	r0 := m.Func.Call([]reflect.Value{pt1, pt2})
	log.Println("tfn", r0[0])
	r0 = pt1.MethodByName("Add").Call([]reflect.Value{pt2})
	log.Println("ifn", r0[0])

	// // ptrtype
	m, _ = reflectx.MethodByName(reflect.PtrTo(typ), "Add")
	r0 = m.Func.Call([]reflect.Value{pt1.Addr(), pt2})
	log.Println("addr tfn", r0[0])
	r0 = pt1.Addr().MethodByName("Add").Call([]reflect.Value{pt2})
	log.Println("addr ifn", r0[0])

	// //return
	m0, _ := reflectx.MethodByName(ptrType, "Set")
	log.Println("addr tfn", m0)
	m0.Func.Call([]reflect.Value{pt1.Addr(), reflect.ValueOf(-100), reflect.ValueOf(-200)})
	log.Println(pt1, pt1.Addr())
	pt1.Addr().MethodByName("Set").Call([]reflect.Value{reflect.ValueOf(1), reflect.ValueOf(2)})
	log.Println(pt1, pt1.Addr())
}

func TestDynamicMethod(t *testing.T) {
	fs := []reflect.StructField{
		reflect.StructField{Name: "X", Type: reflect.TypeOf(0)},
		reflect.StructField{Name: "Y", Type: reflect.TypeOf(0)},
	}
	styp := reflectx.NamedStructOf("main", "Point", fs)
	mString := reflectx.MakeMethod(
		"String",
		false,
		reflect.FuncOf(nil, []reflect.Type{strTyp}, false),
		func(args []reflect.Value) (result []reflect.Value) {
			v := args[0] //.Elem()
			s := fmt.Sprintf("%v-%v", v.Field(0), v.Field(1))
			result = append(result, reflect.ValueOf(s))
			return
		})
	mSet := reflectx.MakeMethod(
		"Set",
		true,
		reflect.FuncOf([]reflect.Type{intTyp, intTyp}, nil, false),
		func(args []reflect.Value) (result []reflect.Value) {
			v := args[0].Elem()
			v.Field(0).Set(args[1])
			v.Field(1).Set(args[2])
			return
		})
	mGet := reflectx.MakeMethod(
		"Get",
		false,
		reflect.FuncOf(nil, []reflect.Type{intTyp, intTyp}, false),
		func(args []reflect.Value) (result []reflect.Value) {
			v := args[0]
			return []reflect.Value{v.Field(0), v.Field(1)}
		})
	mAppend := reflectx.MakeMethod(
		"Append",
		false,
		reflect.FuncOf([]reflect.Type{reflect.SliceOf(intTyp)}, []reflect.Type{intTyp}, true),
		func(args []reflect.Value) (result []reflect.Value) {
			var sum int64
			for i := 0; i < args[1].Len(); i++ {
				sum += args[1].Index(i).Int()
			}
			return []reflect.Value{reflect.ValueOf(int(sum))}
		})

	typ := reflectx.MethodOf(styp, []reflectx.Method{
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
	pt := reflectx.New(typ).Elem()
	pt.Field(0).SetInt(100)
	pt.Field(1).SetInt(200)
	r := pt.MethodByName("Get").Call(nil)
	log.Println(r[0], r[1])
	r = pt.Addr().MethodByName("Get").Call(nil)
	log.Println(r[0], r[1])
	m, _ := reflectx.MethodByName(typ, "Get")
	r = m.Func.Call([]reflect.Value{pt})
	log.Println(r)
	m, _ = reflectx.MethodByName(ptrType, "Get")
	r = m.Func.Call([]reflect.Value{pt.Addr()})
	log.Println(r)
	pt.Addr().MethodByName("Set").Call([]reflect.Value{reflect.ValueOf(300), reflect.ValueOf(400)})
	log.Println(pt, pt.Addr())

	r = pt.MethodByName("Append").Call([]reflect.Value{reflect.ValueOf(100), reflect.ValueOf(200), reflect.ValueOf(300)})
	log.Println(r[0])

	r = pt.Addr().MethodByName("Append").Call([]reflect.Value{reflect.ValueOf(100), reflect.ValueOf(200), reflect.ValueOf(300)})
	log.Println(r[0])
}
