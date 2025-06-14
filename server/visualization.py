import matplotlib.pyplot as plt

# Размеры данных с 10^8 включительно
data_sizes = [10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000]

# Время ns/op из твоих тестов
results = {
    'BigInt':        [448.0, 4844, 48752, 627679, 5766907, 56923867, 927161420, 8794907676],
    'Uint16':        [20.26, 214.3, 2025, 22326, 217134, 4335405, 44193217, 443421414],
    'Uint32':        [20.66, 217.8, 2107, 19973, 218827, 4312343, 44228432, 416975521],
    'Uint64':        [21.58, 212.7, 2044, 19615, 205981, 4295708, 43432388, 418599683],
    'AsmAvxSingle':  [34.79, 393.2, 4014, 40031, 406433, 5305791, 52890096, 519873832],
    'AsmAvxMany':    [13.75, 113.0, 1134, 11026, 169282, 3362617, 36276905, 337306960],
    'CSimdeSingle':  [679.5, 7411, 75448, 734905, 7905077, 72989308, 757140152, 7689624626],
    'CSimdeMany':    [83.22, 162.5, 1343, 14400, 242735, 4822645, 46536181, 453287930],
}

plt.figure(figsize=(14, 9))

for name, times in results.items():
    plt.plot(data_sizes, times, marker='o', label=name)

plt.xscale('log')
plt.yscale('log')
plt.xlabel('Data Size (log scale)')
plt.ylabel('Time per operation (ns/op, log scale)')
plt.title('Benchmark Time per Operation vs Data Size')
plt.grid(True, which='both', linestyle='--', linewidth=0.5)
plt.legend()
plt.tight_layout()
plt.show()
