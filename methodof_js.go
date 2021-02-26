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

func methodOf(styp reflect.Type, methods []Method) reflect.Type {
	sort.Slice(methods, func(i, j int) bool {
		n := strings.Compare(methods[i].Name, methods[j].Name)
		if n == 0 && methods[i].Type == methods[j].Type {
			panic(fmt.Sprintf("method redeclared: %v", methods[j].Name))
		}
		return n < 0
	})
	isPointer := func(m Method) bool {
		return m.Pointer
	}
	var mcount, pcount int
	pcount = len(methods)
	for _, m := range methods {
		if !isPointer(m) {
			mcount++
		}
	}
	orgtyp := styp
	rt, ums := newType(styp.PkgPath(), styp.Name(), styp, mcount, mcount)
	setTypeName(rt, styp.PkgPath(), styp.Name())

	typ := toType(rt)
	jstyp := jsType(rt)
	jstyp.Set("methodSetCache", nil)
	jsms := jstyp.Get("methods")
	jsproto := jstyp.Get("prototype")
	jsmscache := js.Global.Get("Array").New()

	ptyp := reflect.PtrTo(typ)
	prt := totype(ptyp)
	pums := resetUncommonType(prt, pcount, pcount)._methods
	pjstyp := jsType(prt)
	pjstyp.Set("methodSetCache", nil)
	pjsms := pjstyp.Get("methods")
	pjsproto := pjstyp.Get("prototype")
	pjsmscache := js.Global.Get("Array").New()

	index := -1
	pindex := -1
	for i, m := range methods {
		in, out, ntyp, _, _ := toRealType(typ, orgtyp, m.Type)
		var ftyp reflect.Type
		if m.Pointer {
			ftyp = reflect.FuncOf(append([]reflect.Type{ptyp}, in...), out, m.Type.IsVariadic())
			pindex++
		} else {
			ftyp = reflect.FuncOf(append([]reflect.Type{typ}, in...), out, m.Type.IsVariadic())
			index++
		}
		fn := js.Global.Get("Object").New()
		fn.Set("pkg", "")
		fn.Set("name", js.InternalObject(m.Name))
		fn.Set("prop", js.InternalObject(m.Name))
		fn.Set("typ", jsType(totype(ntyp)))
		if m.Pointer {
			pjsms.SetIndex(pindex, fn)
		} else {
			jsms.SetIndex(index, fn)
			jsmscache.SetIndex(index, fn)
		}
		pjsmscache.SetIndex(i, fn)

		mname := resolveReflectName(newName(m.Name, "", true))
		mtyp := resolveReflectType(totype(ntyp))
		pums[i].name = mname
		pums[i].mtyp = mtyp
		if !m.Pointer {
			ums[index].name = mname
			ums[index].mtyp = mtyp
		}
		dfn := reflect.MakeFunc(ftyp, m.Func)
		tfn := tovalue(&dfn)
		nargs := ftyp.NumIn()
		if m.Pointer {
			pjsproto.Set(m.Name, js.MakeFunc(func(this *js.Object, args []*js.Object) interface{} {
				iargs := make([]interface{}, nargs, nargs)
				iargs[0] = this
				for i, arg := range args {
					iargs[i+1] = arg
				}
				return js.InternalObject(tfn.ptr).Invoke(iargs...)
			}))
		} else {
			pjsproto.Set(m.Name, js.MakeFunc(func(this *js.Object, args []*js.Object) interface{} {
				iargs := make([]interface{}, nargs, nargs)
				iargs[0] = *(**js.Object)(unsafe.Pointer(this))
				for i, arg := range args {
					iargs[i+1] = arg
				}
				return js.InternalObject(tfn.ptr).Invoke(iargs...)
			}))
		}
		jsproto.Set(m.Name, js.MakeFunc(func(this *js.Object, args []*js.Object) interface{} {
			iargs := make([]interface{}, nargs, nargs)
			iargs[0] = this.Get("$val")
			for i, arg := range args {
				iargs[i+1] = arg
			}
			return js.InternalObject(tfn.ptr).Invoke(iargs...)
		}))
	}
	jstyp.Set("methodSetCache", jsmscache)
	pjstyp.Set("methodSetCache", pjsmscache)

	return typ
}
