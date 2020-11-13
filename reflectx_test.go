package reflectx_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/goplusjs/reflectx"
)

type Point struct {
	x int
	y int
}

func TestFieldCanSet(t *testing.T) {
	x := &Point{10, 20}
	v := reflect.ValueOf(x).Elem()

	sf := v.Field(0)
	if sf.CanSet() {
		t.Fatal("x unexport cannot set")
	}

	sf = reflectx.CanSet(sf)
	if !sf.CanSet() {
		t.Fatal("CanSet failed")
	}

	sf.Set(reflect.ValueOf(201))
	if x.x != 201 {
		t.Fatalf("x value %v", x.x)
	}
	sf.SetInt(202)
	if x.x != 202 {
		t.Fatalf("x value %v", x.x)
	}
}

type Buffer struct {
	*bytes.Buffer
	size  int
	value reflect.Value
	*bytes.Reader
}

func TestStructOf(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Failed()
		} else {
			t.Log("reflect.StructOf panic", v)
		}
	}()
	typ := reflect.TypeOf((*Buffer)(nil)).Elem()
	var fs []reflect.StructField
	for i := 0; i < typ.NumField(); i++ {
		fs = append(fs, typ.Field(i))
	}
	reflect.StructOf(fs)
}

func TestStructOfX(t *testing.T) {
	defer func() {
		v := recover()
		if v != nil {
			t.Fatalf("reflectx.StructOf %v", v)
		}
	}()
	typ := reflect.TypeOf((*Buffer)(nil)).Elem()
	var fs []reflect.StructField
	for i := 0; i < typ.NumField(); i++ {
		fs = append(fs, typ.Field(i))
	}
	dst := reflectx.StructOf(fs)
	for i := 0; i < dst.NumField(); i++ {
		if dst.Field(i).Anonymous != fs[i].Anonymous {
			t.Errorf("error field %v", dst.Field(i))
		}
	}

	v := reflect.New(dst)
	v.Elem().Field(0).Set(reflect.ValueOf(bytes.NewBufferString("hello")))
	reflectx.CanSet(v.Elem().Field(1)).SetInt(100)
	t.Log(v.Interface())
}
