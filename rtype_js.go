// +build js,!wasm

package reflectx

import (
	"reflect"
	"unsafe"

	"github.com/goplusjs/gopherjs/js"
)

//go:linkname reflectType reflect.reflectType
func reflectType(typ *js.Object) *_rtype

//go:linkname setKindType reflect.setKindType
func setKindType(rt *_rtype, kindType interface{})

//go:linkname newNameOff reflect.newNameOff
func newNameOff(n name) nameOff

//go:linkname newTypeOff reflect.newTypeOff
func newTypeOff(rt *_rtype) typeOff

//go:linkname makeValue reflect.makeValue
func makeValue(t *_rtype, v *js.Object, fl flag) reflect.Value

// func jsType(typ Type) *js.Object {
// 	return js.InternalObject(typ).Get("jsType")
// }

// func reflectType(typ *js.Object) *rtype {
// 	return _reflectType(typ, internalStr)
// }

func toStructType(t *_rtype) *structType {
	kind := js.InternalObject(t).Get("kindType")
	return (*structType)(unsafe.Pointer(kind.Unsafe()))
}

func toKindType(t *_rtype) unsafe.Pointer {
	return unsafe.Pointer(js.InternalObject(t).Get("kindType").Unsafe())
}

func toUncommonType(t *_rtype) *uncommonType {
	kind := js.InternalObject(t).Get("uncommonType")
	if kind == js.Undefined {
		return nil
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

func (t *_rtype) ptrTo() *_rtype {
	return reflectType(js.Global.Call("$ptrType", jsType(t)))
}

func (t *_rtype) uncommon() *uncommonType {
	return toUncommonType(t)
}

func (t *_rtype) exportedMethods() []method {
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
		4, // array
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

func tovalue(v *reflect.Value) *Value {
	return (*Value)(unsafe.Pointer(v))
}

func NamedTypeOf(pkg string, name string, from reflect.Type) (typ reflect.Type) {
	rt, _ := newType(pkg, name, from, 0, 0)
	setTypeName(rt, pkg, name)
	typ = toType(rt)
	nt := &Named{Name: name, PkgPath: pkg, Type: typ, From: from, Kind: TkType}
	ntypeMap[typ] = nt
	return
}

var (
	jsUncommonTyp = js.InternalObject(reflect.TypeOf((*_rtype)(nil))).Get("uncommonType").Get("constructor")
)

func resetUncommonType(rt *_rtype, mcount int, xcount int) *uncommonType {
	ut := jsUncommonTyp.New()
	v := js.InternalObject(ut).Get("_methods").Get("constructor")
	ut.Set("xcount", xcount)
	ut.Set("mcount", mcount)
	ut.Set("_methods", js.Global.Call("$makeSlice", v, mcount, mcount))
	ut.Set("jsType", jsType(rt))
	js.InternalObject(rt).Set("uncommonType", ut)
	return (*uncommonType)(unsafe.Pointer(ut.Unsafe()))
}

func newType(pkg string, name string, styp reflect.Type, xcount int, mcount int) (*_rtype, []method) {
	kind := styp.Kind()
	var obj *js.Object
	switch kind {
	default:
		obj = fnNewType.Invoke(styp.Size(), kind, name, true, pkg, false, nil)
	case reflect.Array:
		obj = fnNewType.Invoke(styp.Size(), kind, name, true, pkg, false, nil)
		obj.Call("init", jsType(styp.Elem()), styp.Len())
	case reflect.Slice:
		obj = fnNewType.Invoke(styp.Size(), kind, name, true, pkg, false, nil)
		obj.Call("init", jsType(styp.Elem()))
	case reflect.Map:
		obj = fnNewType.Invoke(styp.Size(), kind, name, true, pkg, false, nil)
		obj.Call("init", jsType(styp.Key()), jsType(styp.Elem()))
	case reflect.Ptr:
		obj = fnNewType.Invoke(styp.Size(), kind, name, true, pkg, false, nil)
		obj.Call("init", jsType(styp.Elem()))
	case reflect.Chan:
		obj = fnNewType.Invoke(styp.Size(), kind, name, true, pkg, false, nil)
		obj.Call("init", jsType(styp.Elem()))
	case reflect.Func:
		obj = fnNewType.Invoke(styp.Size(), kind, name, true, pkg, false, nil)
		obj.Call("init", jsType(styp).Get("params"), jsType(styp).Get("results"), styp.IsVariadic())
	case reflect.Interface:
		obj = fnNewType.Invoke(styp.Size(), kind, name, true, pkg, false, nil)
		obj.Call("init", jsType(styp).Get("methods"))
	case reflect.Struct:
		fields := js.Global.Get("Array").New()
		for i := 0; i < styp.NumField(); i++ {
			sf := styp.Field(i)
			jsf := js.Global.Get("Object").New()
			jsf.Set("prop", sf.Name)
			jsf.Set("name", sf.Name)
			jsf.Set("exported", true)
			jsf.Set("typ", jsType(sf.Type))
			jsf.Set("tag", sf.Tag)
			jsf.Set("embedded", sf.Anonymous)
			fields.SetIndex(i, jsf)
		}
		fn := js.MakeFunc(func(this *js.Object, args []*js.Object) interface{} {
			this.Set("$val", this)
			for i := 0; i < fields.Length(); i++ {
				f := fields.Index(i)
				if len(args) > i && args[i] != js.Undefined {
					this.Set(f.Get("prop").String(), args[i])
				} else {
					this.Set(f.Get("prop").String(), f.Get("typ").Call("zero"))
				}
			}
			return nil
		})
		obj = fnNewType.Invoke(styp.Size(), kind, styp.Name(), false, pkg, false, fn)
		obj.Call("init", pkg, fields)
	}
	rt := reflectType(obj)
	if kind == reflect.Func || kind == reflect.Interface {
		return rt, nil
	}
	rt.tflag |= tflagUncommon
	ut := resetUncommonType(rt, xcount, mcount)
	return rt, ut._methods
}

func totype(typ reflect.Type) *_rtype {
	v := reflect.Zero(typ)
	rt := (*Value)(unsafe.Pointer(&v)).typ
	return rt
}

func internalStr(strObj *js.Object) string {
	var c struct{ str string }
	js.InternalObject(c).Set("str", strObj) // get string without internalizing
	return c.str
}

type funcType struct {
	_rtype   `reflect:"func"`
	inCount  uint16
	outCount uint16

	_in  []*_rtype
	_out []*_rtype
}

func (t *funcType) in() []*_rtype {
	return t._in
}

func (t *funcType) out() []*_rtype {
	return t._out
}

func jsType(typ interface{}) *js.Object {
	return js.InternalObject(typ).Get("jsType")
}
