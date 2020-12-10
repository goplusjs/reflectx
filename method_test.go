package reflectx

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

// memmove copies size bytes to dst from src. No write barriers are used.
//go:linkname memmove reflect.memmove
func memmove(dst, src unsafe.Pointer, size uintptr)

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
	typ := toType(rt)

	for _, m := range ms {
		mtyp := m.Func.Type()
		var in []reflect.Type
		in = append(in, typ)
		for i := 0; i < mtyp.NumIn(); i++ {
			in = append(in, mtyp.In(i))
		}
		var out []reflect.Type
		for i := 0; i < mtyp.NumOut(); i++ {
			out = append(out, mtyp.Out(i))
		}
		ntyp := reflect.FuncOf(in, out, false)
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

var (
	intTyp = reflect.TypeOf(0)
	strTyp = reflect.TypeOf("")
	iType  = reflect.TypeOf((*interface{})(nil)).Elem()
)

func TestValueMethod(t *testing.T) {
	fs := []reflect.StructField{
		reflect.StructField{Name: "X", Type: reflect.TypeOf(0)},
		reflect.StructField{Name: "Y", Type: reflect.TypeOf(0)},
	}
	typ := NamedStructOf("main", "Point", fs)
	mtyp := reflect.FuncOf([]reflect.Type{}, []reflect.Type{strTyp}, false)
	mfn := reflect.MakeFunc(mtyp, func(args []reflect.Value) []reflect.Value {
		info := fmt.Sprintf("info:{%v %v}", args[0].Field(0), args[0].Field(1))
		return []reflect.Value{reflect.ValueOf(info)}
	})
	nt := MethodOf(typ, []reflect.Method{
		reflect.Method{
			Name: "String",
			Type: mtyp,
			Func: mfn,
		},
	})
	ms := totype(nt).exportedMethods()

	w := wraper{}

	vw := reflect.ValueOf(w)
	vt := tovalue(&vw).typ

	m0 := vt.exportedMethods()[0]
	ms[0].ifn = resolveReflectText(vt.textOff(m0.ifn))
	// vm := tovalue(&vw.Method(0))
	// ms[0].ifn = resolveReflectText(vm.ptr)
	//ms[0].tfn = resolveReflectText(vt.textOff(m0.tfn))

	//ifn := reflect.ValueOf((*wraper).IOs)
	// tfn := reflect.ValueOf((wraper).Test)
	// log.Println(ifn, unsafe.Pointer(ifn.Pointer()), tovalue(&ifn).ptr, tovalue(&ifn).flag, uint8(reflect.Func))
	// log.Println(tfn, unsafe.Pointer(tfn.Pointer()), tovalue(&tfn).ptr, tovalue(&tfn).flag)
	// log.Println(mfn, unsafe.Pointer(mfn.Pointer()), tovalue(&mfn).ptr)

	v0 := reflect.New(nt)
	v := v0.Elem()
	v.Field(0).SetInt(100)
	v.Field(1).SetInt(200)

	valueMap[w] = valueMethod{v, 0}

	r := v.Method(0).Call(nil)
	if len(r) != 1 || r[0].String() != "info:{100 200}" {
		t.Fatal("bad method call", r)
	}
	t.Log("call String() string", v)
}

type valueMethod struct {
	v     reflect.Value
	index int
}

var (
	valueMap = make(map[interface{}]valueMethod)
)

type wraper struct {
}

func (w wraper) IOs() []byte {
	v := valueMap[w]
	r := MethodByType(v.v.Type(), v.index).Func.Call([]reflect.Value{v.v})
	s := r[0].String()
	return *(*[]byte)(unsafe.Pointer(&s))
}

func TestTypeMethod(t *testing.T) {
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
