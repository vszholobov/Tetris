//go:build amd64
// +build amd64

#include "textflag.h"

// func IntersectsAVX(a, b *[16]uint16) bool
TEXT ·IntersectsAVX(SB), NOSPLIT, $0-32
    // Загрузим указатели из стек‑фрейма (Go ABI):
    MOVQ    a+0(FP), DI    // первый аргумент -> DI
    MOVQ    b+8(FP), SI    // второй аргумент -> SI

    // Теперь DI и SI указывают на [16]uint16
    // Загрузим 256 бит (32 байта) из a и b:
    VMOVDQU (DI), Y0       // load unaligned 256-bit from *a
    VMOVDQU (SI), Y1       // load unaligned 256-bit from *b

    // Побитовое AND:
    VPAND   Y1, Y0, Y0

    // Тест на ноль (ZF=1 если всё нули):
    VPTEST  Y0, Y0

    // Если NE (ZF=0) — есть ненулевой бит:
    JNZ     nonzero

zero:
    MOVB    $0, AL         // false
    RET

nonzero:
    MOVB    $1, AL         // true
    RET
