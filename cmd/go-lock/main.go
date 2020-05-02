package main

import (
	"go-lock/internal/config"
	"log"
)

var configuration = config.NewManager()

func main() {
	log.Println(configuration.Redlock.Clients)
}
