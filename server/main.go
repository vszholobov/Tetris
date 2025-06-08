package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"tetrisServer/server"

	"github.com/gorilla/mux"
)

func main() {
	f := openLogFile()
	defer f.Close()
	log.SetOutput(f)
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
