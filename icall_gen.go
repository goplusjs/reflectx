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

func icall(i int, bytes int, ret bool) interface{} {
	if i > max_icall_index || bytes > max_icall_bytes {
		return nil
	}
	if ret {
		return icall_array[i+bytes*max_icall_index/64]
	} else {
		return icall_array_n[i+bytes*max_icall_index/64]
	}
}
`

var templ_0 = `	func(p uintptr) []byte { return icall_x($index, p, nil) },
`
var templ = `	func(p uintptr, a [$bytes]byte) []byte { return icall_x($index, p, a[:]) },
`
var templ_n_0 = `	func(p uintptr) { icall_x($index, p, nil) },
`
var templ_n = `	func(p uintptr, a [$bytes]byte) { icall_x($index, p, a[:]) },
`

const (
	max_index = 4
	max_bytes = 32
)

func main() {
	var buf bytes.Buffer
	buf.WriteString(head)
	buf.WriteString(fmt.Sprintf("\nconst max_icall_index = %v\n", max_index))
	buf.WriteString(fmt.Sprintf("const max_icall_bytes = %v\n", max_bytes))
	buf.WriteString("\nvar icall_array = []interface{}{\n")
	for i := 0; i <= max_index; i++ {
		for j := 0; j <= max_bytes; j += 8 {
			r := strings.NewReplacer("$index", strconv.Itoa(i), "$bytes", strconv.Itoa(j))
			if j == 0 {
				r.WriteString(&buf, templ_0)
			} else {
				r.WriteString(&buf, templ)
			}
		}
	}
	buf.WriteString("}\n")
	buf.WriteString("\nvar icall_array_n = []interface{}{\n")
	for i := 0; i <= max_index; i++ {
		for j := 0; j <= max_bytes; j += 8 {
			r := strings.NewReplacer("$index", strconv.Itoa(i), "$bytes", strconv.Itoa(j))
			if j == 0 {
				r.WriteString(&buf, templ_n_0)
			} else {
				r.WriteString(&buf, templ_n)
			}
		}
	}
	buf.WriteString("}\n")
	ioutil.WriteFile("./icall.go", buf.Bytes(), 0666)
}
