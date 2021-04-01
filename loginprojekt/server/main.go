package main

import (
	"log"
	"os"
	"os/signal"
)

func main() {
	go func() {
		if err := StartFileServer(); err != nil {
			log.Fatal("file server:", err)
		}
	}()

	stopChan := make(chan os.Signal, 2)
	signal.Notify(stopChan, os.Interrupt)

	closer := make(chan struct{})
	errChan := StartAPIServer(closer)
	<-stopChan
	close(closer)
	if err := <-errChan; err != nil {
		log.Print("api server:", err)
	}
}
