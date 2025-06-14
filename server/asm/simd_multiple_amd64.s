//go:build amd64
// +build amd64

#include "textflag.h"

// func IntersectsAVXMultiple(a, b **[16]uint16, n int) int
TEXT ·IntersectsAVXMultiple(SB), NOSPLIT, $0-32
    MOVQ    a+0(FP), R8       // R8 = a
    MOVQ    b+8(FP), R9       // R9 = b
    MOVQ    n+16(FP), CX      // CX = n
    XORQ    AX, AX            // RAX = count = 0
    XORQ    R10, R10          // R10 = i = 0

    MOVQ    CX, R11
    SHRQ    $2, R11           // R11 = n/4 (кол-во полных блоков по 4)
    ANDQ    $3, CX            // CX = остаток n % 4

    TESTQ   R11, R11
    JE      loop_tail         // Если нет блоков по 4, сразу к хвосту

loop_main:
    // Первый элемент
    MOVQ    (R8)(R10*8), R12
    MOVQ    (R9)(R10*8), R13
    VMOVDQU (R12), Y0
    VMOVDQU (R13), Y1
    VPAND   Y1, Y0, Y0
    VPTEST  Y0, Y0
    JZ      skip0
    INCQ    AX
skip0:
    INCQ    R10

    // Второй элемент
    MOVQ    (R8)(R10*8), R12
    MOVQ    (R9)(R10*8), R13
    VMOVDQU (R12), Y0
    VMOVDQU (R13), Y1
    VPAND   Y1, Y0, Y0
    VPTEST  Y0, Y0
    JZ      skip1
    INCQ    AX
skip1:
    INCQ    R10

    // Третий элемент
    MOVQ    (R8)(R10*8), R12
    MOVQ    (R9)(R10*8), R13
    VMOVDQU (R12), Y0
    VMOVDQU (R13), Y1
    VPAND   Y1, Y0, Y0
    VPTEST  Y0, Y0
    JZ      skip2
    INCQ    AX
skip2:
    INCQ    R10

    // Четвертый элемент
    MOVQ    (R8)(R10*8), R12
    MOVQ    (R9)(R10*8), R13
    VMOVDQU (R12), Y0
    VMOVDQU (R13), Y1
    VPAND   Y1, Y0, Y0
    VPTEST  Y0, Y0
    JZ      skip3
    INCQ    AX
skip3:
    INCQ    R10

    DECQ    R11
    JNZ     loop_main

loop_tail:
    TESTQ   CX, CX
    JE      done

loop_tail_loop:
    MOVQ    (R8)(R10*8), R12
    MOVQ    (R9)(R10*8), R13
    VMOVDQU (R12), Y0
    VMOVDQU (R13), Y1
    VPAND   Y1, Y0, Y0
    VPTEST  Y0, Y0
    JZ      skip_tail
    INCQ    AX
skip_tail:
    INCQ    R10
    DECQ    CX
    JNZ     loop_tail_loop

done:
    VZEROUPPER
    MOVQ    AX, ret+24(FP)
    RET
