package main

/*
#cgo CFLAGS: -I./simde/simde -O3 -mavx2
#cgo LDFLAGS:
#include <stdbool.h>
#include <stdint.h>

bool intersects_simde_single(const uint16_t* a, const uint16_t* b);
int intersects_simde_many(const uint16_t* a, const uint16_t* b, size_t n);
*/
import "C"
import (
	"unsafe"
)

func intersectsSimdeSingle(a, b []uint16) bool {
	if len(a) < 16 || len(b) < 16 {
		panic("Arrays must be at least 16 elements")
	}
	// Передаем указатели на элементы
	ret := C.intersects_simde_single(
		(*C.uint16_t)(unsafe.Pointer(&a[0])),
		(*C.uint16_t)(unsafe.Pointer(&b[0])),
	)
	return bool(ret)
}

func intersectsSimdeMany(a, b []uint16) int {
	if len(a) < 16 || len(b) < 16 || len(a) != len(b) {
		panic("Arrays must be of equal length and at least 16 elements")
	}
	n := len(a) / 16
	return int(C.intersects_simde_many(
		(*C.uint16_t)(unsafe.Pointer(&a[0])),
		(*C.uint16_t)(unsafe.Pointer(&b[0])),
		C.size_t(n),
	))
}
