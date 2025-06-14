package main

import (
	"math/big"
	"math/rand"
	"testing"
	"time"
	"unsafe"

	"tetrisServer/asm"
)

// ваша C‑SIMD-функция
// //go:noescape
// func intersectsSimde(a, b []uint16) bool

// новая обёртка для asm
func intersectsAsm(a, b []uint16) bool {
	if len(a) < 16 || len(b) < 16 {
		panic("need at least 16 elements")
	}
	// приводим к массивам [16]uint16
	return asm.IntersectsAVX((*[16]uint16)(unsafe.Pointer(&a[0])), (*[16]uint16)(unsafe.Pointer(&b[0])))
}

func intersectsBigInt(a, b *big.Int) bool {
	var tmp big.Int
	tmp.And(a, b)
	return tmp.Sign() != 0
}

func intersectsGo(a, b []uint16) bool {
	for i := 0; i < 16; i++ {
		if a[i]&b[i] != 0 {
			return true
		}
	}
	return false
}

func prepareBigIntFromSlice(data []uint16) *big.Int {
	bytes := make([]byte, 0, 32)
	for i := len(data) - 1; i >= 0; i-- {
		bytes = append(bytes, byte(data[i]>>8), byte(data[i]&0xff))
	}
	b := new(big.Int)
	b.SetBytes(bytes)
	return b
}

func generateData(n int) [][]uint16 {
	rand.Seed(time.Now().UnixNano())
	data := make([][]uint16, n)
	for i := 0; i < n; i++ {
		arr := make([]uint16, 16)
		for j := 0; j < 16; j++ {
			arr[j] = uint16(rand.Intn(0xFFFF))
		}
		data[i] = arr
	}
	return data
}

func BenchmarkIntersectsBigIntLarge(b *testing.B) {
	data := generateData(1_000_000)
	bigData := make([]*big.Int, len(data))
	for i, arr := range data {
		bigData[i] = prepareBigIntFromSlice(arr)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < len(bigData)-1; j++ {
			if intersectsBigInt(bigData[j], bigData[j+1]) {
				sum++
			}
		}
		if sum == 0 {
			b.Fatalf("impossible")
		}
	}
}

func BenchmarkIntersectsGoLarge(b *testing.B) {
	data := generateData(1_000_000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < len(data)-1; j++ {
			if intersectsGo(data[j], data[j+1]) {
				sum++
			}
		}
		if sum == 0 {
			b.Fatalf("impossible")
		}
	}
}

func BenchmarkIntersectsSimdeLarge(b *testing.B) {
	data := generateData(1_000_000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < len(data)-1; j++ {
			if intersectsSimde(data[j], data[j+1]) {
				sum++
			}
		}
		if sum == 0 {
			b.Fatalf("impossible")
		}
	}
}

func BenchmarkIntersectsAsmLarge(b *testing.B) {
	data := generateData(1_000_000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < len(data)-1; j++ {
			if intersectsAsm(data[j], data[j+1]) {
				sum++
			}
		}
		if sum == 0 {
			b.Fatalf("impossible")
		}
	}
}
