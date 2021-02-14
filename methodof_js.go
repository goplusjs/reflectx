// +build js,!wasm

package reflectx

import (
	"fmt"
	"log"
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
	return typ.Method(index)
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
	rt, ums := newType(styp, mcount, mcount)
	setTypeName(rt, styp.PkgPath(), styp.Name())
	typ := toType(rt)

	//var index int
	jstyp := jsType(rt)
	jstyp.Set("methodSetCache", nil)
	//jstyp.Set("exported", true)
	// jstyp.Set("named", true)
	jstyp.Set("string", styp.String())
	// jstyp.Set("reflectType", js.Undefined)

	jsms := jstyp.Get("methods")
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
	pjstyp.Set("methodSetCache", nil)
	pjsms := pjstyp.Get("methods")
	pjsproto := pjstyp.Get("prototype")

	ut := toUncommonType(prt)
	ut.mcount = uint16(pcount)
	ut.xcount = uint16(pcount)
	ut.moff = uint32(unsafe.Sizeof(uncommonType{}))
	ut._methods = make([]method, pcount, pcount)

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
		}
		ut._methods[i].name = resolveReflectName(newName(m.Name, "", true))
		ut._methods[i].mtyp = resolveReflectType(totype(ntyp))
		if !pointer {
			ums[index].name = resolveReflectName(newName(m.Name, "", true))
			ums[index].mtyp = resolveReflectType(totype(ntyp))
		}
		fnName := m.Name
		pjsproto.Set(m.Name, js.MakeFunc(func(this *js.Object, args []*js.Object) interface{} {
			log.Println("=======> pjs", fnName)
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
		jsproto.Set(js.InternalObject(m.Name).String(), js.MakeFunc(func(this *js.Object, args []*js.Object) interface{} {
			log.Println("=======> js", fnName, len(args))
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
	// t := reflect.TypeOf((*T)(nil)).Elem()
	// t0 := reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	// v0 := reflect.New(t).Elem()
	// v1 := reflect.New(typ).Elem()
	// log.Println("---> conv", t.Name(), v0.Convert(t0))
	// log.Println("---> conv", typ.Name(), v1.Convert(t0))

	return typ
}
