//go:build amd64
// +build amd64

#include "textflag.h"

// func IntersectsAVX(a, b *[16]uint16) bool
TEXT ·IntersectsAVX(SB), NOSPLIT, $0-32
    // a → DI, b → SI
    MOVQ    a+0(FP), DI
    MOVQ    b+8(FP), SI

    // Загрузим 256 бит из [a] → Y0, из [b] → Y1
    VMOVDQU (DI), Y0
    VMOVDQU (SI), Y1

    // Побитовое AND: Y0 = Y0 & Y1
    VPAND   Y1, Y0, Y0

    // Тест всех битов: ZF=1 если Y0==0
    VPTEST  Y0, Y0

    // Обнуляем весь регистр-результат
    XORL    EAX, EAX
    // Если ZF==0 (т.е. какое‑то пересечение), ставим AL=1
    SETNE   AL

    // Сбрасываем AVX‑состояние
    VZEROUPPER

    RET
