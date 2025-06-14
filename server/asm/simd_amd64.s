//go:build amd64
// +build amd64

#include "textflag.h"

// func IntersectsAVXMany(a, b **[16]uint16, n int) int
TEXT ·IntersectsAVXMany(SB), NOSPLIT, $0-32
    MOVQ a+0(FP), R8       // R8 = a
    MOVQ b+8(FP), R9       // R9 = b
    MOVQ n+16(FP), CX      // CX = n
    XORL AX, AX            // AX = result (int)

    XORQ R10, R10          // R10 = index i = 0

loop:
    CMPQ R10, CX
    JAE done

    MOVQ (R8)(R10*8), R11      // R11 = a[i]
    MOVQ (R9)(R10*8), R12      // R12 = b[i]

    VMOVDQU (R11), Y0
    VMOVDQU (R12), Y1
    VPAND Y1, Y0, Y0
    VPTEST Y0, Y0
    JZ skip

    // increment result
    INCQ AX

skip:
    INCQ R10
    JMP loop

done:
    VZEROUPPER
    MOVL AX, ret+24(FP)
    RET





////go:build amd64
//// +build amd64
//
//#include "textflag.h"
//
//// func IntersectsAVX(a, b *[16]uint16) bool
//TEXT ·IntersectsAVX(SB), NOSPLIT, $0-32
//    // Считаем a → DI, b → SI
//    MOVQ    a+0(FP), DI
//    MOVQ    b+8(FP), SI
//
//    // Загружаем 256 бит
//    VMOVDQU (DI), Y0
//    VMOVDQU (SI), Y1
//
//    // Побитовое AND: Y0 = Y0 & Y1
//    VPAND   Y1, Y0, Y0
//
//    // Проверка на ноль
//    VPTEST  Y0, Y0
//
//    // AL := (Y0 != 0)
//    SETNE   R8B
//    // Записываем результат в ret
//    MOVB    R8B, ret+16(FP)
//
//    // Очистка AVX
//    VZEROUPPER
//
//    RET
