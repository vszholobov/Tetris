package main

import (
	"math/big"
	"math/rand"
	"testing"
	"time"
	"unsafe"

	"tetrisServer/asm"
)

//go:noinline
func processBigInt(data []*big.Int) int {
	sum := 0
	for j := 0; j < len(data)-1; j++ {
		var tmp big.Int
		tmp.And(data[j], data[j+1])
		if tmp.Sign() != 0 {
			sum++
		}
	}
	return sum
}

//go:noinline
func processGo(data [][]uint16) int {
	sum := 0
	for j := 0; j < len(data)-1; j++ {
		a, b := data[j], data[j+1]
		for i := 0; i < 16; i++ {
			if a[i]&b[i] != 0 {
				sum++
				break
			}
		}
	}
	return sum
}

func generateData(n int) [][]uint16 {
	rand.Seed(time.Now().UnixNano())
	data := make([][]uint16, n)
	for i := range data {
		arr := make([]uint16, 16)
		for j := range arr {
			arr[j] = uint16(rand.Intn(0xFFFF))
		}
		data[i] = arr
	}
	return data
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

func BenchmarkIntersectsBigIntLarge(b *testing.B) {
	data := generateData(1_000_000)
	bigData := make([]*big.Int, len(data))
	for i, arr := range data {
		bigData[i] = prepareBigIntFromSlice(arr)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := processBigInt(bigData)
		if sum == 0 {
			b.Fatalf("impossible")
		}
	}
}

func BenchmarkIntersectsGoLarge(b *testing.B) {
	data := generateData(1_000_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := processGo(data)
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
			a := (*[16]uint16)(unsafe.Pointer(&data[j][0]))
			c := (*[16]uint16)(unsafe.Pointer(&data[j+1][0]))
			if asm.IntersectsAVX(a, c) {
				sum++
			}
		}
		if sum == 0 {
			b.Fatalf("impossible")
		}
	}
}
