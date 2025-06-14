package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"tetrisServer/asm"
	"tetrisServer/server"
	"unsafe"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

func main() {
	// Пример с данными из [][]uint16
	e := make([]uint16, 16)
	f := make([]uint16, 16)
	e[3] = 0xFFFF
	f[3] = 0x0001

	// Преобразуем в *[16]uint16
	ptr1 := (*[16]uint16)(unsafe.Pointer(&e[0]))
	ptr2 := (*[16]uint16)(unsafe.Pointer(&f[0]))

	result := asm.IntersectsAVXSingle(ptr1, ptr2)
	fmt.Println("intersects (should be true):", result)

	g := [16]uint16{}
	h := [16]uint16{}

	g[5] = 0x1234
	h[5] = 0x1234

	result = asm.IntersectsAVXSingle(&g, &h)
	fmt.Println("intersects (should be true):", result)

	h[5] = 0x0
	result = asm.IntersectsAVXSingle(&g, &h)
	fmt.Println("intersects (should be false):", result)

	N := 100
	c := make([]*[16]uint16, N)
	d := make([]*[16]uint16, N)

	for i := 0; i < N; i++ {
		arr1 := [16]uint16{}
		arr2 := [16]uint16{}

		// Пример заполнения
		if i%2 == 0 {
			arr1[3] = 0xFF
			arr2[3] = 0xFF
		}

		c[i] = &arr1
		d[i] = &arr2
	}

	count := asm.IntersectsAVXMultiple(&c[0], &d[0], N)
	fmt.Println("Intersections:", count)

	a := make([]uint16, 16)
	b := make([]uint16, 16)
	a[5] = 0x0004
	b[5] = 0x0004

	if intersectsSimde(a, b) {
		fmt.Println("Пересечение есть")
	} else {
		fmt.Println("Пересечения нет")
	}

	logFile := openLogFile()
	defer logFile.Close()
	log.SetOutput(logFile)
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	flag.Parse()
	router := mux.NewRouter()
	router.HandleFunc("/session", server.GetSessionsList)
	router.HandleFunc("/session/create", server.CreateSession)
	router.HandleFunc("/session/connect/{sessionId}", server.ConnectToSession)
	//router.HandleFunc("/session/ping/{sessionId}", server.MeasurePing)
	router.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*server.Addr, router))
}

func openLogFile() *os.File {
	err := os.MkdirAll("./tetris-logs", 0777)
	logFile := "./tetris-logs/tetris-log.txt"
	log.SetReportCaller(true)
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println("Failed to create logfile" + logFile)
		panic(err)
	}
	return f
}
