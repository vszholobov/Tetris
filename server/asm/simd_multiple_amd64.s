//go:build amd64
// +build amd64

#include "textflag.h"

// func IntersectsAVXMany(a, b **[16]uint16, n int) int
TEXT ·IntersectsAVXMultiple(SB), NOSPLIT, $0-32
    MOVQ a+0(FP), R8       // R8 = a
    MOVQ b+8(FP), R9       // R9 = b
    MOVQ n+16(FP), CX      // CX = n
    XORQ RAX, RAX          // RAX = 0 (64-битный ноль)

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

    INCQ RAX                  // увеличиваем результат

skip:
    INCQ R10
    JMP loop

done:
    VZEROUPPER
    MOVQ RAX, ret+24(FP)      // записываем 64-бит результат
    RET
