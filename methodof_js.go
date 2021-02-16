// +build js,!wasm

package reflectx

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unsafe"

	"github.com/gopherjs/gopherjs/js"
)

func New(typ reflect.Type) reflect.Value {
	return reflect.New(typ)
}

func Interface(v reflect.Value) interface{} {
	return v.Interface()
}

func MethodByIndex(typ reflect.Type, index int) reflect.Method {
	m := typ.Method(index)
	m.Func = reflect.MakeFunc(m.Type, func(args []reflect.Value) []reflect.Value {
		recv := args[0].MethodByName(m.Name)
		if m.Type.IsVariadic() {
			return recv.CallSlice(args[1:])
		} else {
			return recv.Call(args[1:])
		}
	})
	return m
}

func MethodByName(typ reflect.Type, name string) (m reflect.Method, ok bool) {
	m, ok = typ.MethodByName(name)
	if !ok {
		return
	}
	m.Func = reflect.MakeFunc(m.Type, func(args []reflect.Value) []reflect.Value {
		recv := args[0].MethodByName(name)
		if m.Type.IsVariadic() {
			return recv.CallSlice(args[1:])
		} else {
			return recv.Call(args[1:])
		}
	})
	return
}

func jsFuncOf(in, out []reflect.Type, variadic bool) *js.Object {
	if variadic && (len(in) == 0 || in[len(in)-1].Kind() != reflect.Slice) {
		panic("reflect.FuncOf: last arg of variadic func must be slice")
	}

	jsIn := make([]*js.Object, len(in))
	for i, v := range in {
		jsIn[i] = jsType(totype(v))
	}
	jsOut := make([]*js.Object, len(out))
	for i, v := range out {
		jsOut[i] = jsType(totype(v))
	}
	return js.Global.Call("$funcType", jsIn, jsOut, variadic)
}

func methodOf(styp reflect.Type, methods []reflect.Method) reflect.Type {
	sort.Slice(methods, func(i, j int) bool {
		n := strings.Compare(methods[i].Name, methods[j].Name)
		if n == 0 && methods[i].Type == methods[j].Type {
			panic(fmt.Sprintf("method redeclared: %v", methods[j].Name))
		}
		return n < 0
	})
	isPointer := func(m reflect.Method) bool {
		return m.Type.In(0).Kind() == reflect.Ptr
	}
	var mcount, pcount int
	pcount = len(methods)
	for _, m := range methods {
		if !isPointer(m) {
			mcount++
		}
	}
	_ = pcount
	orgtyp := styp
	rt, ums := newType(styp.PkgPath(), styp.Name(), styp, mcount, mcount)
	setTypeName(rt, styp.PkgPath(), styp.Name())
	typ := toType(rt)

	//var index int
	jstyp := jsType(rt)
	jstyp.Set("methodSetCache", nil)
	jsmscache := js.Global.Get("Array").New()
	pjsmscache := js.Global.Get("Array").New()
	//jstyp.Set("exported", true)
	// jstyp.Set("named", true)
	//jstyp.Set("string", styp.String())
	// jstyp.Set("reflectType", js.Undefined)

	jsms := jstyp.Get("methods")
	//jsms := js.Global.Get("Array").New()
	//jstyp.Set("methods", jsms)
	jsproto := jstyp.Get("prototype")

	// enable uncommonType ptrTo
	pjstyp := js.Global.Call("$ptrType", jstyp)
	pjstyp.Set("named", true)
	//pjstyp.Set("exported", true)
	//pjstyp.Set("string", "*reflectx.X")
	// pjstyp.Set("reflectType", js.Undefined)

	//jstyp.Set("string", "reflectx.T")
	ptyp := reflect.PtrTo(typ)
	prt := totype(ptyp)
	//pjstyp := jsType(prt)
	pjstyp.Set("methodSetCache", nil)
	pjsms := pjstyp.Get("methods")
	//pjsms := js.Global.Get("Array").New()
	//pjstyp.Set("methods", pjsms)
	pjsproto := pjstyp.Get("prototype")

	ut := newUncommonType(pcount, pcount)
	setUncommonType(prt, ut)

	index := -1
	pindex := -1
	for i, m := range methods {
		in, out, ntyp, _, _ := toRealType(typ, orgtyp, m.Type)
		pointer := isPointer(m)
		var ftyp reflect.Type
		if pointer {
			ftyp = reflect.FuncOf(append([]reflect.Type{ptyp}, in...), out, m.Type.IsVariadic())
			pindex++
		} else {
			ftyp = reflect.FuncOf(append([]reflect.Type{typ}, in...), out, m.Type.IsVariadic())
			index++
		}
		_ = ftyp
		fn := js.Global.Get("Object").New()
		fn.Set("pkg", "")
		fn.Set("name", js.InternalObject(m.Name))
		fn.Set("prop", js.InternalObject(m.Name))
		fn.Set("typ", jsType(totype(ntyp)))
		_in := []*rtype{rt}
		_pin := []*rtype{prt}
		_out := []*rtype{}
		for _, t := range in {
			_pin = append(_pin, totype(t))
			_in = append(_in, totype(t))
		}
		for _, t := range out {
			_out = append(_out, totype(t))
		}
		tfn := tovalue(&m.Func)
		fnTyp := (*jsFuncType)(getKindType(tfn.typ))
		fnTyp._in = _in[1:]
		fnTyp._out = _out
		if pointer {
			pjsms.SetIndex(pindex, fn)
		} else {
			jsms.SetIndex(index, fn)
			jsmscache.SetIndex(index, fn)
		}
		pjsmscache.SetIndex(i, fn)

		mname := resolveReflectName(newName(m.Name, "", true))
		mtyp := resolveReflectType(totype(ntyp))
		ut._methods[i].name = mname
		ut._methods[i].mtyp = mtyp
		if !pointer {
			ums[index].name = mname
			ums[index].mtyp = mtyp
		}
		// fnName := m.Name
		pjsproto.Set(m.Name, js.MakeFunc(func(this *js.Object, args []*js.Object) interface{} {
			if pointer {
				fnTyp._in = _pin
			} else {
				this = *(**js.Object)(unsafe.Pointer(this))
				fnTyp._in = _in
			}
			fnTyp._out = _out
			fnTyp.inCount++
			defer func() {
				fnTyp.inCount--
				fnTyp._in = _in[1:]
			}()
			iargs := make([]interface{}, len(_in), len(_in))
			iargs[0] = this
			for i, arg := range args {
				iargs[i+1] = arg
			}
			return js.InternalObject(tfn.ptr).Invoke(iargs...)
		}))
		jsproto.Set(m.Name, js.MakeFunc(func(this *js.Object, args []*js.Object) interface{} {
			// log.Println("=======> js", fnName, len(args))
			fnTyp._in = _in
			fnTyp.inCount++
			defer func() {
				fnTyp.inCount--
				fnTyp._in = _in[1:]
			}()
			iargs := make([]interface{}, len(_in), len(_in))
			iargs[0] = this.Get("$val")
			for i, arg := range args {
				iargs[i+1] = arg
			}
			return js.InternalObject(tfn.ptr).Invoke(iargs...)
		}))
	}
	jstyp.Set("methodSetCache", jsmscache)
	// jstyp.Set("reflectType", js.Undefined)
	pjstyp.Set("methodSetCache", pjsmscache)
	// pjstyp.Set("reflectType", js.Undefined)
	//typ = toType(reflectType(jstyp))
	// ptyp = reflect.PtrTo(typ)

	// typ = toType(reflectType(jstyp))
	// ptyp = toType(reflectType(pjstyp))
	//typ2 := toType(reflectType(jstyp))
	//log.Println("---->", jsmscache.Length(), typ2.NumMethod())

	return typ
}
