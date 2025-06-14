#include <stdio.h>
#include "simde/simde/x86/avx2.h"
#include <stdint.h>
#include <stdbool.h>

bool intersects_simde(const uint16_t* a, const uint16_t* b) {
    simde__m256i va = simde_mm256_loadu_si256((const simde__m256i*)a);
    simde__m256i vb = simde_mm256_loadu_si256((const simde__m256i*)b);
    simde__m256i vand = simde_mm256_and_si256(va, vb);
    return simde_mm256_testz_si256(vand, vand) == 0;
}

int main() {
    uint16_t a[16] = {0};
    uint16_t b[16] = {0};
    a[5] = 0x0004;
    b[5] = 0x0004;
    if (intersects_simde(a, b)) {
        printf("Пересечение есть\n");
    } else {
        printf("Пересечения нет\n");
    }
    return 0;
}