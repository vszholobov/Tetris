//go:build !amd64
// +build !amd64

package asm

func IntersectsAVXSingle(a, b *[16]uint16) bool {
	return false
}

func IntersectsAVXMultiple(a, b **[16]uint16, n int) int {
	return 0
}
