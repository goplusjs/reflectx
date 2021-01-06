// +build js,!wasm

package reflectx

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/goplusjs/gopherjs/js"
)

//go:linkname reflectType reflect.reflectType
func reflectType(typ *js.Object) *rtype

//go:linkname setKindType reflect.setKindType
func setKindType(rt *rtype, kindType interface{})

//go:linkname newNameOff reflect.newNameOff
func newNameOff(n name) nameOff

//go:linkname newTypeOff reflect.newTypeOff
func newTypeOff(rt *rtype) typeOff

// func jsType(typ Type) *js.Object {
// 	return js.InternalObject(typ).Get("jsType")
// }

// func reflectType(typ *js.Object) *rtype {
// 	return _reflectType(typ, internalStr)
// }

func toStructType(t *rtype) *structType {
	kind := js.InternalObject(t).Get("kindType")
	return (*structType)(unsafe.Pointer(kind.Unsafe()))
}

func toKindType(t *rtype) unsafe.Pointer {
	return unsafe.Pointer(js.InternalObject(t).Get("kindType").Unsafe())
}

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

func tovalue(v *reflect.Value) *Value {
	return (*Value)(unsafe.Pointer(v))
}

var (
	index int
)

func unusedName() string {
	index++
	return fmt.Sprintf("Gop_unused_%v", index)
}

func emptyType() reflect.Type {
	typ := reflect.StructOf([]reflect.StructField{
		reflect.StructField{
			Name: unusedName(),
			Type: tyEmptyStruct,
		}})
	rt := totype(typ)
	st := toStructType(rt)
	st.fields = st.fields[:len(st.fields)-1]
	st.str = resolveReflectName(newName("unused", "", false))
	return typ
}

func hashName(pkgpath string, name string) string {
	return fmt.Sprintf("Gop_Named_%d_%d", fnv1(0, pkgpath), fnv1(0, name))
}

func NamedTypeOf(pkgpath string, name string, from reflect.Type) (typ reflect.Type) {
	var rt *rtype
	kind := from.Kind()
	switch kind {
	default:
		if kind >= reflect.Bool && kind <= reflect.Complex128 ||
			(kind == reflect.String || kind == reflect.UnsafePointer) {
			sname := name
			if pkgpath != "" {
				sname = pkgpath + "." + name
			}
			t := fnNewType.Invoke(sizes[kind], kind, sname, true, pkgpath, isExported(name), nil)
			// typ = reflectType(t)
			// rt = totype(typ)
			rt = reflectType(t)
			typ = toType(rt)
		}
	case reflect.Array:
		elem := NamedTypeOf(pkgpath, "_", from.Elem())
		typ = reflect.ArrayOf(from.Len(), elem)
		rt = totype(typ)
		src := totype(from)
		copyType(rt, src)
		d := (*arrayType)(getKindType(rt))
		s := (*arrayType)(getKindType(src))
		d.elem = s.elem
		d.slice = s.slice
		d.len = s.len
	case reflect.Slice:
		elem := NamedTypeOf(pkgpath, "_", from.Elem())
		typ = reflect.SliceOf(elem)
		rt = totype(typ)
		dst := totype(typ)
		src := totype(from)
		copyType(dst, src)
		d := (*sliceType)(getKindType(dst))
		s := (*sliceType)(getKindType(src))
		d.elem = s.elem
	case reflect.Map:
		key := NamedTypeOf(pkgpath, "_", from.Key())
		elem := NamedTypeOf(pkgpath, "_", from.Elem())
		typ = reflect.MapOf(key, elem)
		rt = totype(typ)
		src := totype(from)
		copyType(rt, src)
		d := (*mapType)(getKindType(rt))
		s := (*mapType)(getKindType(src))
		d.key = s.key
		d.elem = s.elem
		d.bucket = s.bucket
		d.hasher = s.hasher
		d.keysize = s.keysize
		d.valuesize = s.valuesize
		d.bucketsize = s.bucketsize
		d.flags = s.flags
	case reflect.Ptr:
		elem := NamedTypeOf(pkgpath, "_", from.Elem())
		typ = reflect.PtrTo(elem)
		rt = totype(typ)
		src := totype(from)
		copyType(rt, src)
		d := (*ptrType)(getKindType(rt))
		s := (*ptrType)(getKindType(src))
		d.elem = s.elem
	case reflect.Chan:
		elem := NamedTypeOf(pkgpath, "_", from.Elem())
		typ = reflect.ChanOf(from.ChanDir(), elem)
		rt = totype(typ)
		src := totype(from)
		copyType(rt, src)
		d := (*chanType)(getKindType(rt))
		s := (*chanType)(getKindType(src))
		d.elem = s.elem
		d.dir = s.dir
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
		rt = totype(typ)
		src := totype(from)
		d := (*jsFuncType)(getKindType(rt))
		s := (*jsFuncType)(getKindType(src))
		d.inCount = s.inCount
		d.outCount = s.outCount
		d._in = s._in
		d._out = s._out
	case reflect.Interface:
		sname := name
		if pkgpath != "" {
			sname = pkgpath + "." + name
		}
		t := fnNewType.Invoke(from.Size(), kind, sname, true, pkgpath, isExported(name), nil)
		rt = reflectType(t)
		typ = toType(rt)
		src := totype(from)
		copyType(rt, src)
		d := (*interfaceType)(getKindType(rt))
		s := (*interfaceType)(getKindType(src))
		for _, m := range s.methods {
			d.methods = append(d.methods, imethod{
				name: resolveReflectName(s.nameOff(m.name)),
				typ:  resolveReflectType(s.typeOff(m.typ)),
			})
		}
		setTypeName(rt, pkgpath, name)
	case reflect.Struct:
		var fields []reflect.StructField
		if from.Kind() == reflect.Struct {
			for i := 0; i < from.NumField(); i++ {
				fs := from.Field(i)
				if !isExported(fs.Name) {
					fs.PkgPath = "main"
				}
				fields = append(fields, fs)
			}
		}
		fields = append(fields, reflect.StructField{
			Name: hashName(pkgpath, name),
			Type: tyEmptyStruct,
		})
		typ = StructOf(fields)
		rt = totype(typ)
		st := toStructType(rt)
		st.fields = st.fields[:len(st.fields)-1]
		copyType(rt, totype(from))
	}
	setTypeName(rt, pkgpath, name)
	nt := &Named{Name: name, PkgPath: pkgpath, Type: typ, From: from, Kind: TkType}
	ntypeMap[typ] = nt
	return typ
}

func totype(typ reflect.Type) *rtype {
	v := reflect.Zero(typ)
	rt := (*Value)(unsafe.Pointer(&v)).typ
	return rt
}

type jsFuncType struct {
	rtype    `reflect:"func"`
	inCount  uint16
	outCount uint16

	_in  []*rtype
	_out []*rtype
}

type makeFuncImpl struct {
	code   uintptr
	stack  *bitVector // ptrmap for both args and results
	argLen uintptr    // just args
	ftyp   *funcType
	fn     func([]reflect.Value) []reflect.Value
}

type bitVector struct {
	n    uint32 // number of bits
	data []byte
}

func internalStr(strObj *js.Object) string {
	var c struct{ str string }
	js.InternalObject(c).Set("str", strObj) // get string without internalizing
	return c.str
}

// func isWrapped(typ Type) bool {
// 	return jsType(typ).Get("wrapped").Bool()
// }

// func copyStruct(dst, src *js.Object, typ Type) {
// 	fields := jsType(typ).Get("fields")
// 	for i := 0; i < fields.Length(); i++ {
// 		prop := fields.Index(i).Get("prop").String()
// 		dst.Set(prop, src.Get(prop))
// 	}
// }

type funcType struct {
	rtype    `reflect:"func"`
	inCount  uint16
	outCount uint16

	_in  []*rtype
	_out []*rtype
}

func (t *funcType) in() []*rtype {
	return t._in
}

func (t *funcType) out() []*rtype {
	return t._out
}

func Interface(v reflect.Value) interface{} {
	return v.Interface()
}
