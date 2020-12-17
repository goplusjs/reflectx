package reflectx

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

const max_icall_index = 4
const max_icall_bytes = 32

var icall_array = []interface{}{
	func(p uintptr) []byte { return icall_x(0, p, nil) },
	func(p uintptr, a [8]byte) []byte { return icall_x(0, p, a[:]) },
	func(p uintptr, a [16]byte) []byte { return icall_x(0, p, a[:]) },
	func(p uintptr, a [24]byte) []byte { return icall_x(0, p, a[:]) },
	func(p uintptr, a [32]byte) []byte { return icall_x(0, p, a[:]) },
	func(p uintptr) []byte { return icall_x(1, p, nil) },
	func(p uintptr, a [8]byte) []byte { return icall_x(1, p, a[:]) },
	func(p uintptr, a [16]byte) []byte { return icall_x(1, p, a[:]) },
	func(p uintptr, a [24]byte) []byte { return icall_x(1, p, a[:]) },
	func(p uintptr, a [32]byte) []byte { return icall_x(1, p, a[:]) },
	func(p uintptr) []byte { return icall_x(2, p, nil) },
	func(p uintptr, a [8]byte) []byte { return icall_x(2, p, a[:]) },
	func(p uintptr, a [16]byte) []byte { return icall_x(2, p, a[:]) },
	func(p uintptr, a [24]byte) []byte { return icall_x(2, p, a[:]) },
	func(p uintptr, a [32]byte) []byte { return icall_x(2, p, a[:]) },
	func(p uintptr) []byte { return icall_x(3, p, nil) },
	func(p uintptr, a [8]byte) []byte { return icall_x(3, p, a[:]) },
	func(p uintptr, a [16]byte) []byte { return icall_x(3, p, a[:]) },
	func(p uintptr, a [24]byte) []byte { return icall_x(3, p, a[:]) },
	func(p uintptr, a [32]byte) []byte { return icall_x(3, p, a[:]) },
	func(p uintptr) []byte { return icall_x(4, p, nil) },
	func(p uintptr, a [8]byte) []byte { return icall_x(4, p, a[:]) },
	func(p uintptr, a [16]byte) []byte { return icall_x(4, p, a[:]) },
	func(p uintptr, a [24]byte) []byte { return icall_x(4, p, a[:]) },
	func(p uintptr, a [32]byte) []byte { return icall_x(4, p, a[:]) },
}

var icall_array_n = []interface{}{
	func(p uintptr) { icall_x(0, p, nil) },
	func(p uintptr, a [8]byte) { icall_x(0, p, a[:]) },
	func(p uintptr, a [16]byte) { icall_x(0, p, a[:]) },
	func(p uintptr, a [24]byte) { icall_x(0, p, a[:]) },
	func(p uintptr, a [32]byte) { icall_x(0, p, a[:]) },
	func(p uintptr) { icall_x(1, p, nil) },
	func(p uintptr, a [8]byte) { icall_x(1, p, a[:]) },
	func(p uintptr, a [16]byte) { icall_x(1, p, a[:]) },
	func(p uintptr, a [24]byte) { icall_x(1, p, a[:]) },
	func(p uintptr, a [32]byte) { icall_x(1, p, a[:]) },
	func(p uintptr) { icall_x(2, p, nil) },
	func(p uintptr, a [8]byte) { icall_x(2, p, a[:]) },
	func(p uintptr, a [16]byte) { icall_x(2, p, a[:]) },
	func(p uintptr, a [24]byte) { icall_x(2, p, a[:]) },
	func(p uintptr, a [32]byte) { icall_x(2, p, a[:]) },
	func(p uintptr) { icall_x(3, p, nil) },
	func(p uintptr, a [8]byte) { icall_x(3, p, a[:]) },
	func(p uintptr, a [16]byte) { icall_x(3, p, a[:]) },
	func(p uintptr, a [24]byte) { icall_x(3, p, a[:]) },
	func(p uintptr, a [32]byte) { icall_x(3, p, a[:]) },
	func(p uintptr) { icall_x(4, p, nil) },
	func(p uintptr, a [8]byte) { icall_x(4, p, a[:]) },
	func(p uintptr, a [16]byte) { icall_x(4, p, a[:]) },
	func(p uintptr, a [24]byte) { icall_x(4, p, a[:]) },
	func(p uintptr, a [32]byte) { icall_x(4, p, a[:]) },
}
