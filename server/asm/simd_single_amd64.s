//go:build amd64
// +build amd64

#include "textflag.h"

// func IntersectsAVX(a, b *[16]uint16) bool
TEXT ·IntersectsAVXSingle(SB), NOSPLIT, $0-32
    // Считаем a → DI, b → SI
    MOVQ    a+0(FP), DI
    MOVQ    b+8(FP), SI

    // Загружаем 256 бит
    VMOVDQU (DI), Y0
    VMOVDQU (SI), Y1

    // Побитовое AND: Y0 = Y0 & Y1
    VPAND   Y1, Y0, Y0

    // Проверка на ноль
    VPTEST  Y0, Y0

    // AL := (Y0 != 0)
    SETNE   R8B
    // Записываем результат в ret
    MOVB    R8B, ret+16(FP)

    // Очистка AVX
    VZEROUPPER

    RET
