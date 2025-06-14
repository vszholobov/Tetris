package asm

//go:noescape
func IntersectsAVXMultiple(a, b **[16]uint16, n int) int
