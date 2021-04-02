package main

import (
	"log"
	"os"
	"os/signal"
)

func main() {
	go StartFileServer()

	killer := make(chan os.Signal, 3)
	signal.Notify(killer, os.Interrupt)
	apistopper := make(chan struct{}, 5)
	finish := StartAPIServer(apistopper)
	<-killer
	log.Print("Stopping server...")
	close(apistopper)
	<-finish
}
