import re
import matplotlib.pyplot as plt

# Вставь сюда свой лог вывода
raw_text = """
vszholobov@simde-experiment:~/Tetris/server$ go test -bench=. -benchmem -dataSize=10
goos: linux
goarch: amd64
pkg: tetrisServer
cpu: Intel Xeon Processor (Icelake)
BenchmarkIntersectsBigInt-8         	 2768467	       439.7 ns/op	     576 B/op	       9 allocs/op
BenchmarkIntersectsUint16-8         	57499832	        20.28 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint32-8         	54535855	        21.23 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint64-8         	55613364	        21.51 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxSingle-8   	32902616	        34.19 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxMany-8     	96009961	        12.46 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeSingle-8             	 1780060	       744.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeMany-8               	12603126	        81.66 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	tetrisServer	10.775s
vszholobov@simde-experiment:~/Tetris/server$ go test -bench=. -benchmem -dataSize=100
goos: linux
goarch: amd64
pkg: tetrisServer
cpu: Intel Xeon Processor (Icelake)
BenchmarkIntersectsBigInt-8         	  238242	      4722 ns/op	    6336 B/op	      99 allocs/op
BenchmarkIntersectsUint16-8         	 5904786	       203.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint32-8         	 5561772	       215.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint64-8         	 5277134	       229.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxSingle-8   	 2747512	       443.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxMany-8     	10588017	       112.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeSingle-8             	  161794	      7423 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeMany-8               	 7243707	       173.6 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	tetrisServer	11.139s
vszholobov@simde-experiment:~/Tetris/server$ go test -bench=. -benchmem -dataSize=1000
goos: linux
goarch: amd64
pkg: tetrisServer
cpu: Intel Xeon Processor (Icelake)
BenchmarkIntersectsBigInt-8         	   24739	     48326 ns/op	   63936 B/op	     999 allocs/op
BenchmarkIntersectsUint16-8         	  602926	      1890 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint32-8         	  554212	      2086 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint64-8         	  558814	      2139 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxSingle-8   	  303235	      3976 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxMany-8     	  994664	      1124 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeSingle-8             	   14581	     77281 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeMany-8               	  925566	      1336 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	tetrisServer	10.857s
vszholobov@simde-experiment:~/Tetris/server$ go test -bench=. -benchmem -dataSize=10000
goos: linux
goarch: amd64
pkg: tetrisServer
cpu: Intel Xeon Processor (Icelake)
BenchmarkIntersectsBigInt-8         	    1942	    616803 ns/op	  639941 B/op	    9999 allocs/op
BenchmarkIntersectsUint16-8         	   58641	     19949 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint32-8         	   58291	     21438 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint64-8         	   54039	     22473 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxSingle-8   	   29344	     41248 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxMany-8     	  105860	     11467 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeSingle-8             	    1520	    813030 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeMany-8               	   94180	     13253 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	tetrisServer	12.267s
vszholobov@simde-experiment:~/Tetris/server$ go test -bench=. -benchmem -dataSize=100000
goos: linux
goarch: amd64
pkg: tetrisServer
cpu: Intel Xeon Processor (Icelake)
BenchmarkIntersectsBigInt-8         	     202	   5899917 ns/op	 6399943 B/op	   99999 allocs/op
BenchmarkIntersectsUint16-8         	    5140	    220792 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint32-8         	    5797	    208943 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint64-8         	    5236	    220428 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxSingle-8   	    2786	    412093 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxMany-8     	    7182	    164203 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeSingle-8             	     158	   7702849 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeMany-8               	    5710	    240176 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	tetrisServer	11.891s
vszholobov@simde-experiment:~/Tetris/server$ go test -bench=. -benchmem -dataSize=1000000
goos: linux
goarch: amd64
pkg: tetrisServer
cpu: Intel Xeon Processor (Icelake)
BenchmarkIntersectsBigInt-8         	      19	  58357315 ns/op	63999956 B/op	  999999 allocs/op
BenchmarkIntersectsUint16-8         	     261	   4536421 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint32-8         	     258	   4446418 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint64-8         	     262	   4496835 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxSingle-8   	     220	   5428542 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxMany-8     	     356	   3371766 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeSingle-8             	      14	  78447687 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeMany-8               	     262	   4486907 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	tetrisServer	20.161s
vszholobov@simde-experiment:~/Tetris/server$ go test -bench=. -benchmem -dataSize=10000000
goos: linux
goarch: amd64
pkg: tetrisServer
cpu: Intel Xeon Processor (Icelake)
BenchmarkIntersectsBigInt-8         	       2	 661690009 ns/op	639999936 B/op	 9999999 allocs/op
BenchmarkIntersectsUint16-8         	      26	  43901718 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint32-8         	      26	  44342148 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint64-8         	      26	  45046367 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxSingle-8   	      22	  52688086 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxMany-8     	      33	  36001602 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeSingle-8             	       2	 781361511 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeMany-8               	      25	  47962587 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	tetrisServer	49.792s
vszholobov@simde-experiment:~/Tetris/server$ go test -bench=. -benchmem -dataSize=100000000
goos: linux
goarch: amd64
pkg: tetrisServer
cpu: Intel Xeon Processor (Icelake)
BenchmarkIntersectsBigInt-8         	       1	6723389505 ns/op	6399999936 B/op	99999999 allocs/op
BenchmarkIntersectsUint16-8         	       3	 430843099 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint32-8         	       3	 428122313 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsUint64-8         	       3	 435192510 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxSingle-8   	       2	 515565099 ns/op	       0 B/op	       0 allocs/op
BenchmarkIntersectsAsmAvxMany-8     	       3	 337486197 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeSingle-8             	       1	7881079159 ns/op	       0 B/op	       0 allocs/op
BenchmarkCSimdeMany-8               	       3	 464823593 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	tetrisServer	447.618s
"""

# Извлекаем dataSize блоки
blocks = raw_text.split('go test -bench=. -benchmem -dataSize=')
results = {}

for block in blocks[1:]:  # пропускаем первый элемент до первого dataSize
    lines = block.strip().splitlines()
    try:
        data_size = int(lines[0])
    except ValueError:
        continue

    for line in lines[1:]:
        if line.startswith("Benchmark"):
            parts = re.split(r'\s+', line)
            name = parts[0].replace("Benchmark", "")
            ns_per_op = float(parts[2].replace("ns/op", ""))
            if name not in results:
                results[name] = []
            results[name].append((data_size, ns_per_op))

# Строим график
plt.figure(figsize=(12, 8))
for name, values in sorted(results.items()):
    values.sort()
    sizes = [v[0] for v in values]
    times = [v[1] for v in values]
    plt.plot(sizes, times, marker='o', label=name)

plt.xlabel("dataSize")
plt.ylabel("ns/op")
plt.xscale("log")
plt.yscale("log")
plt.title("Benchmark Performance by dataSize (ns/op)")
plt.legend()
plt.grid(True, which="both", linestyle="--", alpha=0.5)
plt.tight_layout()
plt.show()
