//go:build amd64
// +build amd64

#include "textflag.h"

// func IntersectsAVX(a, b *[16]uint16) bool
TEXT ·IntersectsAVX(SB), NOSPLIT, $0-32
    MOVQ    a+0(FP), DI
    MOVQ    b+8(FP), SI

    VMOVDQU (DI), Y0
    VMOVDQU (SI), Y1
    VPAND   Y1, Y0, Y0

    // Сохраним результат AND в 256-битный буфер на стеке
    SUBQ    $32, SP
    VMOVDQU Y0, (SP)

    // Проверим все 32 байта на ноль через XOR и TEST
    MOVQ    (SP), AX
    XORQ    AX, AX
    MOVQ    8(SP), CX
    ORQ     CX, AX
    MOVQ    16(SP), DX
    ORQ     DX, AX
    MOVQ    24(SP), BX
    ORQ     BX, AX

    ADDQ    $32, SP

    TESTQ   AX, AX
    JNZ     nonzero

zero:
    MOVB    $0, AL
    RET

nonzero:
    MOVB    $1, AL
    RET
