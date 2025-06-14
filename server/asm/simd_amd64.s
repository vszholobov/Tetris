//go:build amd64
// +build amd64

#include "textflag.h"

// func IntersectsAVX(a, b *[16]uint16) bool
TEXT ·IntersectsAVX(SB), NOSPLIT, $0-40
    // Загрузим из стека аргументы:
    MOVQ    a+0(FP), RDI   // первый аргумент -> RDI
    MOVQ    b+8(FP), RSI   // второй аргумент -> RSI

    // Теперь RDI и RSI указывают на [16]uint16
    VMOVDQU (RDI), Y0      // load 256 бит from a
    VMOVDQU (RSI), Y1      // load 256 бит from b

    VPAND   Y1, Y0, Y0     // v0 = v0 & v1
    VPTEST  Y0, Y0         // test v0 == 0 ?

    JNZ     nonzero

zero:
    MOVB    $0, AL         // result = false
    RET

nonzero:
    MOVB    $1, AL         // result = true
    RET
