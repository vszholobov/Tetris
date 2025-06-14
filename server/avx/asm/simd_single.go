package asm

//go:noescape
func IntersectsAVXSingle(a, b *[16]uint16) bool
