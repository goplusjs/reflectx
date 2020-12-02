// +build js,!wasm

package reflectx

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/goplusjs/gopherjs/js"
)

//go:linkname reflectType reflect.reflectType
func reflectType(typ *js.Object) reflect.Type

//go:linkname _reflectType reflect.reflectType
func _reflectType(typ *js.Object) *rtype

//go:linkname setKindType reflect.setKindType
func setKindType(rt *rtype, kindType interface{})

func toStructType(t *rtype) *structType {
	kind := js.InternalObject(t).Get("kindType")
	return (*structType)(unsafe.Pointer(kind.Unsafe()))
}

func Test() {
	//fn := js.Global.Get("$newType")
	fn := js.Global.Get("makeType") //, 10, 20)
	v := fn.Invoke()
	fmt.Println(v.Get("size"))
	r0 := reflectType(v)
	v0 := reflect.New(r0).Elem()
	v0.Set(reflect.ValueOf(100))
	fmt.Println(v0)

	//	return reflectType(js.InternalObject(i).Get("constructor"))

	// fmt.Println(reflectType(v, "MyInt"))
	// t := emptyType()
	// r0 := totype(t)
	// r1 := reflectType(v, "MyInt")
	// //r1.kind = uint8(reflect.Int)
	// copyType(r0, r1)
	// setTypeName(r0, "main", "MyInt")
	// fmt.Println(t.Kind(), r1.Kind())
	// v1 := reflect.New(t).Elem()
	// v1.SetInt(10)
	// fmt.Println(v1)
}

// func reflectType(typ *js.Object, name string) *rtype {
// 	if typ.Get("reflectType") == js.Undefined {
// 		rt := &rtype{
// 			size: uintptr(typ.Get("size").Int()),
// 			kind: uint8(typ.Get("kind").Int()),
// 			str:  resolveReflectName(newName(name, "", isExported(name))),
// 		}
// 		js.InternalObject(rt).Set("jsType", typ)
// 		typ.Set("reflectType", js.InternalObject(rt))
// 	}
// 	return (*rtype)(unsafe.Pointer(typ.Get("reflectType").Unsafe()))
// }

func toUncommonType(t *rtype) *uncommonType {
	kind := js.InternalObject(t).Get("uncommonType")
	if kind == js.Undefined {
		ut := &uncommonType{}
		js.InternalObject(t).Set("uncommonType", js.InternalObject(ut))
		js.InternalObject(ut).Set("rtype", js.InternalObject(t))
		return ut
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

/*
var $kindBool = 1;
var $kindInt = 2;
var $kindInt8 = 3;
var $kindInt16 = 4;
var $kindInt32 = 5;
var $kindInt64 = 6;
var $kindUint = 7;
var $kindUint8 = 8;
var $kindUint16 = 9;
var $kindUint32 = 10;
var $kindUint64 = 11;
var $kindUintptr = 12;
var $kindFloat32 = 13;
var $kindFloat64 = 14;
var $kindComplex64 = 15;
var $kindComplex128 = 16;
var $kindArray = 17;
var $kindChan = 18;
var $kindFunc = 19;
var $kindInterface = 20;
var $kindMap = 21;
var $kindPtr = 22;
var $kindSlice = 23;
var $kindString = 24;
var $kindStruct = 25;
var $kindUnsafePointer = 26;

var $Bool          = $newType( 1, $kindBool,          "bool",           true, "", false, null);
var $Int           = $newType( 4, $kindInt,           "int",            true, "", false, null);
var $Int8          = $newType( 1, $kindInt8,          "int8",           true, "", false, null);
var $Int16         = $newType( 2, $kindInt16,         "int16",          true, "", false, null);
var $Int32         = $newType( 4, $kindInt32,         "int32",          true, "", false, null);
var $Int64         = $newType( 8, $kindInt64,         "int64",          true, "", false, null);
var $Uint          = $newType( 4, $kindUint,          "uint",           true, "", false, null);
var $Uint8         = $newType( 1, $kindUint8,         "uint8",          true, "", false, null);
var $Uint16        = $newType( 2, $kindUint16,        "uint16",         true, "", false, null);
var $Uint32        = $newType( 4, $kindUint32,        "uint32",         true, "", false, null);
var $Uint64        = $newType( 8, $kindUint64,        "uint64",         true, "", false, null);
var $Uintptr       = $newType( 4, $kindUintptr,       "uintptr",        true, "", false, null);
var $Float32       = $newType( 4, $kindFloat32,       "float32",        true, "", false, null);
var $Float64       = $newType( 8, $kindFloat64,       "float64",        true, "", false, null);
var $Complex64     = $newType( 8, $kindComplex64,     "complex64",      true, "", false, null);
var $Complex128    = $newType(16, $kindComplex128,    "complex128",     true, "", false, null);
var $String        = $newType( 8, $kindString,        "string",         true, "", false, null);
var $UnsafePointer = $newType( 4, $kindUnsafePointer, "unsafe.Pointer", true, "", false, null);
*/
//var $newType = function(size, kind, string, named, pkg, exported, constructor) {

var (
	fnNewType = js.Global.Get("$newType")
)

/*
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer
*/
var (
	kinds = []int{0, 1,
		2, 3, 4, 5, 6, // int
		7, 8, 9, 10, 11, // uint
	}
	sizes = []int{0, 1,
		4, 1, 2, 4, 8, // int
		4, 1, 2, 4, 8, // uint
		4,    // uintptr
		4, 8, // float
		8, 16, // complex
		4, //
		4, //
		4,
		4,
		4,
		4,
		12, // slice
		8,  // string
		4,  // struct
		4,  // UnsafePointer
	}
)

func getKindType(rt *rtype) unsafe.Pointer {
	return (unsafe.Pointer)(js.InternalObject(rt).Get("kindType").Unsafe())
}

func NamedTypeOf(pkgpath string, name string, from reflect.Type) (typ reflect.Type) {
	kind := from.Kind()
	if kind >= reflect.Bool && kind <= reflect.Complex128 ||
		(kind == reflect.String || kind == reflect.Ptr || kind == reflect.UnsafePointer) {
		sname := name
		if pkgpath != "" {
			sname = pkgpath + "." + name
		}
		t := fnNewType.Invoke(sizes[kind], kind, sname, true, pkgpath, isExported(name), nil)
		typ = reflectType(t)
		nt := &Named{Name: name, PkgPath: pkgpath, Type: typ, From: from, Kind: TkType}
		ntypeMap[typ] = nt
		return typ
	}

	switch kind {
	case reflect.Array:
		typ = reflect.ArrayOf(from.Len(), emptyType())
		dst := totype(typ)
		src := totype(from)
		copyType(dst, src)
		d := (*arrayType)(unsafe.Pointer(dst))
		s := (*arrayType)(unsafe.Pointer(src))
		d.elem = s.elem
		d.slice = s.slice
		d.len = s.len
		setTypeName(dst, pkgpath, name)
	case reflect.Slice:
		typ = reflect.SliceOf(NamedTypeOf(pkgpath, "_", from.Elem()))
		rt := totype(typ)
		setTypeName(rt, pkgpath, name)
		dst := totype(typ)
		src := totype(from)
		copyType(dst, src)
		d := (*sliceType)(getKindType(dst))
		s := (*sliceType)(getKindType(src))
		d.elem = s.elem
	case reflect.Map:
		typ = reflect.MapOf(emptyType(), emptyType())
		dst := totype(typ)
		src := totype(from)
		copyType(dst, src)
		d := (*mapType)(unsafe.Pointer(dst))
		s := (*mapType)(unsafe.Pointer(src))
		d.key = s.key
		d.elem = s.elem
		d.bucket = s.bucket
		d.hasher = s.hasher
		d.keysize = s.keysize
		d.valuesize = s.valuesize
		d.bucketsize = s.bucketsize
		d.flags = s.flags
		dst.str = resolveReflectName(newName(name, "", isExported(name)))
		setTypeName(dst, pkgpath, name)
	case reflect.Ptr:
		typ = reflect.PtrTo(emptyType())
		dst := totype(typ)
		src := totype(from)
		copyType(dst, src)
		d := (*ptrType)(unsafe.Pointer(dst))
		s := (*ptrType)(unsafe.Pointer(src))
		d.elem = s.elem
		setTypeName(dst, pkgpath, name)
	case reflect.Chan:
		typ = reflect.ChanOf(from.ChanDir(), emptyType())
		dst := totype(typ)
		src := totype(from)
		copyType(dst, src)
		d := (*chanType)(unsafe.Pointer(dst))
		s := (*chanType)(unsafe.Pointer(src))
		d.elem = s.elem
		d.dir = s.dir
		setTypeName(dst, pkgpath, name)
	case reflect.Func:
		numIn := from.NumIn()
		in := make([]reflect.Type, numIn, numIn)
		for i := 0; i < numIn; i++ {
			in[i] = from.In(i)
		}
		numOut := from.NumOut()
		out := make([]reflect.Type, numOut, numOut)
		for i := 0; i < numOut; i++ {
			out[i] = from.Out(i)
		}
		out = append(out, emptyType())
		typ = reflect.FuncOf(in, out, from.IsVariadic())
		dst := totype(typ)
		src := totype(from)
		d := (*funcType)(unsafe.Pointer(dst))
		s := (*funcType)(unsafe.Pointer(src))
		d.inCount = s.inCount
		d.outCount = s.outCount
		setTypeName(dst, pkgpath, name)
	default:
		var fields []reflect.StructField
		if from.Kind() == reflect.Struct {
			for i := 0; i < from.NumField(); i++ {
				fields = append(fields, from.Field(i))
			}
		}
		fields = append(fields, reflect.StructField{
			Name: hashName(pkgpath, name),
			Type: typEmptyStruct,
		})
		typ = StructOf(fields)
		rt := totype(typ)
		st := toStructType(rt)
		st.fields = st.fields[:len(st.fields)-1]
		copyType(rt, totype(from))
		setTypeName(rt, pkgpath, name)
	}
	nt := &Named{Name: name, PkgPath: pkgpath, Type: typ, From: from, Kind: TkType}
	ntypeMap[typ] = nt
	return typ
}

func totype(typ reflect.Type) *rtype {
	v := reflect.Zero(typ)
	rt := (*Value)(unsafe.Pointer(&v)).typ
	return rt
}
