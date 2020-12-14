// +build ignore

package main

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"strings"
)

var head = `package reflectx

`

var templ_0 = `func (w wrapper) I$index_$bytes() []byte {
	return w.call($index, nil)
}
func (w wrapper) N$index_$bytes() {
	w.call($index, nil)
}
`

var templ = `func (w wrapper) I$index_$bytes(p [$bytes]byte) []byte {
	return w.call($index, p[:])
}
func (w wrapper) N$index_$bytes(p [$bytes]byte) {
	w.call($index, p[:])
}
`

const (
	max_index = 128
	max_bytes = 128
)

func main() {
	var buf bytes.Buffer
	buf.WriteString(head)
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
	ioutil.WriteFile("./wrapper_call.go", buf.Bytes(), 0666)
}
