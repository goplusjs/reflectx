// +build ignore

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

var head = `package reflectx

func icall(i int, bytes int, ret bool, ptrto bool) interface{} {
	if i > max_icall_index || bytes > max_icall_bytes {
		return nil
	}
	index := bytes/8 + i*(max_icall_bytes/8+1)
	if ptrto {
		if ret {
			return icall_ptr[index]
		} else {
			return icall_ptr_n[index]
		}
	} else {
		if ret {
			return icall_struct[index]
		} else {
			return icall_struct_n[index]
		}
	}
}
`

var templ_0 = `	func(p uintptr) []byte { return icall_x($index, p, nil, $ptr) },
`
var templ = `	func(p uintptr, a [$bytes]byte) []byte { return icall_x($index, p, a[:], $ptr) },
`
var templ_n_0 = `	func(p uintptr) { icall_x($index, p, nil, $ptr) },
`
var templ_n = `	func(p uintptr, a [$bytes]byte) { icall_x($index, p, a[:], $ptr) },
`

const (
	max_index = 128
	max_bytes = 256
)

func main() {
	var buf bytes.Buffer
	buf.WriteString(head)
	buf.WriteString(fmt.Sprintf("\nconst max_icall_index = %v\n", max_index))
	buf.WriteString(fmt.Sprintf("const max_icall_bytes = %v\n", max_bytes))

	fnWrite := func(name string, t string, t0 string, ptr string) {
		buf.WriteString(fmt.Sprintf("\nvar %v = []interface{}{\n", name))
		for i := 0; i <= max_index; i++ {
			for j := 0; j <= max_bytes; j += 8 {
				r := strings.NewReplacer("$index", strconv.Itoa(i), "$bytes", strconv.Itoa(j), "$ptr", ptr)
				if j == 0 {
					r.WriteString(&buf, t)
				} else {
					r.WriteString(&buf, t0)
				}
			}
		}
		buf.WriteString("}\n")
	}
	fnWrite("icall_struct", templ, templ_0, "false")
	fnWrite("icall_struct_n", templ_0, templ_n, "false")
	fnWrite("icall_ptr", templ, templ_0, "true")
	fnWrite("icall_ptr_n", templ_0, templ_n, "true")

	ioutil.WriteFile("./icall.go", buf.Bytes(), 0666)
}
