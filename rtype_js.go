// +build js,!wasm

package reflectx

import (
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

// func _NamedTypeOf(pkgpath string, name string, from reflect.Type) (typ reflect.Type) {
// 	kind := from.Kind()
// 	if kind >= reflect.Bool && kind <= reflect.Complex128 ||
// 		(kind == reflect.String || kind == reflect.UnsafePointer) {
// 		sname := name
// 		if pkgpath != "" {
// 			sname = pkgpath + "." + name
// 		}
// 		t := fnNewType.Invoke(sizes[kind], kind, sname, true, pkgpath, isExported(name), nil)
// 		typ = reflectType(t)
// 		nt := &Named{Name: name, PkgPath: pkgpath, Type: typ, From: from, Kind: TkType}
// 		ntypeMap[typ] = nt
// 		return typ
// 	}
// 	switch kind {
// 	case reflect.Array:
// 		elem := NamedTypeOf(pkgpath, "_", from.Elem())
// 		typ = reflect.ArrayOf(from.Len(), elem)
// 		dst := totype(typ)
// 		src := totype(from)
// 		copyType(dst, src)
// 		d := (*arrayType)(getKindType(dst))
// 		s := (*arrayType)(getKindType(src))
// 		d.elem = s.elem
// 		d.slice = s.slice
// 		d.len = s.len
// 		setTypeName(dst, pkgpath, name)
// 	case reflect.Slice:
// 		elem := NamedTypeOf(pkgpath, "_", from.Elem())
// 		typ = reflect.SliceOf(elem)
// 		rt := totype(typ)
// 		setTypeName(rt, pkgpath, name)
// 		dst := totype(typ)
// 		src := totype(from)
// 		copyType(dst, src)
// 		d := (*sliceType)(getKindType(dst))
// 		s := (*sliceType)(getKindType(src))
// 		d.elem = s.elem
// 	case reflect.Map:
// 		key := NamedTypeOf(pkgpath, "_", from.Key())
// 		elem := NamedTypeOf(pkgpath, "_", from.Elem())
// 		typ = reflect.MapOf(key, elem)
// 		dst := totype(typ)
// 		src := totype(from)
// 		copyType(dst, src)
// 		d := (*mapType)(getKindType(dst))
// 		s := (*mapType)(getKindType(src))
// 		d.key = s.key
// 		d.elem = s.elem
// 		d.bucket = s.bucket
// 		d.hasher = s.hasher
// 		d.keysize = s.keysize
// 		d.valuesize = s.valuesize
// 		d.bucketsize = s.bucketsize
// 		d.flags = s.flags
// 		dst.str = resolveReflectName(newName(name, "", isExported(name)))
// 		setTypeName(dst, pkgpath, name)
// 	case reflect.Ptr:
// 		elem := NamedTypeOf(pkgpath, "_", from.Elem())
// 		typ = reflect.PtrTo(elem)
// 		dst := totype(typ)
// 		src := totype(from)
// 		copyType(dst, src)
// 		d := (*ptrType)(getKindType(dst))
// 		s := (*ptrType)(getKindType(src))
// 		d.elem = s.elem
// 		setTypeName(dst, pkgpath, name)
// 	case reflect.Chan:
// 		elem := NamedTypeOf(pkgpath, "_", from.Elem())
// 		typ = reflect.ChanOf(from.ChanDir(), elem)
// 		dst := totype(typ)
// 		src := totype(from)
// 		copyType(dst, src)
// 		d := (*chanType)(getKindType(dst))
// 		s := (*chanType)(getKindType(src))
// 		d.elem = s.elem
// 		d.dir = s.dir
// 		setTypeName(dst, pkgpath, name)
// 	case reflect.Func:
// 		numIn := from.NumIn()
// 		in := make([]reflect.Type, numIn, numIn)
// 		for i := 0; i < numIn; i++ {
// 			in[i] = from.In(i)
// 		}
// 		numOut := from.NumOut()
// 		out := make([]reflect.Type, numOut, numOut)
// 		for i := 0; i < numOut; i++ {
// 			out[i] = from.Out(i)
// 		}
// 		out = append(out, emptyType())
// 		typ = reflect.FuncOf(in, out, from.IsVariadic())
// 		dst := totype(typ)
// 		src := totype(from)
// 		d := (*jsFuncType)(getKindType(dst))
// 		s := (*jsFuncType)(getKindType(src))
// 		d.inCount = s.inCount
// 		d.outCount = s.outCount
// 		d._in = s._in
// 		d._out = s._out
// 		setTypeName(dst, pkgpath, name)
// 	default:
// 		var fields []reflect.StructField
// 		if from.Kind() == reflect.Struct {
// 			for i := 0; i < from.NumField(); i++ {
// 				fields = append(fields, from.Field(i))
// 			}
// 		}
// 		fields = append(fields, reflect.StructField{
// 			Name: hashName(pkgpath, name),
// 			Type: typEmptyStruct,
// 		})
// 		typ = StructOf(fields)
// 		rt := totype(typ)
// 		st := toStructType(rt)
// 		st.fields = st.fields[:len(st.fields)-1]
// 		copyType(rt, totype(from))
// 		setTypeName(rt, pkgpath, name)
// 	}
// 	nt := &Named{Name: name, PkgPath: pkgpath, Type: typ, From: from, Kind: TkType}
// 	ntypeMap[typ] = nt
// 	return typ
// }

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

// func newType(styp reflect.Type, mcount int, xcount int) (rt *rtype, tt reflect.Value) {
// 	ort := totype(styp)
// 	switch styp.Kind() {
// 	case reflect.Struct:
// 		tt = reflect.New(reflect.StructOf([]reflect.StructField{
// 			{Name: "S", Type: reflect.TypeOf(structType{})},
// 			{Name: "U", Type: reflect.TypeOf(uncommonType{})},
// 			{Name: "M", Type: reflect.ArrayOf(mcount, reflect.TypeOf(method{}))},
// 		}))
// 		st := (*structType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
// 		ost := (*structType)(unsafe.Pointer(ort))
// 		st.fields = ost.fields
// 	case reflect.Ptr:
// 		tt = reflect.New(reflect.StructOf([]reflect.StructField{
// 			{Name: "S", Type: reflect.TypeOf(ptrType{})},
// 			{Name: "U", Type: reflect.TypeOf(uncommonType{})},
// 			{Name: "M", Type: reflect.ArrayOf(mcount, reflect.TypeOf(method{}))},
// 		}))
// 		st := (*ptrType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
// 		st.elem = totype(styp.Elem())
// 	case reflect.Interface:
// 		tt = reflect.New(reflect.StructOf([]reflect.StructField{
// 			{Name: "S", Type: reflect.TypeOf(interfaceType{})},
// 			{Name: "U", Type: reflect.TypeOf(uncommonType{})},
// 		}))
// 		st := (*interfaceType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
// 		ost := (*interfaceType)(unsafe.Pointer(ort))
// 		for _, m := range ost.methods {
// 			st.methods = append(st.methods, imethod{
// 				name: resolveReflectName(ost.nameOff(m.name)),
// 				typ:  resolveReflectType(ost.typeOff(m.typ)),
// 			})
// 		}
// 	case reflect.Slice:
// 		tt = reflect.New(reflect.StructOf([]reflect.StructField{
// 			{Name: "S", Type: reflect.TypeOf(sliceType{})},
// 			{Name: "U", Type: reflect.TypeOf(uncommonType{})},
// 			{Name: "M", Type: reflect.ArrayOf(mcount, reflect.TypeOf(method{}))},
// 		}))
// 		st := (*sliceType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
// 		st.elem = totype(styp.Elem())
// 	case reflect.Array:
// 		tt = reflect.New(reflect.StructOf([]reflect.StructField{
// 			{Name: "S", Type: reflect.TypeOf(arrayType{})},
// 			{Name: "U", Type: reflect.TypeOf(uncommonType{})},
// 			{Name: "M", Type: reflect.ArrayOf(mcount, reflect.TypeOf(method{}))},
// 		}))
// 		st := (*arrayType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
// 		ost := (*arrayType)(unsafe.Pointer(ort))
// 		st.elem = ost.elem
// 		st.slice = ost.slice
// 		st.len = ost.len
// 	case reflect.Chan:
// 		tt = reflect.New(reflect.StructOf([]reflect.StructField{
// 			{Name: "S", Type: reflect.TypeOf(chanType{})},
// 			{Name: "U", Type: reflect.TypeOf(uncommonType{})},
// 			{Name: "M", Type: reflect.ArrayOf(mcount, reflect.TypeOf(method{}))},
// 		}))
// 		st := (*chanType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
// 		ost := (*chanType)(unsafe.Pointer(ort))
// 		st.elem = ost.elem
// 		st.dir = ost.dir
// 	case reflect.Func:
// 		narg := styp.NumIn() + styp.NumOut()
// 		tt = reflect.New(reflect.StructOf([]reflect.StructField{
// 			{Name: "S", Type: reflect.TypeOf(funcType{})},
// 			{Name: "U", Type: reflect.TypeOf(uncommonType{})},
// 			{Name: "M", Type: reflect.ArrayOf(narg, reflect.TypeOf((*rtype)(nil)))},
// 		}))
// 		st := (*funcType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
// 		ost := (*funcType)(unsafe.Pointer(ort))
// 		st.inCount = ost.inCount
// 		st.outCount = ost.outCount
// 		if narg > 0 {
// 			args := make([]*rtype, narg, narg)
// 			for i := 0; i < styp.NumIn(); i++ {
// 				args[i] = totype(styp.In(i))
// 			}
// 			index := styp.NumIn()
// 			for i := 0; i < styp.NumOut(); i++ {
// 				args[index+i] = totype(styp.Out(i))
// 			}
// 			copy(tt.Elem().Field(2).Slice(0, narg).Interface().([]*rtype), args)
// 		}
// 	case reflect.Map:
// 		tt = reflect.New(reflect.StructOf([]reflect.StructField{
// 			{Name: "S", Type: reflect.TypeOf(mapType{})},
// 			{Name: "U", Type: reflect.TypeOf(uncommonType{})},
// 			{Name: "M", Type: reflect.ArrayOf(mcount, reflect.TypeOf(method{}))},
// 		}))
// 		st := (*mapType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
// 		ost := (*mapType)(unsafe.Pointer(ort))
// 		st.key = ost.key
// 		st.elem = ost.elem
// 		st.bucket = ost.bucket
// 		st.hasher = ost.hasher
// 		st.keysize = ost.keysize
// 		st.valuesize = ost.valuesize
// 		st.bucketsize = ost.bucketsize
// 		st.flags = ost.flags
// 	default:
// 		tt = reflect.New(reflect.StructOf([]reflect.StructField{
// 			{Name: "S", Type: reflect.TypeOf(rtype{})},
// 			{Name: "U", Type: reflect.TypeOf(uncommonType{})},
// 			{Name: "M", Type: reflect.ArrayOf(mcount, reflect.TypeOf(method{}))},
// 		}))
// 	}
// 	rt = (*rtype)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
// 	rt.size = ort.size
// 	rt.tflag = ort.tflag | tflagUncommon
// 	rt.kind = ort.kind
// 	rt.align = ort.align
// 	rt.fieldAlign = ort.fieldAlign
// 	rt.gcdata = ort.gcdata
// 	rt.ptrdata = ort.ptrdata
// 	rt.str = resolveReflectName(ort.nameOff(ort.str))
// 	ut := (*uncommonType)(unsafe.Pointer(tt.Elem().Field(1).UnsafeAddr()))
// 	// copy(tt.Elem().Field(2).Slice(0, len(methods)).Interface().([]method), methods)
// 	ut.mcount = uint16(mcount)
// 	ut.xcount = uint16(xcount)
// 	ut.moff = uint32(unsafe.Sizeof(uncommonType{}))
// 	return
// }

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

func newType(styp reflect.Type, mcount int, xcount int) (*rtype, reflect.Value) {
	typ := js.InternalObject(styp).Get("jsType")
	fnObjStr := internalStr
	rt := &rtype{
		size: uintptr(typ.Get("size").Int()),
		kind: uint8(typ.Get("kind").Int()),
	}
	ntyp := js.InternalObject(toType(rt))
	js.InternalObject(rt).Set("jsType", ntyp)
	ntyp.Set("reflectType", js.InternalObject(rt))

	ut := &uncommonType{
		mcount: uint16(mcount),
		xcount: uint16(xcount),
	}
	js.InternalObject(ut).Set("jsType", ntyp)
	js.InternalObject(rt).Set("uncommonType", js.InternalObject(ut))

	switch rt.Kind() {
	case reflect.Array:
		setKindType(rt, &arrayType{
			elem: reflectType(typ.Get("elem")),
			len:  uintptr(typ.Get("len").Int()),
		})
	case reflect.Chan:
		dir := reflect.BothDir
		if typ.Get("sendOnly").Bool() {
			dir = reflect.SendDir
		}
		if typ.Get("recvOnly").Bool() {
			dir = reflect.RecvDir
		}
		setKindType(rt, &chanType{
			elem: reflectType(typ.Get("elem")),
			dir:  uintptr(dir),
		})
	case reflect.Func:
		params := typ.Get("params")
		in := make([]*rtype, params.Length())
		for i := range in {
			in[i] = reflectType(params.Index(i))
		}
		results := typ.Get("results")
		out := make([]*rtype, results.Length())
		for i := range out {
			out[i] = reflectType(results.Index(i))
		}
		outCount := uint16(results.Length())
		if typ.Get("variadic").Bool() {
			outCount |= 1 << 15
		}
		setKindType(rt, &funcType{
			rtype:    *rt,
			inCount:  uint16(params.Length()),
			outCount: outCount,
			_in:      in,
			_out:     out,
		})
	case reflect.Interface:
		methods := typ.Get("methods")
		imethods := make([]imethod, methods.Length())
		for i := range imethods {
			m := methods.Index(i)
			imethods[i] = imethod{
				name: newNameOff(newName(fnObjStr(m.Get("name")), "", fnObjStr(m.Get("pkg")) == "")),
				typ:  newTypeOff(reflectType(m.Get("typ"))),
			}
		}
		setKindType(rt, &interfaceType{
			rtype:   *rt,
			pkgPath: newName(fnObjStr(typ.Get("pkg")), "", false),
			methods: imethods,
		})
	case reflect.Map:
		setKindType(rt, &mapType{
			key:  reflectType(typ.Get("key")),
			elem: reflectType(typ.Get("elem")),
		})
	case reflect.Ptr:
		setKindType(rt, &ptrType{
			elem: reflectType(typ.Get("elem")),
		})
	case reflect.Slice:
		setKindType(rt, &sliceType{
			elem: reflectType(typ.Get("elem")),
		})
	case reflect.Struct:
		fields := typ.Get("fields")
		reflectFields := make([]structField, fields.Length())
		for i := range reflectFields {
			f := fields.Index(i)
			offsetEmbed := uintptr(i) << 1
			if f.Get("embedded").Bool() {
				offsetEmbed |= 1
			}
			reflectFields[i] = structField{
				name:        newName(fnObjStr(f.Get("name")), fnObjStr(f.Get("tag")), f.Get("exported").Bool()),
				typ:         reflectType(f.Get("typ")),
				offsetEmbed: offsetEmbed,
			}
		}
		setKindType(rt, &structType{
			rtype:   *rt,
			pkgPath: newName(fnObjStr(typ.Get("pkgPath")), "", false),
			fields:  reflectFields,
		})
	}

	return (*rtype)(unsafe.Pointer(typ.Get("reflectType").Unsafe())), reflect.Value{}
}
