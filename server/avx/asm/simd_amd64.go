//go:build amd64
// +build amd64

package asm

//go:noescape
func IntersectsAVXSingle(a, b *[16]uint16) bool

//go:noescape
func IntersectsAVXMultiple(a, b **[16]uint16, n int) int
