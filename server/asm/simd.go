package asm

// IntersectsAVX принимает указатели на 16 uint16 (256 бит) и возвращает bool,
// true, если есть пересечение (AND != 0)
//
//go:noescape
func IntersectsAVX(a, b *[16]uint16) bool
