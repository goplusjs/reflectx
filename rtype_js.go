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

func (t *uncommonType) exportedMethods() []method {
	if t.xcount == 0 {
		return nil
	}
	return t._methods[:t.xcount:t.xcount]
}

func (t *rtype) ptrTo() *rtype {
	return reflectType(js.Global.Call("$ptrType", jsType(t)))
}

func (t *rtype) uncommon() *uncommonType {
	return toUncommonType(t)
}

func (t *rtype) exportedMethods() []method {
	ut := t.uncommon()
	if ut == nil {
		return nil
	}
	return ut.exportedMethods()
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
	rt, _ := newType(from, 0, 0)
	setTypeName(rt, pkgpath, name)
	typ = toType(rt)
	nt := &Named{Name: name, PkgPath: pkgpath, Type: typ, From: from, Kind: TkType}
	ntypeMap[typ] = nt
	return
}

func newType(styp reflect.Type, xcount int, mcount int) (*rtype, []method) {
	var rt *rtype
	var typ reflect.Type
	kind := styp.Kind()
	switch kind {
	default:
		obj := fnNewType.Invoke(sizes[kind], kind, "", true, "", false, nil)
		rt = reflectType(obj)
	case reflect.Array:
		elem := NamedTypeOf("", "_", styp.Elem())
		typ = reflect.ArrayOf(styp.Len(), elem)
		rt = totype(typ)
		src := totype(styp)
		copyType(rt, src)
		d := (*arrayType)(getKindType(rt))
		s := (*arrayType)(getKindType(src))
		d.elem = s.elem
		d.slice = s.slice
		d.len = s.len
	case reflect.Slice:
		elem := NamedTypeOf("", "_", styp.Elem())
		typ = reflect.SliceOf(elem)
		rt = totype(typ)
		dst := totype(typ)
		src := totype(styp)
		copyType(dst, src)
		d := (*sliceType)(getKindType(dst))
		s := (*sliceType)(getKindType(src))
		d.elem = s.elem
	case reflect.Map:
		key := NamedTypeOf("", "_", styp.Key())
		elem := NamedTypeOf("", "_", styp.Elem())
		typ = reflect.MapOf(key, elem)
		rt = totype(typ)
		src := totype(styp)
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
		elem := NamedTypeOf("", "_", styp.Elem())
		typ = reflect.PtrTo(elem)
		rt = totype(typ)
		src := totype(styp)
		copyType(rt, src)
		d := (*ptrType)(getKindType(rt))
		s := (*ptrType)(getKindType(src))
		d.elem = s.elem
	case reflect.Chan:
		elem := NamedTypeOf("", "_", styp.Elem())
		typ = reflect.ChanOf(styp.ChanDir(), elem)
		rt = totype(typ)
		src := totype(styp)
		copyType(rt, src)
		d := (*chanType)(getKindType(rt))
		s := (*chanType)(getKindType(src))
		d.elem = s.elem
		d.dir = s.dir
	case reflect.Func:
		numIn := styp.NumIn()
		in := make([]reflect.Type, numIn, numIn)
		for i := 0; i < numIn; i++ {
			in[i] = styp.In(i)
		}
		numOut := styp.NumOut()
		out := make([]reflect.Type, numOut, numOut)
		for i := 0; i < numOut; i++ {
			out[i] = styp.Out(i)
		}
		out = append(out, emptyType())
		typ = reflect.FuncOf(in, out, styp.IsVariadic())
		rt = totype(typ)
		src := totype(styp)
		d := (*jsFuncType)(getKindType(rt))
		s := (*jsFuncType)(getKindType(src))
		d.inCount = s.inCount
		d.outCount = s.outCount
		d._in = s._in
		d._out = s._out
	case reflect.Interface:
		t := fnNewType.Invoke(styp.Size(), kind, "", true, "", false, nil)
		rt = reflectType(t)
		typ = toType(rt)
		src := totype(styp)
		copyType(rt, src)
		d := (*interfaceType)(getKindType(rt))
		s := (*interfaceType)(getKindType(src))
		for _, m := range s.methods {
			d.methods = append(d.methods, imethod{
				name: resolveReflectName(s.nameOff(m.name)),
				typ:  resolveReflectType(s.typeOff(m.typ)),
			})
		}
	case reflect.Struct:
		var fields []reflect.StructField
		if styp.Kind() == reflect.Struct {
			for i := 0; i < styp.NumField(); i++ {
				fs := styp.Field(i)
				if !isExported(fs.Name) {
					fs.PkgPath = "main"
				}
				fields = append(fields, fs)
			}
		}
		fields = append(fields, reflect.StructField{
			Name: unusedName(),
			Type: tyEmptyStruct,
		})
		typ = StructOf(fields)
		rt = totype(typ)
		st := toStructType(rt)
		st.fields = st.fields[:len(st.fields)-1]
		copyType(rt, totype(styp))
	}
	ut := toUncommonType(rt)
	ut.mcount = uint16(mcount)
	ut.xcount = uint16(xcount)
	ut.moff = uint32(unsafe.Sizeof(uncommonType{}))
	ut._methods = make([]method, mcount, mcount)
	if kind == reflect.Func || kind == reflect.Interface {
		return rt, nil
	}
	return rt, ut._methods
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

func jsType(typ *rtype) *js.Object {
	return js.InternalObject(typ).Get("jsType")
}

// func (v Value) object() *js.Object {
// 	if v.typ.Kind() == reflect.Array || v.typ.Kind() == reflect.Struct {
// 		return js.InternalObject(v.ptr)
// 	}
// 	if v.flag&flagIndir != 0 {
// 		val := js.InternalObject(v.ptr).Call("$get")
// 		if val != js.Global.Get("$ifaceNil") && val.Get("constructor") != jsType(v.typ) {
// 			switch v.typ.Kind() {
// 			case reflect.Uint64, reflect.Int64:
// 				val = jsType(v.typ).New(val.Get("$high"), val.Get("$low"))
// 			case reflect.Complex64, reflect.Complex128:
// 				val = jsType(v.typ).New(val.Get("$real"), val.Get("$imag"))
// 			case reflect.Slice:
// 				if val == val.Get("constructor").Get("nil") {
// 					val = jsType(v.typ).Get("nil")
// 					break
// 				}
// 				newVal := jsType(v.typ).New(val.Get("$array"))
// 				newVal.Set("$offset", val.Get("$offset"))
// 				newVal.Set("$length", val.Get("$length"))
// 				newVal.Set("$capacity", val.Get("$capacity"))
// 				val = newVal
// 			}
// 		}
// 		return js.InternalObject(val.Unsafe())
// 	}
// 	return js.InternalObject(v.ptr)
// }
