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
	AddVerifyField  = true
	verifyFieldType = reflect.TypeOf(unsafe.Pointer(nil))
	verifyFieldName = "_reflectx_verify"
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
	var mcount, pcount int
	pcount = len(methods)
	var mlist []string
	for _, m := range methods {
		if !m.Pointer {
			mlist = append(mlist, m.Name)
			mcount++
		}
	}
	orgtyp := styp
	if AddVerifyField && styp.Kind() == reflect.Struct {
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
	rt, _ := premakeMethodType(styp, mcount, mcount)
	prt, _ := premakeMethodType(reflect.PtrTo(styp), pcount, pcount)
	rt.ptrToThis = resolveReflectType(prt)
	(*ptrType)(unsafe.Pointer(prt)).elem = rt
	typ := toType(rt)
	ptyp := reflect.PtrTo(typ)
	ms := rt.methods()
	pms := prt.methods()
	var infos []*methodInfo
	var pinfos []*methodInfo
	var index int
	for i, m := range methods {
		ptr := tovalue(&m.Func).ptr
		name := resolveReflectName(newName(m.Name, "", true))
		in, out, ntyp, inTyp, outTyp := toRealType(typ, orgtyp, m.Type)
		mtyp := resolveReflectType(totype(ntyp))
		var ftyp reflect.Type
		if m.Pointer {
			ftyp = reflect.FuncOf(append([]reflect.Type{ptyp}, in...), out, m.Type.IsVariadic())
		} else {
			ftyp = reflect.FuncOf(append([]reflect.Type{typ}, in...), out, m.Type.IsVariadic())
		}
		funcImpl := (*makeFuncImpl)(tovalue(&m.Func).ptr)
		funcImpl.ftyp = (*funcType)(unsafe.Pointer(totype(ftyp)))
		sz := totype(inTyp).size
		_, ifunc := icall(i, int(sz), m.Type.NumOut() > 0, true)
		var pifn, tfn, ptfn textOff
		if ifunc == nil {
			log.Printf("warning cannot wrapper method index:%v, size: %v\n", i, sz)
		} else {
			pifn = resolveReflectText(unsafe.Pointer(reflect.ValueOf(ifunc).Pointer()))
		}
		tfn = resolveReflectText(unsafe.Pointer(ptr))
		pindex := i
		if !m.Pointer {
			for i, s := range mlist {
				if s == m.Name {
					pindex = i
					break
				}
			}
			ctyp := reflect.FuncOf(append([]reflect.Type{ptyp}, in...), out, m.Type.IsVariadic())
			cv := reflect.MakeFunc(ctyp, func(args []reflect.Value) (results []reflect.Value) {
				return args[0].Elem().Method(pindex).Call(args[1:])
			})
			ptfn = resolveReflectText(tovalue(&cv).ptr)
		} else {
			ptfn = tfn
		}

		pms[i].name = name
		pms[i].mtyp = mtyp
		pms[i].tfn = ptfn
		pms[i].ifn = pifn
		pinfos = append(pinfos, &methodInfo{
			inTyp:    inTyp,
			outTyp:   outTyp,
			name:     m.Name,
			index:    pindex,
			pointer:  m.Pointer,
			variadic: m.Type.IsVariadic(),
		})
		if !m.Pointer {
			_, ifunc := icall(index, int(sz), m.Type.NumOut() > 0, false)
			var ifn textOff
			if ifunc == nil {
				log.Printf("warning cannot wrapper method index:%v, size: %v\n", i, sz)
			} else {
				ifn = resolveReflectText(unsafe.Pointer(reflect.ValueOf(ifunc).Pointer()))
			}
			ms[index].name = name
			ms[index].mtyp = mtyp
			ms[index].tfn = tfn
			ms[index].ifn = ifn
			infos = append(infos, &methodInfo{
				inTyp:    inTyp,
				outTyp:   outTyp,
				name:     m.Name,
				index:    index,
				pointer:  m.Pointer,
				variadic: m.Type.IsVariadic(),
			})
			index++
		}
	}
	typInfoMap[typ] = infos
	typInfoMap[ptyp] = pinfos
	nt := &Named{Name: styp.Name(), PkgPath: styp.PkgPath(), Type: typ, Kind: TkStruct}
	ntypeMap[typ] = nt
	return typ
}

func toRealType(typ, orgtyp, mtyp reflect.Type) (in, out []reflect.Type, ntyp, inTyp, outTyp reflect.Type) {
	fn := func(t reflect.Type) reflect.Type {
		if t == orgtyp {
			return typ
		} else if t.Kind() == reflect.Ptr && t.Elem() == orgtyp {
			return reflect.PtrTo(typ)
		}
		return t
	}
	var inFields []reflect.StructField
	var outFields []reflect.StructField
	for i := 0; i < mtyp.NumIn(); i++ {
		t := fn(mtyp.In(i))
		in = append(in, t)
		inFields = append(inFields, reflect.StructField{
			Name: fmt.Sprintf("Arg%v", i),
			Type: t,
		})
	}
	for i := 0; i < mtyp.NumOut(); i++ {
		t := fn(mtyp.Out(i))
		out = append(out, t)
		outFields = append(outFields, reflect.StructField{
			Name: fmt.Sprintf("Out%v", i),
			Type: t,
		})
	}
	ntyp = reflect.FuncOf(in, out, mtyp.IsVariadic())
	inTyp = reflect.StructOf(inFields)
	outTyp = reflect.StructOf(outFields)
	return
}

func premakeMethodType(styp reflect.Type, mcount int, xcount int) (rt *rtype, tt reflect.Value) {
	ort := totype(styp)
	switch styp.Kind() {
	case reflect.Struct:
		tt = reflect.New(reflect.StructOf([]reflect.StructField{
			{Name: "S", Type: reflect.TypeOf(structType{})},
			{Name: "U", Type: reflect.TypeOf(uncommonType{})},
			{Name: "M", Type: reflect.ArrayOf(mcount, reflect.TypeOf(method{}))},
		}))
		st := (*structType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
		ost := toStructType(ort)
		st.fields = ost.fields
		rt = (*rtype)(unsafe.Pointer(st))
	case reflect.Ptr:
		tt = reflect.New(reflect.StructOf([]reflect.StructField{
			{Name: "S", Type: reflect.TypeOf(ptrType{})},
			{Name: "U", Type: reflect.TypeOf(uncommonType{})},
			{Name: "M", Type: reflect.ArrayOf(mcount, reflect.TypeOf(method{}))},
		}))
		st := (*ptrType)(unsafe.Pointer(tt.Elem().Field(0).UnsafeAddr()))
		rt = (*rtype)(unsafe.Pointer(st))
	}
	ut := (*uncommonType)(unsafe.Pointer(tt.Elem().Field(1).UnsafeAddr()))
	// copy(tt.Elem().Field(2).Slice(0, len(methods)).Interface().([]method), methods)
	ut.mcount = uint16(mcount)
	ut.xcount = uint16(xcount)
	ut.moff = uint32(unsafe.Sizeof(uncommonType{}))

	rt.size = ort.size
	rt.tflag = ort.tflag | tflagUncommon
	rt.kind = ort.kind
	rt.align = ort.align
	rt.fieldAlign = ort.fieldAlign
	rt.str = resolveReflectName(ort.nameOff(ort.str))
	return
}

func methodOf(styp, orgtyp, elem reflect.Type, ms []Method) (*rtype, reflect.Type) {
	ptrto := styp.Kind() == reflect.Ptr
	var methods []method
	for _, m := range ms {
		ptr := tovalue(&m.Func).ptr
		methods = append(methods, method{
			name: resolveReflectName(newName(m.Name, "", true)),
			mtyp: resolveReflectType(totype(m.Type)),
			ifn:  resolveReflectText(unsafe.Pointer(ptr)),
			tfn:  resolveReflectText(unsafe.Pointer(ptr)),
		})
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
			t := mtyp.In(i)
			in = append(in, t)
		}
		var out []reflect.Type
		for i := 0; i < mtyp.NumOut(); i++ {
			t := mtyp.Out(i)
			out = append(out, t)
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
	name     string
	index    int
	pointer  bool
	variadic bool
}

func MethodByIndex(typ reflect.Type, index int) reflect.Method {
	m := typ.Method(index)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
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
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
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
	if AddVerifyField {
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
			if !AddVerifyField {
				log.Printf("no verify, found type %v by %v\n", typ, ptr)
			}
			return typ
		}
	}
	return nil
}

func i_x(i int, this uintptr, p []byte, ptrto bool) []byte {
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
	var method reflect.Method
	if ptrto && !info.pointer {
		method = MethodByIndex(typ.Elem(), info.index)
	} else {
		method = MethodByIndex(typ, info.index)
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
