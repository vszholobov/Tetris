//go:build amd64
// +build amd64

#include "textflag.h"

// func IntersectsAVX(a, b *[16]uint16) bool
TEXT ·IntersectsAVX(SB), NOSPLIT, $0-32
    // Загрузим указатели
    MOVQ    a+0(FP), DI
    MOVQ    b+8(FP), SI

    // Загрузим 256 бит
    VMOVDQU (DI), Y0
    VMOVDQU (SI), Y1

    // Побитовое AND
    VPAND   Y1, Y0, Y0

    // Тест на ноль (ZF=1 если всё нули)
    VPTEST  Y0, Y0

    // Установим AL=1, если ZF=0 (т.е. есть ненулевой бит)
    SETNE   AL

    // Очистим верхние полубайты для безопасности
    VZEROUPPER

    RET
