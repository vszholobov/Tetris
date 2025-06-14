package main

/*
#cgo CFLAGS: -I./simde/simde -O3 -mavx2
#cgo LDFLAGS:
#include <stdbool.h>
#include <stdint.h>

bool intersects_simde(const uint16_t* a, const uint16_t* b);
*/
import "C"
import (
	"unsafe"
)

func intersectsSimde(a, b []uint16) bool {
	if len(a) < 16 || len(b) < 16 {
		panic("Arrays must be at least 16 elements")
	}
	// Передаем указатели на элементы
	ret := C.intersects_simde(
		(*C.uint16_t)(unsafe.Pointer(&a[0])),
		(*C.uint16_t)(unsafe.Pointer(&b[0])),
	)
	return bool(ret)
}
