package reflectx

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

// MakeMethod make reflect.Method for MethodOf
// - name: method name
// - pointer: flag receiver struct or pointer
// - typ: method func type without receiver
// - fn: func with receiver as first argument
func MakeMethod(name string, pointer bool, typ reflect.Type, fn func(args []reflect.Value) (result []reflect.Value)) reflect.Method {
	var in []reflect.Type
	var out []reflect.Type
	if pointer {
		in = append(in, tyEmptyInterfacePtr)
	} else {
		in = append(in, tyEmptyInterface)
	}
	for i := 0; i < typ.NumIn(); i++ {
		in = append(in, typ.In(i))
	}
	for i := 0; i < typ.NumOut(); i++ {
		out = append(out, typ.Out(i))
	}
	return reflect.Method{
		Name: name,
		Type: reflect.FuncOf(in, out, typ.IsVariadic()),
		Func: reflect.MakeFunc(typ, fn),
	}
}

func extraFieldMethod(ifield int, typ reflect.Type, skip map[string]bool) (methods []reflect.Method) {
	isPtr := typ.Kind() == reflect.Ptr
	for i := 0; i < typ.NumMethod(); i++ {
		m := MethodByIndex(typ, i)
		if skip[m.Name] {
			continue
		}
		var fn func(args []reflect.Value) []reflect.Value
		if isPtr {
			fn = func(args []reflect.Value) []reflect.Value {
				args[0] = args[0].Elem().Field(ifield).Addr()
				return m.Func.Call(args)
			}
		} else {
			fn = func(args []reflect.Value) []reflect.Value {
				args[0] = args[0].Field(ifield)
				return m.Func.Call(args)
			}
		}
		methods = append(methods, reflect.Method{
			Name:    m.Name,
			PkgPath: m.PkgPath,
			Type:    m.Type,
			Func:    reflect.MakeFunc(m.Type, fn),
		})
	}
	return
}

func parserFuncIO(typ reflect.Type) (in, out []reflect.Type) {
	for i := 0; i < typ.NumIn(); i++ {
		in = append(in, typ.In(i))
	}
	for i := 0; i < typ.NumOut(); i++ {
		out = append(out, typ.Out(i))
	}
	return
}

func extraPtrFieldMethod(ifield int, typ reflect.Type) (methods []reflect.Method) {
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		in, out := parserFuncIO(m.Type)
		in[0] = tyEmptyInterface
		mtyp := reflect.FuncOf(in, out, m.Type.IsVariadic())
		imethod := i
		methods = append(methods, reflect.Method{
			Name:    m.Name,
			PkgPath: m.PkgPath,
			Type:    mtyp,
			Func: reflect.MakeFunc(
				mtyp,
				func(args []reflect.Value) []reflect.Value {
					var recv = args[0]
					return recv.Field(ifield).Method(imethod).Call(args[1:])
				},
			),
		})
	}
	return
}

func extraInterfaceFieldMethod(ifield int, typ reflect.Type) (methods []reflect.Method) {
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		in, out := parserFuncIO(m.Type)
		in = append([]reflect.Type{tyEmptyInterface}, in...)
		mtyp := reflect.FuncOf(in, out, m.Type.IsVariadic())
		imethod := i
		methods = append(methods, reflect.Method{
			Name:    m.Name,
			PkgPath: m.PkgPath,
			Type:    mtyp,
			Func: reflect.MakeFunc(
				mtyp,
				func(args []reflect.Value) []reflect.Value {
					var recv = args[0]
					return recv.Field(ifield).Method(imethod).Call(args[1:])
				},
			),
		})
	}
	return
}

func extractEmbedMethod(styp reflect.Type) []reflect.Method {
	var methods []reflect.Method
	for i := 0; i < styp.NumField(); i++ {
		sf := styp.Field(i)
		if !sf.Anonymous {
			continue
		}
		switch sf.Type.Kind() {
		case reflect.Interface:
			ms := extraInterfaceFieldMethod(i, sf.Type)
			methods = append(methods, ms...)
		case reflect.Ptr:
			ms := extraPtrFieldMethod(i, sf.Type)
			methods = append(methods, ms...)
		default:
			skip := make(map[string]bool)
			ms := extraFieldMethod(i, sf.Type, skip)
			for _, m := range ms {
				skip[m.Name] = true
			}
			pms := extraFieldMethod(i, reflect.PtrTo(sf.Type), skip)
			methods = append(methods, ms...)
			methods = append(methods, pms...)
		}
	}
	// ambiguous selector check
	chk := make(map[string]int)
	for _, m := range methods {
		chk[m.Name]++
	}
	var ms []reflect.Method
	for _, m := range methods {
		if chk[m.Name] == 1 {
			ms = append(ms, m)
		}
	}
	return ms
}

func MethodOf(styp reflect.Type, methods []reflect.Method) reflect.Type {
	chk := make(map[string]int)
	for _, m := range methods {
		chk[m.Name]++
		if chk[m.Name] > 1 {
			panic(fmt.Sprintf("method redeclared: %v", m.Name))
		}
	}
	if styp.Kind() == reflect.Struct {
		ms := extractEmbedMethod(styp)
		for _, m := range ms {
			if chk[m.Name] == 1 {
				continue
			}
			methods = append(methods, m)
		}
	}
	return methodOf(styp, methods)
}

func MakeEmptyInterface(pkgpath string, name string) reflect.Type {
	return NamedTypeOf(pkgpath, name, tyEmptyInterface)
}

func NamedInterfaceOf(pkgpath string, name string, embedded []reflect.Type, methods []reflect.Method) reflect.Type {
	styp := NamedTypeOf(pkgpath, name, tyEmptyInterface)
	return InterfaceOf(styp, embedded, methods)
}

func InterfaceOf(styp reflect.Type, embedded []reflect.Type, methods []reflect.Method) reflect.Type {
	if styp.Kind() != reflect.Interface {
		panic(fmt.Errorf("non-interface %v", styp))
	}
	for _, e := range embedded {
		if e.Kind() != reflect.Interface {
			panic(fmt.Errorf("interface contains embedded non-interface %v", e))
		}
		for i := 0; i < e.NumMethod(); i++ {
			m := e.Method(i)
			methods = append(methods, reflect.Method{
				Name: m.Name,
				Type: m.Type,
			})
		}
	}
	sort.Slice(methods, func(i, j int) bool {
		n := strings.Compare(methods[i].Name, methods[j].Name)
		if n == 0 && methods[i].Type != methods[j].Type {
			panic(fmt.Sprintf("duplicate method %v", methods[j].Name))
		}
		return n < 0
	})
	rt, _ := newType(styp, 0, 0)
	st := (*interfaceType)(unsafe.Pointer(rt))
	st.methods = nil
	var lastname string
	for _, m := range methods {
		if m.Name == lastname {
			continue
		}
		lastname = m.Name
		st.methods = append(st.methods, imethod{
			name: resolveReflectName(newName(m.Name, "", isExported(m.Name))),
			typ:  resolveReflectType(totype(m.Type)),
		})
	}
	return toType(rt)
}

func toElem(typ reflect.Type) reflect.Type {
	if typ.Kind() == reflect.Ptr {
		return typ.Elem()
	}
	return typ
}

func toElemValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}

func toRealType(typ, orgtyp, mtyp reflect.Type) (in, out []reflect.Type, ntyp, inTyp, outTyp reflect.Type) {
	var fnx func(t reflect.Type) (reflect.Type, bool)
	fnx = func(t reflect.Type) (reflect.Type, bool) {
		if t == orgtyp {
			return typ, true
		}
		switch t.Kind() {
		case reflect.Ptr:
			if e, ok := fnx(t.Elem()); ok {
				return reflect.PtrTo(e), true
			}
		case reflect.Slice:
			if e, ok := fnx(t.Elem()); ok {
				return reflect.SliceOf(e), true
			}
		case reflect.Array:
			if e, ok := fnx(t.Elem()); ok {
				return reflect.ArrayOf(t.Len(), e), true
			}
		case reflect.Map:
			k, ok1 := fnx(t.Key())
			v, ok2 := fnx(t.Elem())
			if ok1 || ok2 {
				return reflect.MapOf(k, v), true
			}
		}
		return t, false
	}
	fn := func(t reflect.Type) reflect.Type {
		if r, ok := fnx(t); ok {
			return r
		}
		return t
	}
	var inFields []reflect.StructField
	var outFields []reflect.StructField
	for i := 1; i < mtyp.NumIn(); i++ {
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
