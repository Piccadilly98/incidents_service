package main

import (
	"log"

	"github.com/Piccadilly98/incidents_service/internal/server"
)

func main() {
	err := server.ServerStart()
	if err != nil {
		log.Fatal(err)
	}
}
