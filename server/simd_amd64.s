// +build amd64

#include "textflag.h"

// func IntersectsAVX(a, b *[16]uint16) bool
TEXT ·IntersectsAVX(SB), NOSPLIT, $0-24
    // a в DI, b в SI (первый и второй аргументы - указатели)
    
    // Загружаем 256 бит из a и b в ymm0 и ymm1
    VMOVDQU (DI), Y0
    VMOVDQU (SI), Y1

    // Выполняем AND: ymm0 = ymm0 & ymm1
    VPAND Y1, Y0, Y0

    // Тестируем результат (проверяем, что в ymm0 хоть один бит установлен)
    // vptest проверяет равенство нулю для ymm
    VPTEST Y0, Y0

    // CF=1 и ZF=1 означает, что результат ноль (AND == 0)
    // CF=0 или ZF=0 - есть биты установлены (AND != 0)

    // Возвращаем bool в AL:
    // если не ноль -> 1
    // если ноль -> 0

    // jnz - прыжок, если результат не ноль
    JNZ nonzero

zero:
    MOVB $0, AL
    RET

nonzero:
    MOVB $1, AL
    RET
