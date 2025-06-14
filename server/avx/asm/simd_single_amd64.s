//go:build amd64
// +build amd64

#include "textflag.h"

// func IntersectsAVX(a, b *[16]uint16) bool
TEXT ·IntersectsAVXSingle(SB), NOSPLIT, $0-32
    MOVQ    a+0(FP), DI
    MOVQ    b+8(FP), SI
    VMOVDQU (DI), Y0
    VMOVDQU (SI), Y1

    VPAND   Y1, Y0, Y0
    VPTEST  Y0, Y0
    SETNE   AL
    MOVB    AL, ret+16(FP)

    VZEROUPPER
    RET
