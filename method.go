package reflectx

import (
	"fmt"
	"log"
	"reflect"
	"unsafe"
)

// memmove copies size bytes to dst from src. No write barriers are used.
//go:linkname memmove reflect.memmove
func memmove(dst, src unsafe.Pointer, size uintptr)

//go:linkname storeRcvr reflect.storeRcvr
func storeRcvr(v reflect.Value, p unsafe.Pointer)

//go:linkname unsafe_New reflect.unsafe_New
func unsafe_New(*rtype) unsafe.Pointer

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

	// update receiver type
	wt := reflect.TypeOf((*wrapper)(nil)).Elem()
	vt := totype(wt)
	vm := vt.exportedMethods()
	em := rt.exportedMethods()
	var infos []*methodInfo
	for i, m := range ms {
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
		// rewrite tfn
		ntyp := reflect.FuncOf(in, out, false)
		funcImpl := (*makeFuncImpl)(tovalue(&m.Func).ptr)
		funcImpl.ftyp = (*funcType)(unsafe.Pointer(totype(ntyp)))

		// rewrite ifn
		var inFields []reflect.StructField
		for i := 1; i < len(in); i++ {
			inFields = append(inFields, reflect.StructField{
				Name: fmt.Sprintf("Arg%v", i),
				Type: in[i],
			})
		}
		inTyp := reflect.StructOf(inFields)
		var outFields []reflect.StructField
		for i := 0; i < len(out); i++ {
			outFields = append(outFields, reflect.StructField{
				Name: fmt.Sprintf("Out%v", i),
				Type: out[i],
			})
		}
		outTyp := reflect.StructOf(outFields)
		sz := totype(inTyp).size
		var fnName string
		if len(out) > 0 {
			fnName = fmt.Sprintf("I%v_%v", i, sz)
		} else {
			fnName = fmt.Sprintf("N%v_%v", i, sz)
		}
		log.Println("--->", m.Name, fnName)
		if fm, ok := wt.MethodByName(fnName); ok {
			em[i].ifn = resolveReflectText(vt.textOff(vm[fm.Index].ifn))
		} else {
			log.Printf("warning cannot found wrapper method wrapper.%v\n", fnName)
		}
		infos = append(infos, &methodInfo{i, inTyp, outTyp})
	}
	typInfoMap[typ] = infos
	log.Println("---> typMap", typ, typInfoMap[typ])

	nt := &Named{Name: styp.Name(), PkgPath: styp.PkgPath(), Type: typ, Kind: TkStruct}
	ntypeMap[typ] = nt

	return typ
}

var (
	typInfoMap = make(map[reflect.Type][]*methodInfo)
	ptrTypeMap = make(map[unsafe.Pointer]reflect.Type)
)

type methodInfo struct {
	index  int
	inTyp  reflect.Type
	outTyp reflect.Type
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

func New(typ reflect.Type) reflect.Value {
	v := reflect.New(typ)
	if IsNamed(typ) {
		log.Println("new", typInfoMap[typ])
		storeValue(v)
	}
	return v
}

func toElem(typ reflect.Type) reflect.Type {
	if typ.Kind() == reflect.Ptr {
		return typ.Elem()
	}
	return typ
}

func storeValue(v reflect.Value) {
	ptrTypeMap[tovalue(&v).ptr] = toElem(v.Type())
}
