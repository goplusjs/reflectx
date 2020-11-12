package reflectx_test

import (
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
