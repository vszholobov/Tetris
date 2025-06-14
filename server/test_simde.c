#include <stdio.h>
#include "simde/simde/x86/avx2.h"
#include <stdint.h>
#include <stdbool.h>

bool intersects_simde_single(const uint16_t* a, const uint16_t* b) {
    simde__m256i va = simde_mm256_loadu_si256((const simde__m256i*)a);
    simde__m256i vb = simde_mm256_loadu_si256((const simde__m256i*)b);
    simde__m256i vand = simde_mm256_and_si256(va, vb);
    return simde_mm256_testz_si256(vand, vand) == 0;
}

int intersects_simde_many(const uint16_t* a, const uint16_t* b, size_t n) {
    int count = 0;
    for (size_t i = 0; i < n; i++) {
        const uint16_t* pa = a + i * 16;
        const uint16_t* pb = b + i * 16;

        simde__m256i va = simde_mm256_loadu_si256((const simde__m256i*)pa);
        simde__m256i vb = simde_mm256_loadu_si256((const simde__m256i*)pb);
        simde__m256i vand = simde_mm256_and_si256(va, vb);

        // simde_mm256_testz_si256 возвращает 1 если и-или 0 (== пусто),
        // нам нужно противоположное — непустое пересечение
        if (simde_mm256_testz_si256(vand, vand) == 0) {
            count++;
        }
    }
    return count;
}

// int main() {
//     uint16_t a[16] = {0};
//     uint16_t b[16] = {0};
//     a[5] = 0x0004;
//     b[5] = 0x0004;
//     if (intersects_simde(a, b)) {
//         printf("Пересечение есть\n");
//     } else {
//         printf("Пересечения нет\n");
//     }
//     return 0;
// }
