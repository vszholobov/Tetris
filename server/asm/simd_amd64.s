//go:build amd64
// +build amd64

#include "textflag.h"

// func IntersectsAVX(a, b *[16]uint16) bool
TEXT ·IntersectsAVX(SB), NOSPLIT, $0-32
    // a → DI, b → SI
    MOVQ    a+0(FP), DI
    MOVQ    b+8(FP), SI

    // ОБНУЛЯЕМ целиком Y0 и Y1
    VPXOR   Y0, Y0, Y0
    VPXOR   Y1, Y1, Y1

    // Загружаем 256 бит из [a] → Y0, из [b] → Y1
    VMOVDQU (DI), Y0
    VMOVDQU (SI), Y1

    // Побитовое AND: Y0 = Y0 & Y1
    VPAND   Y1, Y0, Y0

    // VPTEST выставляет ZF=1, если Y0==0
    VPTEST  Y0, Y0

    // AL = (ZF == 0) → 1, если есть хотя бы один бит
    SETNE   AL

    // Сбрасываем AVX‑состояние для возвращения в SSE‑режим
    VZEROUPPER

    RET
