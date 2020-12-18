package reflectx

import (
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

var (
	EnableVerifyField = true
	EnableAllMethods  = true
	verifyFieldType   = reflect.TypeOf(unsafe.Pointer(nil))
	verifyFieldName   = "_reflectx_verify"
)

// memmove copies size bytes to dst from src. No write barriers are used.
//go:linkname memmove reflect.memmove
func memmove(dst, src unsafe.Pointer, size uintptr)

type Method struct {
	Name    string        // method Name
	Type    reflect.Type  // method type without receiver
	Func    reflect.Value // func with receiver as first argument
	Pointer bool          // receiver is pointer
}

// MakeMethod returns a new Method of the given Type
// that wraps the function fn.
//
//	- name: method name
//	- pointer: flag receiver struct or pointer
//	- typ: method func type without receiver
//	- fn: func with receiver as first argument
func MakeMethod(name string, pointer bool, typ reflect.Type, fn func(args []reflect.Value) (result []reflect.Value)) Method {
	return Method{
		Name:    name,
		Type:    typ,
		Func:    reflect.MakeFunc(typ, fn),
		Pointer: pointer,
	}
}

func MethodOf(styp reflect.Type, methods []Method) reflect.Type {
	sort.Slice(methods, func(i, j int) bool {
		n := strings.Compare(methods[i].Name, methods[j].Name)
		if n == 0 {
			panic(fmt.Sprintf("method redeclared: %v", methods[j].Name))
		}
		return n < 0
	})
	var ms []Method
	for _, m := range methods {
		if !m.Pointer {
			ms = append(ms, m)
		}
	}
	if EnableVerifyField && styp.Kind() == reflect.Struct {
		var fs []reflect.StructField
		for i := 0; i < styp.NumField(); i++ {
			fs = append(fs, styp.Field(i))
		}
		fs = append(fs, reflect.StructField{
			Name:    verifyFieldName,
			PkgPath: "main",
			Type:    verifyFieldType,
		})
		styp = NamedStructOf(styp.PkgPath(), styp.Name(), fs)
	}

	rt, typ := methodOf(styp, nil, ms)
	prt, _ := methodOf(reflect.PtrTo(styp), typ, methods)
	rt.ptrToThis = resolveReflectType(prt)
	(*ptrType)(unsafe.Pointer(prt)).elem = rt
	return typ
}

func methodOf(styp reflect.Type, elem reflect.Type, ms []Method) (*rtype, reflect.Type) {
	ptrto := styp.Kind() == reflect.Ptr
	var methods []method
	var exported int
	for _, m := range ms {
		ptr := tovalue(&m.Func).ptr
		isexport := EnableAllMethods || isExported(m.Name)
		methods = append(methods, method{
			name: resolveReflectName(newName(m.Name, "", isexport)),
			mtyp: resolveReflectType(totype(m.Type)),
			ifn:  resolveReflectText(unsafe.Pointer(ptr)),
			tfn:  resolveReflectText(unsafe.Pointer(ptr)),
		})
		if isexport {
			exported++
		}
	}
	if len(methods) == 0 {
		return totype(styp), styp
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
		st.fields = ost.fields
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
	ut.xcount = uint16(exported)
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

		// if ptrto && !m.Pointer {
		// 	in = append(in, typ.Elem())
		// } else {
		in = append(in, typ)
		// }

		for i := 0; i < mtyp.NumIn(); i++ {
			in = append(in, mtyp.In(i))
		}
		var out []reflect.Type
		for i := 0; i < mtyp.NumOut(); i++ {
			out = append(out, mtyp.Out(i))
		}
		// rewrite tfn
		ntyp := reflect.FuncOf(in, out, false)
		nindex := i
		if ptrto && !m.Pointer {
			mm, _ := elem.MethodByName(m.Name)
			nindex = mm.Index
			cv := reflect.MakeFunc(ntyp, func(args []reflect.Value) (results []reflect.Value) {
				return args[0].Elem().Method(nindex).Call(args[1:])
			})
			em[i].tfn = resolveReflectText(tovalue(&cv).ptr)
		} else {
			funcImpl := (*makeFuncImpl)(tovalue(&m.Func).ptr)
			funcImpl.ftyp = (*funcType)(unsafe.Pointer(totype(ntyp)))
		}

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
		_, ifn := icall(i, int(sz), len(out) > 0, ptrto)
		//log.Println("--->", i, index, int(sz), len(out), typ, ntyp, ms[i].Name, em[i].name, ifn, em[i].tfn)
		if ifn == nil {
			log.Printf("warning cannot wrapper method index:%v, size: %v\n", i, sz)
		} else {
			em[i].ifn = resolveReflectText(unsafe.Pointer(reflect.ValueOf(ifn).Pointer()))
		}
		infos = append(infos, &methodInfo{
			inTyp:    inTyp,
			outTyp:   outTyp,
			index:    nindex,
			pointer:  m.Pointer,
			variadic: m.Type.IsVariadic(),
		})
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
	inTyp    reflect.Type
	outTyp   reflect.Type
	index    int
	pointer  bool
	variadic bool
}

func MethodByType(typ reflect.Type, index int) reflect.Method {
	m := typ.Method(index)
	if _, ok := ntypeMap[typ]; ok {
		tovalue(&m.Func).flag |= flagIndir
	}
	return m
}

func MethodByName(typ reflect.Type, name string) (m reflect.Method, ok bool) {
	m, ok = typ.MethodByName(name)
	if !ok {
		return
	}
	if _, ok := ntypeMap[typ]; ok {
		tovalue(&m.Func).flag |= flagIndir
	}
	return
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
	if EnableVerifyField {
		if v.Kind() == reflect.Ptr {
			elem := v.Elem()
			if elem.Kind() == reflect.Struct {
				item := FieldByName(v.Elem(), verifyFieldName)
				if item.IsValid() {
					item.SetPointer(ptr)
				}
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
			if !EnableVerifyField {
				log.Printf("no verify, found type %v by %v\n", typ, ptr)
			}
			return typ
		}
	}
	return nil
}

func icall_x(i int, this uintptr, p []byte, ptrto bool) []byte {
	ptr := unsafe.Pointer(this)
	typ := foundTypeByPtr(ptr)
	if typ == nil {
		log.Println("cannot found ptr type", ptr)
		return nil
	}
	if ptrto {
		typ = reflect.PtrTo(typ)
	}
	infos, ok := typInfoMap[typ]
	if !ok {
		log.Println("cannot found type info", typ)
	}
	info := infos[i]
	log.Println("------->", typ, i, info.index)
	var method reflect.Method
	if ptrto && !info.pointer {
		method = MethodByType(typ.Elem(), info.index)
	} else {
		method = MethodByType(typ, info.index)
	}
	var in []reflect.Value
	var receiver reflect.Value
	if ptrto {
		receiver = reflect.NewAt(typ.Elem(), ptr)
		if !info.pointer {
			receiver = receiver.Elem()
		}
	} else {
		receiver = reflect.NewAt(typ, ptr).Elem()
	}
	in = append(in, receiver)
	inCount := method.Type.NumIn()
	if inCount > 1 {
		inArgs := reflect.NewAt(info.inTyp, unsafe.Pointer(&p[0])).Elem()
		if info.variadic {
			for i := 1; i < inCount-1; i++ {
				in = append(in, inArgs.Field(i-1))
			}
			slice := inArgs.Field(inCount - 2)
			for i := 0; i < slice.Len(); i++ {
				in = append(in, slice.Index(i))
			}
		} else {
			for i := 1; i < inCount; i++ {
				in = append(in, inArgs.Field(i-1))
			}
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
