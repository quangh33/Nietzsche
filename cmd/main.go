package main

import (
	"Nietzsche/internal/server"
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
	//go s.Start(&wg)
	go s.StartMultiListeners(&wg)
	go server.WaitForSignal(&wg, signals)
	wg.Wait()
}
