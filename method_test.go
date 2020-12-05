package reflectx

import (
	"fmt"
	"log"
	"reflect"
	"testing"
	"unsafe"
)

type Point struct {
	X int
	Y int
}

func (pt Point) String() string {
	fmt.Println("~~~~", pt.X, pt.Y)
	return "[MyPoint]"
}

func MyFunc(pt Point) {
	fmt.Println(pt)
}

func makeMethod(ms []reflect.Method) *structType {
	var methods []method
	for _, m := range ms {
		ptr := m.Func.Pointer()
		methods = append(methods, method{
			name: resolveReflectName(newName(m.Name, "", isExported(m.Name))),
			mtyp: resolveReflectType(totype(m.Type)),
			ifn:  resolveReflectText(unsafe.Pointer(ptr)),
			tfn:  resolveReflectText(unsafe.Pointer(ptr)),
		})
	}
	tt := reflect.New(reflect.StructOf([]reflect.StructField{
		{Name: "S", Type: reflect.TypeOf(structType{})},
		{Name: "U", Type: reflect.TypeOf(uncommonType{})},
		{Name: "M", Type: reflect.ArrayOf(len(methods), reflect.TypeOf(methods[0]))},
	}))

	typ := (*structType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
	ut := (*uncommonType)(unsafe.Pointer(tt.Elem().Field(1).UnsafeAddr()))
	copy(tt.Elem().Field(2).Slice(0, len(methods)).Interface().([]method), methods)
	ut.mcount = uint16(len(methods))
	ut.xcount = ut.mcount
	ut.moff = uint32(unsafe.Sizeof(uncommonType{}))
	typ.kind = uint8(reflect.Struct)
	typ.tflag = tflagUncommon
	return typ
}

func TestMe(t *testing.T) {
	t.Log("Test")
	_t := reflect.TypeOf((*Point)(nil)).Elem()
	//totype(_t).kind |= kindDirectIface
	_v := reflect.New(_t).Elem()
	_m := _t.Method(0)
	_m2 := _v.Method(0)
	log.Print("--->", _v, _m.Type, tovalue(&_m2).flag, totype(_t).kind)
	mm := totype(_t).exportedMethods()[0]
	//_m2.Call(nil)
	//return
	styp := reflect.TypeOf("")
	mtyp := reflect.FuncOf(nil, []reflect.Type{styp}, false)
	mfn := reflect.MakeFunc(mtyp, func(args []reflect.Value) []reflect.Value {
		log.Println("-->", args)
		return nil
	})
	typ := makeMethod([]reflect.Method{
		reflect.Method{
			Name: "String",
			Type: mtyp,
			Func: mfn,
		},
	})
	rt := (*rtype)(unsafe.Pointer(typ))
	//rt.kind |= kindDirectIface
	setTypeName(rt, "main", "Point")
	mtyp = reflect.FuncOf([]reflect.Type{toType(rt)}, []reflect.Type{styp}, false)
	totype(mtyp).kind |= kindDirectIface
	//	totype(mtyp).kind |= flagMethod
	ms := typ.exportedMethods()
	ms[0].name = resolveReflectName(newName("String", "", true))
	//ms[0].mtyp = resolveReflectType(totype(mtyp))
	mfn = reflect.MakeFunc(mtyp, func(args []reflect.Value) []reflect.Value {
		log.Println("===>", args)
		return []reflect.Value{reflect.ValueOf("Hello")}
	})
	//tovalue(&mfn).flag |= flagMethod
	ptr := unsafe.Pointer(tovalue(&mfn).ptr)
	ptr = totype(_t).textOff(mm.ifn)
	//ptr = unsafe.Pointer(&ptr)          //*(*unsafe.Pointer)(&ptr)
	ms[0].ifn = resolveReflectText(ptr) //tovalue(&mfn).ptr)
	ms[0].tfn = resolveReflectText(ptr) //tovalue(&mfn).ptr)

	log.Println(typ, typ.Kind(), typ.NumMethod())
	//v := reflect.New(toType(rt)).Elem()
	//log.Println(v, v.NumMethod())
	m := typ.Method(0)
	log.Println(tovalue(&m.Func).flag, uint16(reflect.Func), flagMethod)
	//tovalue(&m.Func).flag |= flagMethod
	log.Println(tovalue(&m.Func).flag)

	mv := tovalue(&m.Func)
	tovalue(&m.Func).flag |= flagIndir
	v := reflect.New(toType(rt)).Elem()
	//m.Func.Call([]reflect.Value{v})
	m2 := v.Method(0)
	//tovalue(&m2).flag |= flagIndir
	log.Println("-->m2", tovalue(&m2).flag)
	//m2.Call(nil)
	log.Println(v)
	return

	//log.Println("new", v)
	m.Func.Call([]reflect.Value{v})
	//m2.Func.Call([]reflect.Value{v})
	log.Println("--->", unsafe.Pointer(m.Func.Pointer()))
	log.Println("--->", tovalue(&mfn).ptr, mv.ptr, *(*unsafe.Pointer)(mv.ptr))
	//m.Func.Call(nil)
	//log.Println("-->", tovalue(m.Func).flag&flagIndir)
	//tovalue(m.Func).flag |= flagIndir
	// tovalue(m).flag &= ^flagMethod
	// dst := tovalue(m.Func).typ
	// src := tovalue(mfn).typ
	// copyType(dst, src)
	// d := (*funcType)(unsafe.Pointer(dst))
	// s := (*funcType)(unsafe.Pointer(src))
	// d.inCount = s.inCount
	// d.outCount = s.outCount

	//log.Println(m.Func.Pointer(), mfn.Pointer())
	//m.Func.Call([]reflect.Value{v})
	//m.Call(nil)
	//mfn.Call(nil)
	//log.Println(m)
	//log.Println(mfn.Type())
}
