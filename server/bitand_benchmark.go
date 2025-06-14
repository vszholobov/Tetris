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
func processUint16(data [][]uint16) int {
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
	for i := 0; i < n; i++ {
		arr := make([]uint16, 16)
		if i%2 == 1 {
			// копируем предыдущий элемент, чтобы гарантировать пересечение
			copy(arr, data[i-1])
		} else {
			for j := 0; j < 16; j++ {
				arr[j] = uint16(rand.Intn(0xFFFF))
			}
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

func BenchmarkIntersectsBigInt(b *testing.B) {
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

func BenchmarkIntersectsUint16(b *testing.B) {
	data := generateData(1_000_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := processUint16(data)
		if sum == 0 {
			b.Fatalf("impossible")
		}
	}
}

func BenchmarkIntersectsAsmAvxSingle(b *testing.B) {
	data := generateData(1_000_000)

	// Преобразуем в указатели *[16]uint16
	ptrs := make([]*[16]uint16, len(data))
	for i := range data {
		ptrs[i] = (*[16]uint16)(unsafe.Pointer(&data[i][0]))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < len(ptrs)-1; j++ {
			if asm.IntersectsAVXSingle(ptrs[j], ptrs[j+1]) {
				sum++
			}
		}
		if sum == 0 {
			b.Fatalf("impossible")
		}
	}
}

func BenchmarkIntersectsAsmAvxMany(b *testing.B) {
	data := generateData(1_000_000)

	aPtrs := make([]*[16]uint16, len(data)-1)
	bPtrs := make([]*[16]uint16, len(data)-1)

	for i := 0; i < len(data)-1; i++ {
		aPtrs[i] = (*[16]uint16)(unsafe.Pointer(&data[i][0]))
		bPtrs[i] = (*[16]uint16)(unsafe.Pointer(&data[i+1][0]))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := asm.IntersectsAVXMultiple(&aPtrs[0], &bPtrs[0], len(aPtrs))
		if sum == 0 {
			b.Fatalf("impossible")
		}
	}
}

func BenchmarkCSimdeSingle(b *testing.B) {
	data := generateData(1_000_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < len(data)-1; j++ {
			if intersectsSimdeSingle(data[j], data[j+1]) {
				sum++
			}
		}
		if sum == 0 {
			b.Fatalf("impossible")
		}
	}
}

func BenchmarkCSimdeMany(b *testing.B) {
	data := generateData(1_000_000)

	// Формируем два больших слайса uint16 для передачи в C,
	// размер: 16 * (len(data) - 1) — чтобы сравнивать пары
	aFlat := make([]uint16, 16*(len(data)-1))
	bFlat := make([]uint16, 16*(len(data)-1))

	for i := 0; i < len(data)-1; i++ {
		copy(aFlat[i*16:(i+1)*16], data[i])
		copy(bFlat[i*16:(i+1)*16], data[i+1])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := intersectsSimdeMany(aFlat, bFlat)
		if sum == 0 {
			b.Fatalf("impossible")
		}
	}
}
