package reflectx

import (
	"fmt"
	"log"
	"reflect"
	"unsafe"
)

var (
	AddVerifyField  = true
	verifyFieldType = totype(reflect.TypeOf(unsafe.Pointer(nil)))
	verifyFieldName = "reflectx_verify"
)

// memmove copies size bytes to dst from src. No write barriers are used.
//go:linkname memmove reflect.memmove
func memmove(dst, src unsafe.Pointer, size uintptr)

func MethodOf(styp reflect.Type, ms []reflect.Method, pms []reflect.Method) reflect.Type {
	rt, typ := methodOf(styp, ms)
	prt, _ := methodOf(reflect.PtrTo(styp), pms)
	rt.ptrToThis = resolveReflectType(prt)
	(*ptrType)(unsafe.Pointer(prt)).elem = rt
	return typ
}

func methodOf(styp reflect.Type, ms []reflect.Method) (*rtype, reflect.Type) {
	if ms == nil {
		return totype(styp), styp
	}
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
	ort := totype(styp)
	var tt reflect.Value
	var rt *rtype
	switch styp.Kind() {
	case reflect.Struct:
		tt = reflect.New(reflect.StructOf([]reflect.StructField{
			{Name: "S", Type: reflect.TypeOf(structType{})},
			{Name: "U", Type: reflect.TypeOf(uncommonType{})},
			{Name: "M", Type: reflect.ArrayOf(len(methods), reflect.TypeOf(methods[0]))},
		}))
		st := (*structType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
		ost := toStructType(ort)
		if styp.Kind() == reflect.Struct {
			st.fields = ost.fields
			if AddVerifyField {
				st.fields = append(st.fields, structField{
					name: newName(verifyFieldName, "", false),
					typ:  verifyFieldType,
				})
			}
		}
		rt = (*rtype)(unsafe.Pointer(st))
	case reflect.Ptr:
		tt = reflect.New(reflect.StructOf([]reflect.StructField{
			{Name: "S", Type: reflect.TypeOf(ptrType{})},
			{Name: "U", Type: reflect.TypeOf(uncommonType{})},
			{Name: "M", Type: reflect.ArrayOf(len(methods), reflect.TypeOf(methods[0]))},
		}))
		st := (*ptrType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
		rt = (*rtype)(unsafe.Pointer(st))
		st.elem = ((*ptrType)(unsafe.Pointer(ort))).elem
	}

	ut := (*uncommonType)(unsafe.Pointer(tt.Elem().Field(1).UnsafeAddr()))
	copy(tt.Elem().Field(2).Slice(0, len(methods)).Interface().([]method), methods)

	ut.mcount = uint16(len(methods))
	ut.xcount = ut.mcount
	ut.moff = uint32(unsafe.Sizeof(uncommonType{}))

	rt.size = ort.size
	rt.tflag = ort.tflag | tflagUncommon
	rt.kind = ort.kind
	rt.fieldAlign = ort.fieldAlign
	rt.str = resolveReflectName(ort.nameOff(ort.str))

	typ := toType(rt)

	// update receiver type, rewrite tfn&ifn
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
		ifn := icall(i, int(sz), len(out) > 0)
		if ifn == nil {
			log.Printf("warning cannot wrapper method index:%v, size: %v\n", i, sz)
		} else {
			em[i].ifn = resolveReflectText(unsafe.Pointer(reflect.ValueOf(ifn).Pointer()))
		}
		infos = append(infos, &methodInfo{i, inTyp, outTyp})
	}
	typInfoMap[typ] = infos

	nt := &Named{Name: styp.Name(), PkgPath: styp.PkgPath(), Type: typ, Kind: TkStruct}
	ntypeMap[typ] = nt

	return rt, typ
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
	fn     func([]reflect.Value) []reflect.Value
}

type bitVector struct {
	n    uint32 // number of bits
	data []byte
}

func New(typ reflect.Type) reflect.Value {
	v := reflect.New(typ)
	if IsNamed(typ) {
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
	ptr := tovalue(&v).ptr
	ptrTypeMap[ptr] = toElem(v.Type())
	if AddVerifyField {
		if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
			v := FieldByName(v.Elem(), verifyFieldName)
			if v.IsValid() {
				v.SetPointer(ptr)
			}
		}
	}
}

func foundTypeByPtr(ptr unsafe.Pointer) reflect.Type {
	typ, ok := ptrTypeMap[ptr]
	if ok {
		return typ
	}
	for p, typ := range ptrTypeMap {
		v2 := reflect.NewAt(typ, ptr).Elem()
		v1 := reflect.NewAt(typ, p).Elem()
		if reflect.DeepEqual(v1.Interface(), v2.Interface()) {
			if !AddVerifyField {
				log.Printf("no verify, found type %v by %v\n", typ, ptr)
			}
			return typ
		}
	}
	return nil
}

func icall_x(i int, this uintptr, p []byte) []byte {
	ptr := unsafe.Pointer(this)

	log.Println("-----------> icall", i, unsafe.Pointer(this), p)

	typ := foundTypeByPtr(ptr)
	if typ == nil {
		log.Println("cannot found ptr type", ptr)
		return nil
	}
	infos, ok := typInfoMap[typ]
	if !ok {
		log.Println("cannot found type info", typ)
	}
	info := infos[i]
	method := MethodByType(typ, info.index)
	var in []reflect.Value
	inCount := method.Type.NumIn()
	in = make([]reflect.Value, inCount, inCount)
	in[0] = reflect.NewAt(typ, ptr).Elem()
	if inCount > 1 {
		inArgs := reflect.NewAt(info.inTyp, unsafe.Pointer(&p[0])).Elem()
		for i := 1; i < inCount; i++ {
			in[i] = inArgs.Field(i - 1)
		}
	}
	r := method.Func.Call(in)
	if len(r) > 0 {
		out := reflect.New(info.outTyp).Elem()
		for i, v := range r {
			out.Field(i).Set(v)
		}
		return *(*[]byte)(tovalue(&out).ptr)
	}
	return nil
}
