//go:build amd64
// +build amd64

#include "textflag.h"

// func IntersectsAVX(a, b *[16]uint16) bool
TEXT ·IntersectsAVX(SB), NOSPLIT, $0-32
    // a → DI, b → SI
    MOVQ    a+0(FP), DI
    MOVQ    b+8(FP), SI

    // Загрузим 256 бит
    VMOVDQU (DI), Y0
    VMOVDQU (SI), Y1

    // Побитовое AND
    VPAND   Y1, Y0, Y0

    // Тест всех битов: ZF=1, если Y0==0
    VPTEST  Y0, Y0

    // Обнуляем весь низ RAX (через 32‑битный XOR)
    XORL    AX, AX
    // Если ZF==0 (есть хотя бы один бит) → AL=1
    SETNE   AL

    // Сброс AVX-контекста
    VZEROUPPER

    RET
