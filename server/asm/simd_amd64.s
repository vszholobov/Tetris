//go:build amd64
// +build amd64

#include "textflag.h"

// func IntersectsAVX(a, b *[16]uint16) bool
TEXT ·IntersectsAVX(SB), NOSPLIT, $0-24
    VMOVDQU (DI), Y0
    VMOVDQU (SI), Y1
    VPAND  Y1, Y0, Y0
    VPTEST Y0, Y0
    JNZ    nonzero
zero:
    MOVB   $0, AL
    RET
nonzero:
    MOVB   $1, AL
    RET
