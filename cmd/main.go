package main

import (
	"Nietzsche/internal/server"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	var signals = make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	var wg sync.WaitGroup
	wg.Add(2)

	s := server.NewServer()
	//go server.RunIoMultiplexingServer(&wg)
	go s.StartSingleListener(&wg)
	//go s.StartMultiListeners(&wg)
	go server.WaitForSignal(&wg, signals)
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	wg.Wait()
}
