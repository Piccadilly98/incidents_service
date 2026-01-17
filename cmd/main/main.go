package main

import (
	"log"

	"github.com/Piccadilly98/incidents_service/internal/server"
)

func main() {
	ch, err := server.ServerStart()
	if err != nil {
		log.Fatal(err)
	}
	err = <-ch
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server stop")
}
