package reflectx

import (
	"fmt"
	"log"
	"reflect"
	"testing"
	"unsafe"
)

// memmove copies size bytes to dst from src. No write barriers are used.
//go:linkname memmove reflect.memmove
func memmove(dst, src unsafe.Pointer, size uintptr)

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

func MethodOf(styp reflect.Type, ms []reflect.Method) reflect.Type {
	var methods []method
	for _, m := range ms {
		ptr := tovalue(&m.Func).ptr
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

	st := (*structType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
	ut := (*uncommonType)(unsafe.Pointer(tt.Elem().Field(1).UnsafeAddr()))
	copy(tt.Elem().Field(2).Slice(0, len(methods)).Interface().([]method), methods)
	ut.mcount = uint16(len(methods))
	ut.xcount = ut.mcount
	ut.moff = uint32(unsafe.Sizeof(uncommonType{}))

	ort := totype(styp)
	ost := toStructType(ort)

	st.size = ort.size
	st.tflag = ort.tflag
	st.kind = ort.kind
	st.fields = ost.fields
	st.fieldAlign = ost.fieldAlign
	st.str = resolveReflectName(ort.nameOff(ort.str))

	rt := (*rtype)(unsafe.Pointer(st))
	setTypeName(rt, "main", "PointX")
	typ := toType(rt)

	for _, m := range ms {
		mtyp := m.Func.Type()
		var out []reflect.Type
		for i := 0; i < mtyp.NumOut(); i++ {
			out = append(out, mtyp.Out(i))
		}
		ntyp := reflect.FuncOf([]reflect.Type{typ}, out, false)
		funcImpl := (*makeFuncImpl)(tovalue(&m.Func).ptr)
		funcImpl.ftyp = (*funcType)(unsafe.Pointer(totype(ntyp)))
	}

	nt := &Named{Name: styp.Name(), PkgPath: styp.PkgPath(), Type: typ, Kind: TkStruct}
	ntypeMap[typ] = nt

	return typ
}

func MethodByType(typ reflect.Type, index int) reflect.Method {
	m := typ.Method(index)
	if _, ok := ntypeMap[typ]; ok {
		tovalue(&m.Func).flag |= flagIndir
	}
	return m
}

func myString(s struct {
	x int
	y int
}) string {
	log.Println("myString---->", s)
	return "myString"
}

type makeFuncImpl struct {
	code   uintptr
	stack  *bitVector // ptrmap for both args and results
	argLen uintptr    // just args
	ftyp   *funcType
	fn     func([]Value) []Value
}

type bitVector struct {
	n    uint32 // number of bits
	data []byte
}

func TestMethod(t *testing.T) {
	fs := []reflect.StructField{
		reflect.StructField{Name: "X", Type: reflect.TypeOf(0)},
		reflect.StructField{Name: "Y", Type: reflect.TypeOf(0)},
	}
	typ := NamedStructOf("main", "Point", fs)
	t.Log(typ)

	styp := reflect.TypeOf("")
	mtyp := reflect.FuncOf(nil, []reflect.Type{styp}, false)
	mfn := reflect.MakeFunc(mtyp, func(args []reflect.Value) []reflect.Value {
		log.Println("--->", args[0].Type(), args[0].Field(1), args[0].NumMethod())
		return []reflect.Value{reflect.ValueOf("Hello")}
	})
	nt := MethodOf(typ, []reflect.Method{
		reflect.Method{
			Name: "String",
			Type: mtyp,
			Func: mfn,
		},
	})
	t.Log(nt.NumField(), nt.NumMethod())
	m := MethodByType(nt, 0)
	v := reflect.New(nt).Elem()
	v.Field(0).SetInt(100)
	v.Field(1).SetInt(200)
	log.Println("--->", m.Func.Type(), v.Type())
	r := m.Func.Call([]reflect.Value{v})
	t.Log(r)
	//log.Println("--->", v.Interface())
}

func _TestMe(t *testing.T) {
	//t.Log("Test")
	// _t := reflect.TypeOf((*Point)(nil)).Elem()
	// //totype(_t).kind |= kindDirectIface
	// _v := reflect.New(_t).Elem()
	// _m := _t.Method(0)
	// _m2 := _v.Method(0)
	// log.Print("--->", _v, _m.Type, tovalue(&_m2).flag, totype(_t).kind)
	// mm := totype(_t).exportedMethods()[0]
	// //_m2.Call(nil)
	// //return
	// styp := reflect.TypeOf("")
	// mtyp := reflect.FuncOf(nil, []reflect.Type{styp}, false)
	// mfn := reflect.MakeFunc(mtyp, func(args []reflect.Value) []reflect.Value {
	// 	log.Println("-->", args)
	// 	return nil
	// })
	// typ := makeMethod([]reflect.Method{
	// 	reflect.Method{
	// 		Name: "String",
	// 		Type: mtyp,
	// 		Func: mfn,
	// 	},
	// })
	// rt := (*rtype)(unsafe.Pointer(typ))
	// //rt.kind |= kindDirectIface
	// setTypeName(rt, "main", "Point")
	// mtyp = reflect.FuncOf([]reflect.Type{toType(rt)}, []reflect.Type{styp}, false)
	// totype(mtyp).kind |= kindDirectIface
	// //	totype(mtyp).kind |= flagMethod
	// ms := typ.exportedMethods()
	// ms[0].name = resolveReflectName(newName("String", "", true))
	// //ms[0].mtyp = resolveReflectType(totype(mtyp))
	// mfn = reflect.MakeFunc(mtyp, func(args []reflect.Value) []reflect.Value {
	// 	log.Println("===>", args)
	// 	return []reflect.Value{reflect.ValueOf("Hello")}
	// })
	// //tovalue(&mfn).flag |= flagMethod
	// ptr := unsafe.Pointer(tovalue(&mfn).ptr)
	// ptr = totype(_t).textOff(mm.ifn)
	// //memmove(tovalue(&mfn).ptr, totype(_t).textOff(mm.ifn), 8)
	// memmove(unsafe.Pointer(&ptr), unsafe.Pointer(&tovalue(&mfn).ptr), 16)

	// //*(*unsafe.Pointer)(unsafe.Pointer(ptr)) = ptr
	// //ptr = *(*unsafe.Pointer)(ptr)

	// _ = mm
	// //ptr = unsafe.Pointer(&ptr)          //*(*unsafe.Pointer)(&ptr)
	// ms[0].ifn = resolveReflectText(ptr) //tovalue(&mfn).ptr)
	// ms[0].tfn = resolveReflectText(ptr) //tovalue(&mfn).ptr)

	// log.Println(typ, typ.Kind(), typ.NumMethod())
	// //v := reflect.New(toType(rt)).Elem()
	// //log.Println(v, v.NumMethod())
	// m := typ.Method(0)
	// log.Println(tovalue(&m.Func).flag, uint16(reflect.Func), flagMethod)
	// //tovalue(&m.Func).flag |= flagMethod
	// log.Println(tovalue(&m.Func).flag)

	// mv := tovalue(&m.Func)
	// //tovalue(&m.Func).flag |= flagIndir
	// v := reflect.New(toType(rt)).Elem()
	// // m.Func.Call([]reflect.Value{v})

	// // return
	// //m2 := v.Method(0)
	// //tovalue(&m2).flag |= flagIndir
	// // log.Println("-->m2", tovalue(&m2).flag)
	// // tovalue(&m2).ptr = *(*unsafe.Pointer)(tovalue(&m2).ptr)
	// // m2.Call(nil)
	// log.Println(v)
	return

	//log.Println("new", v)
	// m.Func.Call([]reflect.Value{v})
	// //m2.Func.Call([]reflect.Value{v})
	// log.Println("--->", unsafe.Pointer(m.Func.Pointer()))
	// log.Println("--->", tovalue(&mfn).ptr, mv.ptr, *(*unsafe.Pointer)(mv.ptr))
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
