package asm

// IntersectsAVXSingle принимает указатели на 16 uint16 (256 бит) и возвращает bool,
// true, если есть пересечение (AND != 0)
//
//go:noescape
func IntersectsAVXSingle(a, b *[16]uint16) bool
