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

    // Сохраним результат AND в 256‑битный буфер
    SUBQ    $32, SP
    VMOVDQU Y0, (SP)

    // Загрузим и объединим 4×64‑бита без обнуления
    MOVQ    (SP), AX       // AX = первыe 8 байт
    MOVQ    8(SP), CX      // CX = следующие 8 байт
    ORQ     CX, AX         // AX |= CX
    MOVQ    16(SP), CX     // ...
    ORQ     CX, AX
    MOVQ    24(SP), CX
    ORQ     CX, AX

    ADDQ    $32, SP        // восстановим стек

    TESTQ   AX, AX         // проверяем, стала ли ненулевой
    JNZ     nonzero

zero:
    MOVB    $0, AL
    RET

nonzero:
    MOVB    $1, AL
    RET
