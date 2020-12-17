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

func TestValueMethod2(t *testing.T) {
	fs := []reflect.StructField{
		reflect.StructField{Name: "X", Type: reflect.TypeOf(0)},
		reflect.StructField{Name: "Y", Type: reflect.TypeOf(0)},
		reflect.StructField{Name: "_mhash", PkgPath: "main", Type: reflect.TypeOf(0)},
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
